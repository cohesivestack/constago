package constago

import (
	"fmt"
	"go/ast"
	"go/parser"
	goScanner "go/scanner"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// modelBuilder builds a Model by scanning Go source according to Config
type modelBuilder struct {
	config *Config
	model  *Model
}

// BuildModel builds and returns a populated Model for the given config
func (b *modelBuilder) Build() (*Model, error) {

	err := b.scanFiles()
	if err != nil {
		return nil, err
	}

	return b.model, nil
}

func NewModelBuilder(config *Config) *modelBuilder {
	return &modelBuilder{config: config, model: NewModel(config)}
}

// findFiles resolves include/exclude patterns into a set of Go files
func (b *modelBuilder) findFiles() ([]string, error) {
	config := b.config

	// Exclusions
	mustExclude := make(map[string]bool)
	for _, exclude := range config.Input.Exclude {
		paths, err := b.expandPattern(exclude)
		if err != nil {
			return nil, fmt.Errorf("failed to expand pattern %s: %w", exclude, err)
		}
		for _, p := range paths {
			mustExclude[p] = true
		}
	}

	// Inclusions
	includeSet := make(map[string]bool)
	for _, include := range config.Input.Include {
		paths, err := b.expandPattern(include)
		if err != nil {
			return nil, fmt.Errorf("failed to expand pattern %s: %w", include, err)
		}
		for _, p := range paths {
			if !mustExclude[p] {
				includeSet[p] = true
			}
		}
	}

	files := make([]string, 0, len(includeSet))
	for p := range includeSet {
		files = append(files, p)
	}
	return files, nil
}

func (b *modelBuilder) scanFiles() error {

	files, err := b.findFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := b.scanFile(file); err != nil {
			return err
		}
	}

	return nil
}

// expandPattern expands a single include/exclude pattern
func (b *modelBuilder) expandPattern(pattern string) ([]string, error) {
	config := b.config

	if strings.HasPrefix(pattern, "package:") {
		pkg := strings.TrimPrefix(pattern, "package:")
		return b.findPackageFiles(pkg)
	}

	matches, err := doublestar.Glob(os.DirFS(config.Input.Dir), pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid glob pattern %s: %w", pattern, err)
	}
	var goFiles []string
	for _, m := range matches {
		if strings.HasSuffix(m, ".go") {
			goFiles = append(goFiles, filepath.Join(config.Input.Dir, m))
		}
	}
	return goFiles, nil
}

// findPackageFiles finds .go files that belong to a given package name
func (b *modelBuilder) findPackageFiles(packageName string) ([]string, error) {
	config := b.config

	var files []string
	err := filepath.Walk(config.Input.Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.PackageClauseOnly)
		if err != nil {
			// best effort; skip invalid files
			return nil
		}
		if node != nil && node.Name.Name == packageName {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func (s *modelBuilder) mustIncludeStruct(genDecl *ast.GenDecl, typeSpec *ast.TypeSpec, fset *token.FileSet, filePath string) bool {

	includeDirective, excludeDirective := s.structDirectives(genDecl, typeSpec)

	if includeDirective && excludeDirective {
		s.model.AddError(filePath, fset.Position(typeSpec.Pos()).Line, "struct has both include and exclude directives")
		return false
	}

	if excludeDirective || (s.config.Input.Struct.isExplicit() && !includeDirective) {
		return false
	}

	// Skip unexported structs if configured, unless explicitly included via directive
	if !includeDirective &&
		!s.config.Input.Struct.isIncludeUnexported() &&
		!ast.IsExported(typeSpec.Name.Name) {
		return false
	}

	return true
}

// structDirectives inspects comments attached to a type declaration/spec
// and returns whether include/exclude directives are present.
func (s *modelBuilder) structDirectives(genDecl *ast.GenDecl, typeSpec *ast.TypeSpec) (bool, bool) {
	hasInclude := false
	hasExclude := false

	checkCommentGroup := func(cg *ast.CommentGroup) {
		if cg == nil {
			return
		}
		for _, c := range cg.List {
			txt := strings.TrimSpace(c.Text)
			// Support both //constago:include and // constago:exclude (with optional space)
			if strings.Contains(txt, "constago:include") {
				hasInclude = true
			}
			if strings.Contains(txt, "constago:exclude") {
				hasExclude = true
			}
		}
	}

	// Comments may be on the declaration or on the specific spec/name doc
	checkCommentGroup(genDecl.Doc)
	// If the TypeSpec has its own doc/comments (rare but possible), check those
	checkCommentGroup(typeSpec.Doc)

	return hasInclude, hasExclude
}

func (b *modelBuilder) scanFile(filePath string) error {

	b.model.FilesScanned++

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		// attach parsing error with line when available
		line := 0
		if se, ok := err.(goScanner.Error); ok {
			line = se.Pos.Line
		} else if sel, ok := err.(goScanner.ErrorList); ok && len(sel) > 0 {
			line = sel[0].Pos.Line
		}
		b.model.AddError(filePath, line, fmt.Sprintf("failed to parse file: %v", err))
		return nil
	}

	packagePath := b.extractPackagePath(filePath)
	packageName := node.Name.Name
	// Build import index for resolving selector types to full import info
	importIndex, modulePath := b.buildImportIndex(node, filePath)
	moduleDir, _ := locateGoModule(filePath)

	// Aggregations are per-struct, so they will be initialized inside the struct loop
	ast.Inspect(node, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			return true
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			if !b.mustIncludeStruct(genDecl, typeSpec, fset, filePath) {
				continue
			}

			structModel := &StructModel{
				Name:       typeSpec.Name.Name,
				File:       filePath,
				LineNumber: fset.Position(typeSpec.Pos()).Line,
				Constants:  []*ConstantOutput{},
				Structs:    []*StructOutput{},
				Getters:    []*GetterOutput{},
			}

			// Per-field+element constants cache
			constantsByFieldAndElement := map[string]map[string]*ConstantOutput{}
			// Per-field none cache
			noneByFieldAndElement := map[string]map[string]*NoneOutput{}
			// Per-element struct outputs cache (element name -> struct output)
			structByElement := map[string]*StructOutput{}
			// Per-field of struct-field outputs cache
			structFieldByFieldAndElement := map[string]map[string]*FieldOutput{}

			// Process fields
			for _, field := range structType.Fields.List {
				// Skip anonymous fields
				if len(field.Names) == 0 {
					continue
				}
				if !b.mustIncludeField(field) {
					continue
				}

				var tagText string
				if field.Tag != nil {
					tagText = strings.Trim(field.Tag.Value, "`")
				}

				for _, ident := range field.Names {
					fieldName := ident.Name

					// Build per-element artifacts
					for i := range b.config.Elements {
						el := &b.config.Elements[i]
						value := b.computeElementValue(fieldName, tagText, el)
						if value == "" {
							continue
						}

						switch el.Output.Mode {
						case OutputModeConstant:
							// Top-level constant name
							constName := b.buildName(el.Output.Format.Prefix, structModel.Name, fieldName, el.Output.Format.Suffix, el.Output.Format.Struct)
							c := &ConstantOutput{Name: constName, Value: value}
							structModel.Constants = append(structModel.Constants, c)
							if _, ok := constantsByFieldAndElement[fieldName]; !ok {
								constantsByFieldAndElement[fieldName] = map[string]*ConstantOutput{}
							}
							constantsByFieldAndElement[fieldName][el.Name] = c

						case OutputModeStruct:
							// Ensure struct output exists for this element
							so, ok := structByElement[el.Name]
							if !ok {
								structName := b.buildName(el.Output.Format.Prefix, structModel.Name, "", el.Output.Format.Suffix, el.Output.Format.Struct)
								so = &StructOutput{Name: structName, Package: packageName}
								structByElement[el.Name] = so
								structModel.Structs = append(structModel.Structs, so)
							}
							// Field name inside struct uses holder format
							fieldConstName := b.buildName("", fieldName, "", "", el.Output.Format.Holder)
							fieldOutput := &FieldOutput{StructName: so.Name, Name: fieldConstName, Value: value}
							so.Fields = append(so.Fields, fieldOutput)

							if _, ok := structFieldByFieldAndElement[fieldName]; !ok {
								structFieldByFieldAndElement[fieldName] = map[string]*FieldOutput{}
							}
							structFieldByFieldAndElement[fieldName][el.Name] = fieldOutput
						case OutputModeNone:
							if _, ok := noneByFieldAndElement[fieldName]; !ok {
								noneByFieldAndElement[fieldName] = map[string]*NoneOutput{}
							}
							noneByFieldAndElement[fieldName][el.Name] = &NoneOutput{Name: fieldName, Value: value}
						}
					}

					// Build getters for this field
					for gi := range b.config.Getters {
						g := &b.config.Getters[gi]
						getterName := b.buildName(g.Output.Prefix, fieldName, g.Output.Suffix, "", g.Output.Format)
						getter := &GetterOutput{Name: getterName}

						for _, ret := range g.Returns {
							// Handle special returns
							if strings.HasPrefix(ret, ":") {
								if ret == ":value" {
									// Create ValueOutput for field value return
									valueOutput := b.createValueOutput(field, fieldName, packageName, importIndex, modulePath, moduleDir)
									if valueOutput != nil {
										getter.Returns = append(getter.Returns, &ReturnOutput{Value: valueOutput})
									}
								}
								// Skip other special returns that imply external deps at this stage
								continue
							}
							// Prefer constant if produced
							if cm, ok := constantsByFieldAndElement[fieldName][ret]; ok {
								getter.Returns = append(getter.Returns, &ReturnOutput{Constant: cm})
							} else if no, ok := noneByFieldAndElement[fieldName][ret]; ok {
								// Since the name is not set in a Constant or a Field, then the name should be the
								// element name
								no.Name = ret
								getter.Returns = append(getter.Returns, &ReturnOutput{None: no})
							} else if so, ok := structFieldByFieldAndElement[fieldName][ret]; ok {
								getter.Returns = append(getter.Returns, &ReturnOutput{Field: so})
							}
						}

						// Add getter if all returns are satisfied
						if len(getter.Returns) == len(g.Returns) {
							structModel.Getters = append(structModel.Getters, getter)
						}
					}
				}
			}
			if len(structModel.Constants) > 0 || len(structModel.Structs) > 0 || len(structModel.Getters) > 0 {
				b.model.AddStruct(packagePath, packageName, structModel)
			}
		}
		return true
	})

	return nil
}

// extractPackagePath from file path
func (b *modelBuilder) extractPackagePath(filePath string) string {
	abs, err := filepath.Abs(filePath)
	if err != nil {
		abs = filePath
	}
	dir := filepath.Dir(abs)
	dir = filepath.ToSlash(dir)
	dir = strings.TrimPrefix(dir, "./")
	return dir
}

// mustIncludeField decides if a field should be processed according to config and tags
func (b *modelBuilder) mustIncludeField(field *ast.Field) bool {
	// Parse tags
	var tag reflect.StructTag
	if field.Tag != nil {
		tag = parseStructTags(strings.Trim(field.Tag.Value, "`"))
	}
	constagoTag, hasConstago := lookupTag(tag, "constago")

	if hasConstago && constagoTag == "exclude" {
		return false
	}
	if hasConstago && constagoTag == "include" {
		return true
	}
	if b.config.Input.Field.isExplicit() && !hasConstago {
		return false
	}
	if !b.config.Input.Field.isIncludeUnexported() && len(field.Names) > 0 && !ast.IsExported(field.Names[0].Name) {
		return false
	}
	return true
}

// computeElementValue computes element value considering mode, tag priority and transforms
func (b *modelBuilder) computeElementValue(fieldName string, tagText string, el *ConfigTag) string {
	// helper: pick first non-empty tag value by priority
	getFromTags := func() (string, bool) {
		if tagText == "" {
			return "", false
		}
		tags := parseStructTags(tagText)
		for _, key := range el.Input.TagPriority {
			if key == ":field" {
				// special pseudo-tag: refers to field name
				return fieldName, true
			}
			if v, ok := lookupTag(tags, key); ok {
				// Use the value up to first comma (e.g., json:"name,omitempty")
				parts := strings.SplitN(v, ",", 2)
				return parts[0], true
			}
		}
		return "", false
	}

	applyTransform := func(s string, cfg *ConfigTag) string {
		// If taken from tag and TagValues is false, return as-is
		return transformFieldValue(s, cfg.Output.Transform.ValueCase, cfg.Output.Transform.ValueSeparator)
	}

	switch el.Input.Mode {
	case InputModeTypeTag:
		if v, ok := getFromTags(); ok {
			if el.Output.Transform.TagValues != nil && *el.Output.Transform.TagValues {
				return applyTransform(v, el)
			}
			return v
		}
		return ""
	case InputModeTypeField:
		return applyTransform(fieldName, el)
	case InputModeTypeTagThenField:
		if v, ok := getFromTags(); ok {
			if el.Output.Transform.TagValues != nil && *el.Output.Transform.TagValues {
				return applyTransform(v, el)
			}
			return v
		}
		return applyTransform(fieldName, el)
	default:
		return ""
	}
}

// buildName builds a Go identifier from parts using a format
func (b *modelBuilder) buildName(prefix string, mid string, mid2 string, suffix string, fmtType ConstantFormatType) string {
	var parts []string
	if prefix != "" {
		parts = append(parts, prefix)
	}
	if mid != "" {
		parts = append(parts, mid)
	}
	if mid2 != "" {
		parts = append(parts, mid2)
	}
	if suffix != "" {
		parts = append(parts, suffix)
	}
	base := strings.Join(parts, " ")
	switch fmtType {
	case ConstantFormatCamel:
		return toCamelCase(base)
	case ConstantFormatPascal:
		return toPascalCase(base)
	case ConstantFormatSnake:
		return strings.ToLower(strings.Join(splitIntoWords(base), "_"))
	case ConstantFormatSnakeUpper:
		return strings.ToUpper(strings.Join(splitIntoWords(base), "_"))
	default:
		return toPascalCase(base)
	}
}

// transformFieldValue applies case and separator rules
func transformFieldValue(value string, caseType TransformCaseType, sep string) string {
	// If we need to apply separator with case change, normalize to space-separated words first
	if sep != "" && caseType != TransformCaseAsIs {
		value = strings.Join(splitIntoWords(value), " ")
	}
	switch caseType {
	case TransformCaseAsIs:
		// leave as is
	case TransformCaseCamel:
		value = toCamelCase(value)
	case TransformCasePascal:
		value = toPascalCase(value)
	case TransformCaseUpper:
		value = strings.ToUpper(value)
	case TransformCaseLower:
		value = strings.ToLower(value)
	}
	if sep != "" {
		value = strings.Join(splitIntoWords(value), sep)
	}
	return value
}

// Tag helpers
// parseStructTags converts a raw tag string to reflect.StructTag
func parseStructTags(tagString string) reflect.StructTag {
	return reflect.StructTag(tagString)
}

func lookupTag(tags reflect.StructTag, key string) (string, bool) {
	v := tags.Get(key)
	if strings.TrimSpace(v) == "" {
		return "", false
	}
	return v, true
}

// createValueOutput creates a ValueOutput from an AST field
func (b *modelBuilder) createValueOutput(field *ast.Field, fieldName string, packageName string, importIndex map[string]*TypePackageOutput, modulePath string, moduleDir string) *ValueOutput {
	if field.Type == nil {
		return nil
	}

	typeName, pkg := b.extractTypeInfo(field.Type, importIndex, modulePath)
	if typeName == "" {
		return nil
	}

	// Last-chance: if pkg has name but no path, try to resolve from importIndex
	// Look for any entry with the same name that has a path
	if pkg != nil && pkg.Path == "" && pkg.Name != "" {
		for _, imp := range importIndex {
			if imp != nil && imp.Name == pkg.Name && imp.Path != "" {
				pkg = &TypePackageOutput{Path: imp.Path, Name: imp.Name}
				break
			}
		}
	}
	// If still no path and typeName encodes selector (e.g., yaml.Node), map by the selector's package name
	if (pkg == nil || pkg.Path == "") && strings.Contains(typeName, ".") {
		parts := strings.SplitN(typeName, ".", 2)
		if len(parts) == 2 {
			pkgName := parts[0]
			// First try to find an entry with the name that has a path
			for _, imp := range importIndex {
				if imp != nil && imp.Name == pkgName && imp.Path != "" {
					pkg = &TypePackageOutput{Path: imp.Path, Name: imp.Name}
					break
				}
			}
			// If still not found, try to find by the identifier used in code (e.g., "yaml.v3")
			if pkg == nil || pkg.Path == "" {
				for key, imp := range importIndex {
					if imp != nil && strings.Contains(key, pkgName) && imp.Path != "" {
						// Check if this looks like the right import (contains the package name)
						if strings.Contains(imp.Path, pkgName) || importPathHasSegment(imp.Path, pkgName) {
							// Use the package name from the entry, but if it looks like an identifier (contains dots),
							// try to get the actual package name from go list or infer from path
							name := imp.Name
							if strings.Contains(name, ".") {
								// The name looks like an identifier (e.g., "yaml.v3"), try to get the real package name
								if realPkgName := getPackageNameFromGoList(imp.Path, moduleDir); realPkgName != "" {
									name = realPkgName
								} else {
									// Fallback: for patterns like gopkg.in/yaml.v3, the package name is usually the part before the dot
									// Extract from path if it matches common patterns
									if strings.HasPrefix(imp.Path, "gopkg.in/") {
										// For gopkg.in/x.y, package name is usually x
										parts := strings.Split(strings.TrimPrefix(imp.Path, "gopkg.in/"), ".")
										if len(parts) > 0 {
											name = parts[0]
										}
									}
								}
							}
							pkg = &TypePackageOutput{Path: imp.Path, Name: name}
							break
						}
					}
				}
			}
		}
	}

	valueOutput := &ValueOutput{
		FieldName: fieldName,
		TypeName:  typeName,
		TypePackage: func() *TypePackageOutput {
			if pkg != nil {
				// If package path is missing (external standard alias style), try to map using modulePath
				if pkg.Path == "" && modulePath != "" {
					// Keep Name as is (resolved real package name), leave Path empty since we cannot infer full path reliably
				}
				return pkg
			}
			// For unqualified/basic types, associate with current package
			return &TypePackageOutput{Path: "", Name: packageName}
		}(),
	}

	return valueOutput
}

// extractTypeInfo extracts type name and package info from an AST expression
func (b *modelBuilder) extractTypeInfo(expr ast.Expr, importIndex map[string]*TypePackageOutput, modulePath string) (typeName string, pkg *TypePackageOutput) {
	switch t := expr.(type) {
	case *ast.Ident:
		// Simple type like int, string, MyType
		return t.Name, nil
	case *ast.SelectorExpr:
		// Qualified type like pkg.Type
		if ident, ok := t.X.(*ast.Ident); ok {
			// ident.Name is the import alias or package identifier used
			if imp, ok := importIndex[ident.Name]; ok {
				// For external packages, preserve the full qualified type name
				// For local packages, return just the type name
				if imp.Path == "" {
					// Local package - return just the type name
					return t.Sel.Name, &TypePackageOutput{Path: imp.Path, Name: imp.Name}
				} else {
					// External package - return qualified name
					return ident.Name + "." + t.Sel.Name, &TypePackageOutput{Path: imp.Path, Name: imp.Name}
				}
			}
			// Secondary lookup: find any import whose real Name matches the ident
			// Prefer entries with non-empty paths
			var bestMatch *TypePackageOutput
			for _, imp := range importIndex {
				if imp != nil && imp.Name == ident.Name {
					if imp.Path != "" {
						bestMatch = imp
						break // Found a match with a path, use it
					} else if bestMatch == nil {
						// Keep this as fallback if no path match found yet
						bestMatch = imp
					}
				}
			}
			if bestMatch != nil {
				if bestMatch.Path == "" {
					return t.Sel.Name, &TypePackageOutput{Path: bestMatch.Path, Name: bestMatch.Name}
				} else {
					return ident.Name + "." + t.Sel.Name, &TypePackageOutput{Path: bestMatch.Path, Name: bestMatch.Name}
				}
			}
			// Tertiary lookup: find any import whose path contains the ident as a segment (handles paths like gopkg.in/yaml.v3)
			// Prefer entries with non-empty paths
			for _, imp := range importIndex {
				if imp != nil && importPathHasSegment(imp.Path, ident.Name) && imp.Path != "" {
					return ident.Name + "." + t.Sel.Name, &TypePackageOutput{Path: imp.Path, Name: imp.Name}
				}
			}
			// Fallback: unknown import; keep ident as package name but do not lose selector context
			return ident.Name + "." + t.Sel.Name, &TypePackageOutput{Path: "", Name: ident.Name}
		}
	case *ast.ArrayType:
		// Array type like []int, []pkg.Type
		elementType, elementPkg := b.extractTypeInfo(t.Elt, importIndex, modulePath)
		return "[]" + elementType, elementPkg
	case *ast.MapType:
		// Map type like map[string]int
		keyType, _ := b.extractTypeInfo(t.Key, importIndex, modulePath)
		valueType, valuePkg := b.extractTypeInfo(t.Value, importIndex, modulePath)
		return fmt.Sprintf("map[%s]%s", keyType, valueType), valuePkg
	case *ast.StarExpr:
		// Pointer type like *int, *pkg.Type
		elementType, elementPkg := b.extractTypeInfo(t.X, importIndex, modulePath)
		return "*" + elementType, elementPkg
	case *ast.ChanType:
		// Channel type like chan int, <-chan int, chan<- int
		elementType, elementPkg := b.extractTypeInfo(t.Value, importIndex, modulePath)
		dir := ""
		switch t.Dir {
		case ast.SEND:
			dir = "chan<- "
		case ast.RECV:
			dir = "<-chan "
		default:
			dir = "chan "
		}
		return dir + elementType, elementPkg
	case *ast.FuncType:
		// Function type - simplified representation
		return "func", nil
	case *ast.InterfaceType:
		// Interface type
		return "interface{}", nil
	case *ast.StructType:
		// Struct type
		return "struct{}", nil
	case *ast.IndexExpr:
		// Generic type with single type parameter like Generic[string], Generic[yaml.Node]
		baseType, basePkg := b.extractTypeInfo(t.X, importIndex, modulePath)
		indexType, _ := b.extractTypeInfo(t.Index, importIndex, modulePath)
		return fmt.Sprintf("%s[%s]", baseType, indexType), basePkg
	case *ast.IndexListExpr:
		// Generic type with multiple type parameters like Generic[string, int]
		baseType, basePkg := b.extractTypeInfo(t.X, importIndex, modulePath)
		indexTypes := make([]string, len(t.Indices))
		for i, index := range t.Indices {
			indexType, _ := b.extractTypeInfo(index, importIndex, modulePath)
			indexTypes[i] = indexType
		}
		return fmt.Sprintf("%s[%s]", baseType, strings.Join(indexTypes, ", ")), basePkg
	}
	return "", nil
}

// buildImportIndex indexes file imports by the identifier used in code (alias or default name)
func (b *modelBuilder) buildImportIndex(node *ast.File, currentFilePath string) (map[string]*TypePackageOutput, string) {
	idx := make(map[string]*TypePackageOutput)
	// Try to locate module directory and module path (from go.mod)
	moduleDir, modulePath := locateGoModule(currentFilePath)
	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, "\"")

		// Determine the identifier used in code: alias if provided; otherwise derive from path
		var ident string
		if imp.Name != nil && imp.Name.Name != "" {
			ident = imp.Name.Name
		} else {
			// For external packages, Go uses the package name from the module's go.mod
			// For local packages, derive from the last segment of the import path
			if moduleDir != "" && modulePath != "" && strings.HasPrefix(path, modulePath) {
				// Local package - derive from path
				if i := strings.LastIndex(path, "/"); i >= 0 {
					ident = path[i+1:]
				} else {
					ident = path
				}
			} else {
				// External package - use last segment as initial identifier
				// The actual package name will be resolved later by reading from source
				if i := strings.LastIndex(path, "/"); i >= 0 {
					ident = path[i+1:]
				} else {
					ident = path
				}
			}
		}

		// Skip blank or dot imports for our purposes
		if ident == "_" || ident == "." {
			continue
		}

		// Try to read the actual package name from source files
		realName := ident // Default to the identifier
		if moduleDir != "" && modulePath != "" && strings.HasPrefix(path, modulePath) {
			// Local package - read from local directory
			rel := strings.TrimPrefix(path, modulePath)
			rel = strings.TrimPrefix(rel, "/")
			pkgDir := filepath.Join(moduleDir, filepath.FromSlash(rel))
			if name := readPackageName(pkgDir); name != "" {
				realName = name
			}
		} else {
			// External package - use go list to get the actual package name
			// This is the most reliable way to get the package name
			if pkgName := getPackageNameFromGoList(path, moduleDir); pkgName != "" {
				realName = pkgName
			} else if pkgName := readPackageNameFromImportPath(path); pkgName != "" {
				// Fallback: try to read from module cache
				realName = pkgName
			}
		}

		// Store under both the identifier used in code and the real package name
		// (helps when selector uses real name and import used alias, or vice versa)
		idx[ident] = &TypePackageOutput{Path: path, Name: realName}
		if ident != realName {
			// Always store/update the entry under realName with the path to ensure it's available
			// This ensures lookups by package name can find the correct import path
			existing, exists := idx[realName]
			if !exists || existing.Path == "" {
				idx[realName] = &TypePackageOutput{Path: path, Name: realName}
			}
		}
	}
	return idx, modulePath
}

// locateGoModule walks up from the current file to find a go.mod and returns (moduleDir, modulePath)
func locateGoModule(currentFilePath string) (string, string) {
	dir := filepath.Dir(currentFilePath)
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Read module path
			data, err := os.ReadFile(goModPath)
			if err != nil {
				return dir, ""
			}
			lines := strings.Split(string(data), "\n")
			for _, l := range lines {
				l = strings.TrimSpace(l)
				if strings.HasPrefix(l, "module ") {
					return dir, strings.TrimSpace(strings.TrimPrefix(l, "module "))
				}
			}
			return dir, ""
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", ""
}

// readPackageName parses any .go file in the directory to get the declared package name
func readPackageName(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		filePath := filepath.Join(dir, name)
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, filePath, nil, parser.PackageClauseOnly)
		if err != nil || node == nil {
			continue
		}
		if node.Name != nil && node.Name.Name != "" {
			return node.Name.Name
		}
	}
	return ""
}

// importPathHasSegment returns true if the import path contains the given segment as a path element
func importPathHasSegment(path string, seg string) bool {
	if seg == "" {
		return false
	}
	parts := strings.Split(path, "/")
	for _, p := range parts {
		if p == seg {
			return true
		}
	}
	return false
}

// getPackageNameFromGoList uses `go list` to get the actual package name for an import path.
// This is the most reliable way to get the package name for external packages.
func getPackageNameFromGoList(importPath string, moduleDir string) string {
	// Use go list to get the package name
	// This works for any import path, including versioned modules
	cmd := exec.Command("go", "list", "-f", "{{.Name}}", importPath)
	if moduleDir != "" {
		cmd.Dir = moduleDir
	}
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	name := strings.TrimSpace(string(output))
	if name != "" && name != "main" {
		return name
	}
	return ""
}

// readPackageNameFromImportPath attempts to read the actual package name from an external import path
// by looking in the Go module cache. Returns empty string if not found or not accessible.
// This is a fallback when go list is not available or fails.
func readPackageNameFromImportPath(importPath string) string {
	// Try to find the package in GOPATH/pkg/mod or GOMODCACHE
	// This is a best-effort approach
	gopath := os.Getenv("GOPATH")
	gomodcache := os.Getenv("GOMODCACHE")

	var searchPaths []string
	if gomodcache != "" {
		searchPaths = append(searchPaths, gomodcache)
	}
	if gopath != "" {
		searchPaths = append(searchPaths, filepath.Join(gopath, "pkg", "mod"))
	}

	for _, basePath := range searchPaths {
		// For versioned modules like github.com/gofrs/uuid/v5, the cache structure is:
		// basePath/github.com/gofrs/uuid@v5.4.0/v5
		// We need to find the module directory and then the versioned subdirectory

		// Split the import path to find the module base
		parts := strings.Split(importPath, "/")
		if len(parts) < 2 {
			continue
		}

		// Try to find the module directory (e.g., github.com/gofrs/uuid@v5.4.0)
		// by looking for directories that match the module pattern
		moduleBase := strings.Join(parts[:len(parts)-1], "/")
		lastSegment := parts[len(parts)-1]

		// Check if last segment is a version suffix (v1, v2, v5, etc.)
		isVersionSuffix := len(lastSegment) >= 2 && lastSegment[0] == 'v' &&
			func() bool {
				for i := 1; i < len(lastSegment); i++ {
					if lastSegment[i] < '0' || lastSegment[i] > '9' {
						return false
					}
				}
				return true
			}()

		if isVersionSuffix {
			// Look for module directories matching the base path
			moduleBaseDir := filepath.Join(basePath, filepath.FromSlash(moduleBase))
			parentDir := filepath.Dir(moduleBaseDir)
			moduleName := filepath.Base(moduleBaseDir)

			if entries, err := os.ReadDir(parentDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						continue
					}
					// Look for directories like "uuid@v5.4.0" or "uuid@v5.x.x"
					entryName := entry.Name()
					if strings.HasPrefix(entryName, moduleName+"@") {
						// Found the module directory, now look for the versioned subdirectory
						moduleDir := filepath.Join(parentDir, entryName)
						versionedPath := filepath.Join(moduleDir, lastSegment)
						if name := readPackageName(versionedPath); name != "" {
							return name
						}
						// Also try the module root in case the package is at the root
						if name := readPackageName(moduleDir); name != "" {
							return name
						}
					}
				}
			}
		}

		// Fallback: try the import path directly (for non-versioned or different structures)
		pkgPath := filepath.Join(basePath, filepath.FromSlash(importPath))
		if name := readPackageName(pkgPath); name != "" {
			return name
		}
	}

	return ""
}
