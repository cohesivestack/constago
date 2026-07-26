# Constago Go library reference

Use this reference for the importable Constago API.

## Module and import

Add the selected released version through the consuming project's normal module
workflow:

```bash
go get github.com/cohesivestack/constago@v0.1.0
```

Treat `v0.1.0` as an example. Preserve an existing version or use the version
selected for the project.

Import the `/lib` path. Its package declaration is `constago`:

```go
import constago "github.com/cohesivestack/constago/lib"
```

## Generate from YAML

```go
package main

import (
	"fmt"

	constago "github.com/cohesivestack/constago/lib"
)

func run() error {
	cfg, err := constago.LoadConfig("constago.yaml")
	if err != nil {
		return fmt.Errorf("load Constago config: %w", err)
	}

	if err := constago.Generate(cfg); err != nil {
		return fmt.Errorf("generate Constago output: %w", err)
	}
	return nil
}
```

`LoadConfig` reads YAML, applies defaults, and validates. `Generate` applies
defaults and validates again before scanning and writing output.

## Generate from Go configuration

```go
cfg := &constago.Config{
	Input: constago.ConfigInput{
		Dir:     ".",
		Include: []string{"**/*.go"},
		Exclude: []string{"**/*_test.go", "**/constago.gen.go"},
	},
	Output: constago.ConfigOutput{
		FileName: "constago.gen.go",
	},
	Elements: []constago.ConfigTag{
		{
			Name: "json",
			Input: constago.ConfigTagInput{
				Mode:        constago.InputModeTypeTag,
				TagPriority: []string{"json"},
			},
			Output: constago.ConfigTagOutput{
				Mode: constago.OutputModeConstant,
			},
		},
	},
}

if err := constago.Generate(cfg); err != nil {
	return fmt.Errorf("generate Constago output: %w", err)
}
```

Omitted fields receive defaults. Pointer booleans and `Verbose *int`
distinguish omitted values from explicit `false` or `0`. Create local pointer
helpers only when an explicit non-default value is required:

```go
func boolPointer(value bool) *bool { return &value }
func intPointer(value int) *int    { return &value }
```

## Validate and inspect defaults

```go
cfg, err := constago.NewConfig(raw)
if err != nil {
	return fmt.Errorf("validate Constago config: %w", err)
}
```

Pass a non-nil `*constago.Config`. `NewConfig` mutates and returns the supplied
configuration after applying defaults.

## Build a model without writing files

```go
cfg, err := constago.NewConfig(raw)
if err != nil {
	return fmt.Errorf("validate Constago config: %w", err)
}

model, err := constago.NewModelBuilder(cfg).Build()
if err != nil {
	return fmt.Errorf("build Constago model: %w", err)
}

fmt.Println(model.FilesScanned)
fmt.Println(model.PackagesFound)
fmt.Println(model.StructsFound)
fmt.Println(model.FieldsFound)
```

The returned `Model` exposes:

- `Packages map[string]*PackageModel`;
- scan counts for files, packages, structs, and fields; and
- `Errors []*ScanError`.

Package and struct models expose the constants, grouped accessors, getters,
imports, source files, and line information discovered during the scan. Building
the model scans inputs but does not write generated output.

## Public enum constants

| Purpose | Constants |
| --- | --- |
| Input modes | `InputModeTypeTag`, `InputModeTypeField`, `InputModeTypeTagThenField` |
| Output modes | `OutputModeNone`, `OutputModeStruct`, `OutputModeConstant` |
| Identifier formats | `ConstantFormatCamel`, `ConstantFormatPascal`, `ConstantFormatSnake`, `ConstantFormatSnakeUpper` |
| Value cases | `TransformCaseAsIs`, `TransformCaseCamel`, `TransformCasePascal`, `TransformCaseUpper`, `TransformCaseLower` |

Prefer these constants to raw strings in Go code.

## Side effects and verification

- Resolve `LoadConfig` paths and `Config.Input.Dir` from the process working
  directory.
- Expect `Generate` to write `Output.FileName` in every selected package that
  produces output.
- Exclude the generated filename from `Input.Exclude`.
- Format generated files after `Generate`.
- Compile and test the affected packages.
- Use `NewModelBuilder` rather than `Generate` for side-effect-free inspection.
