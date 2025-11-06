package constago

import "fmt"

type PackageModel struct {
	// Package information
	Name string
	Path string

	// Imports to use in the generated code
	Imports map[string]*TypePackageOutput

	// Structs to generate validators for
	Structs []*StructModel
}

// StructInfo represents a struct that should have code to generate
type StructModel struct {
	Name       string
	File       string
	LineNumber int

	// Fields that should have code to generate
	Constants []*ConstantOutput
	Structs   []*StructOutput
	Getters   []*GetterOutput
}

type ScanError struct {
	File    string
	Line    int
	Message string
}

type StructOutput struct {
	Name    string
	Package string

	Fields []*FieldOutput
}

type ConstantOutput struct {
	Name  string
	Value string
}

type FieldOutput struct {
	StructName string
	Name       string
	Value      string
}

type NoneOutput struct {
	Name  string
	Value string
}

type ValueOutput struct {
	FieldName   string
	TypeName    string
	TypePackage *TypePackageOutput
}

type TypePackageOutput struct {
	Path  string
	Name  string
	Alias string
}

type ReturnOutput struct {
	Field    *FieldOutput
	Constant *ConstantOutput
	None     *NoneOutput
	Value    *ValueOutput
}

type GetterOutput struct {
	Name    string
	Returns []*ReturnOutput
}

type Model struct {
	// Packages organized by path
	Packages map[string]*PackageModel

	// Scanning statistics
	FilesScanned  int
	PackagesFound int
	StructsFound  int
	FieldsFound   int

	// Errors encountered during scanning
	Errors []*ScanError
}

func NewModel(config *Config) *Model {
	return &Model{
		Packages: make(map[string]*PackageModel),
	}
}

func (m *Model) AddStruct(packagePath string, packageName string, structModel *StructModel) {

	pkg := m.Packages[packagePath]

	// Initialize package if it doesn't exist
	if pkg == nil {
		pkg = &PackageModel{
			Name:    packageName,
			Path:    packagePath,
			Imports: map[string]*TypePackageOutput{},
			Structs: []*StructModel{},
		}
		m.Packages[packagePath] = pkg
		m.PackagesFound++
	}

	var setRecursiveAlias func(pkg *PackageModel, currentImport *TypePackageOutput, currentNameOrAlias string, level int)
	setRecursiveAlias = func(pkg *PackageModel, currentImport *TypePackageOutput, currentNameOrAlias string, level int) {
		for _, imp := range pkg.Imports {
			if imp.Path != currentImport.Path &&
				(imp.Name == currentNameOrAlias || imp.Alias == currentNameOrAlias) {
				currentImport.Alias = fmt.Sprintf("_%s", currentNameOrAlias)
				setRecursiveAlias(pkg, currentImport, currentImport.Alias, level+1)
			}
		}
	}

	for _, g := range structModel.Getters {
		for _, r := range g.Returns {
			if r.Value != nil {
				if _, exists := pkg.Imports[r.Value.TypePackage.Path]; !exists {
					pkg.Imports[r.Value.TypePackage.Path] = r.Value.TypePackage
					setRecursiveAlias(pkg, r.Value.TypePackage, r.Value.TypePackage.Name, 0)
				}
			}
		}
	}

	pkg.Structs = append(pkg.Structs, structModel)

	m.StructsFound++
}

// AddError appends a scanning error to the model
func (m *Model) AddError(file string, line int, message string) {
	m.Errors = append(m.Errors, &ScanError{
		File:    file,
		Line:    line,
		Message: message,
	})
}
