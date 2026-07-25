---
title: Library API
description: Import Constago's Go package to load configuration, build a scan model, or generate source code programmatically.
---

The command is implemented with the importable package at
`github.com/cohesivestack/constago/lib`. Its declared Go package name is
`constago`.

## Install

```bash
go get github.com/cohesivestack/constago@v0.0.7
```

Import it with an explicit alias if you want the path and package name to be
obvious:

```go
import constago "github.com/cohesivestack/constago/lib"
```

## Generate from YAML

```go
package main

import (
	"log"

	constago "github.com/cohesivestack/constago/lib"
)

func main() {
	cfg, err := constago.LoadConfig("constago.yaml")
	if err != nil {
		log.Fatal(err)
	}

	if err := constago.Generate(cfg); err != nil {
		log.Fatal(err)
	}
}
```

`LoadConfig` reads YAML, applies defaults, and validates the result. `Generate`
also applies defaults and validates, so passing a raw configuration is safe.

## Generate from Go

```go
package main

import (
	"log"

	constago "github.com/cohesivestack/constago/lib"
)

func main() {
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
		log.Fatal(err)
	}
}
```

Pointer booleans in `ConfigInputStruct`, `ConfigInputField`, and
`ConfigTagOutputTransform` distinguish an omitted value from an explicit
`false`. If they are nil, `NewConfig` supplies defaults.

## Validate and inspect defaults

```go
cfg, err := constago.NewConfig(raw)
if err != nil {
	return err
}
```

`NewConfig` mutates and returns the supplied configuration after applying
defaults. Pass a non-nil `*Config`.

## Build a model without writing files

```go
cfg, err := constago.NewConfig(raw)
if err != nil {
	return err
}

model, err := constago.NewModelBuilder(cfg).Build()
if err != nil {
	return err
}

fmt.Println(model.FilesScanned)
fmt.Println(model.PackagesFound)
fmt.Println(model.StructsFound)
```

The returned `Model` exposes packages, selected structs, constants, grouped
accessors, getters, scan counts, and scan errors. This is useful for custom
reporting or inspecting what would be generated.

## Public enum constants

Use exported constants instead of raw strings in Go configuration:

| Category | Constants |
| --- | --- |
| Input mode | `InputModeTypeTag`, `InputModeTypeField`, `InputModeTypeTagThenField` |
| Output mode | `OutputModeNone`, `OutputModeStruct`, `OutputModeConstant` |
| Identifier format | `ConstantFormatCamel`, `ConstantFormatPascal`, `ConstantFormatSnake`, `ConstantFormatSnakeUpper` |
| Value case | `TransformCaseAsIs`, `TransformCaseCamel`, `TransformCasePascal`, `TransformCaseUpper`, `TransformCaseLower` |
