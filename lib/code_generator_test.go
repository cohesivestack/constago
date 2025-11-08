package constago

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate_Constants(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test Go file with structs
	testFile := filepath.Join(tempDir, "user.go")
	content := `package main

type User struct {
	Name string ` + "`json:\"name\" title:\"Name\"`" + `
	Age  int    ` + "`json:\"age\" title:\"Age\"`" + `
	Email string ` + "`json:\"email\" title:\"Email\"`" + `
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	config := &Config{
		Input: ConfigInput{
			Dir: tempDir,
			Struct: ConfigInputStruct{
				Explicit:          boolPtr(false),
				IncludeUnexported: boolPtr(false),
			},
			Field: ConfigInputField{
				Explicit:          boolPtr(false),
				IncludeUnexported: boolPtr(false),
			},
		},
		Output: ConfigOutput{
			FileName: "constants_gen.go",
		},
		Elements: []ConfigTag{
			{
				Name: "json",
				Input: ConfigTagInput{
					Mode:        InputModeTypeTagThenField,
					TagPriority: []string{"json"},
				},
				Output: ConfigTagOutput{
					Mode: OutputModeConstant,
				},
			},
		},
	}

	err := Generate(config)
	require.NoError(t, err)

	// Check that the output file was created
	outputFile := filepath.Join(tempDir, "constants_gen.go")
	assert.FileExists(t, outputFile)

	generated, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	generatedStr := string(generated)

	expectedOutput := `
// Constants for User
const (
	JsonUserName = "name"
	JsonUserAge = "age"
	JsonUserEmail = "email"
)`
	assert.Contains(t, generatedStr, expectedOutput)
}

func TestGenerate_Structs(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "user.go")
	content := `package main

type User struct {
	Name string ` + "`json:\"name\" title:\"Name\"`" + `
	Age  int    ` + "`json:\"age\" title:\"Age\"`" + `
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	config := &Config{
		Input: ConfigInput{
			Dir: tempDir,
			Struct: ConfigInputStruct{
				Explicit:          boolPtr(false),
				IncludeUnexported: boolPtr(false),
			},
			Field: ConfigInputField{
				Explicit:          boolPtr(false),
				IncludeUnexported: boolPtr(false),
			},
		},
		Output: ConfigOutput{
			FileName: "structs_gen.go",
		},
		Elements: []ConfigTag{
			{
				Name: "json",
				Input: ConfigTagInput{
					Mode:        InputModeTypeTagThenField,
					TagPriority: []string{"json"},
				},
				Output: ConfigTagOutput{
					Mode: OutputModeStruct,
				},
			},
		},
	}

	err := Generate(config)
	require.NoError(t, err)

	outputFile := filepath.Join(tempDir, "structs_gen.go")
	assert.FileExists(t, outputFile)

	generated, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	generatedStr := string(generated)

	expectedOutput := `
type JsonUser struct {
	Name string
	Age string
}`
	assert.Contains(t, generatedStr, expectedOutput)
}

func TestGenerate_Getters(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "user.go")
	content := `package main

type User struct {
	Name string ` + "`json:\"name\" title:\"Name\"`" + `
	Age  int    ` + "`json:\"age\" title:\"Age\"`" + `
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	config := &Config{
		Input: ConfigInput{
			Dir: tempDir,
			Struct: ConfigInputStruct{
				Explicit:          boolPtr(false),
				IncludeUnexported: boolPtr(false),
			},
			Field: ConfigInputField{
				Explicit:          boolPtr(false),
				IncludeUnexported: boolPtr(false),
			},
		},
		Output: ConfigOutput{
			FileName: "getters_gen.go",
		},
		Elements: []ConfigTag{
			{
				Name: "json",
				Input: ConfigTagInput{
					Mode:        InputModeTypeTagThenField,
					TagPriority: []string{"json"},
				},
				Output: ConfigTagOutput{
					Mode: OutputModeConstant,
				},
			},
			{
				Name: "title",
				Input: ConfigTagInput{
					Mode:        InputModeTypeTagThenField,
					TagPriority: []string{"title"},
				},
				Output: ConfigTagOutput{
					Mode: OutputModeConstant,
				},
			},
		},
		Getters: []ConfigGetter{
			{
				Name:    "GetFields",
				Returns: []string{"json", "title"},
				Output: ConfigGetterOutput{
					Format: ConstantFormatPascal,
				},
			},
		},
	}

	err := Generate(config)
	require.NoError(t, err)

	outputFile := filepath.Join(tempDir, "getters_gen.go")
	assert.FileExists(t, outputFile)

	generated, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	generatedStr := string(generated)

	expectedOutput := `
// GetFieldsName returns the configured values for User
func (_struct *User) GetFieldsName() (string, string) {
	return "name", "Name"
}`
	assert.Contains(t, generatedStr, expectedOutput)
}

func TestGenerate_WithImports(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "user.go")
	content := `package main

import "strings"

type User struct {
	Name strings.Builder ` + "`json:\"name\"`" + `
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	config := &Config{
		Input: ConfigInput{
			Dir: tempDir,
			Struct: ConfigInputStruct{
				Explicit:          boolPtr(false),
				IncludeUnexported: boolPtr(false),
			},
			Field: ConfigInputField{
				Explicit:          boolPtr(false),
				IncludeUnexported: boolPtr(false),
			},
		},
		Output: ConfigOutput{
			FileName: "imports_gen.go",
		},
		Elements: []ConfigTag{
			{
				Name: "json",
				Input: ConfigTagInput{
					Mode:        InputModeTypeTagThenField,
					TagPriority: []string{"json"},
				},
				Output: ConfigTagOutput{
					Mode: OutputModeNone,
				},
			},
		},
		Getters: []ConfigGetter{
			{
				Name:    "GetValue",
				Returns: []string{":value"},
				Output: ConfigGetterOutput{
					Format: ConstantFormatPascal,
				},
			},
		},
	}

	err := Generate(config)
	require.NoError(t, err)

	outputFile := filepath.Join(tempDir, "imports_gen.go")
	assert.FileExists(t, outputFile)

	generated, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	generatedStr := string(generated)

	expectedOutput := `
import (
	strings "strings"
)`
	assert.Contains(t, generatedStr, expectedOutput)

	expectedGetter := `
// GetValueName returns the configured values for User
func (_struct *User) GetValueName() (strings.Builder) {
	return  _struct.Name
}`
	assert.Contains(t, generatedStr, expectedGetter)
}

func TestGenerate_MultiplePackages(t *testing.T) {
	tempDir := t.TempDir()

	// Create multiple packages
	modelDir := filepath.Join(tempDir, "model")
	serviceDir := filepath.Join(tempDir, "service")
	require.NoError(t, os.MkdirAll(modelDir, 0755))
	require.NoError(t, os.MkdirAll(serviceDir, 0755))

	modelFile := filepath.Join(modelDir, "user.go")
	modelContent := `package model

type User struct {
	Name string ` + "`json:\"name\"`" + `
}
`
	require.NoError(t, os.WriteFile(modelFile, []byte(modelContent), 0644))

	serviceFile := filepath.Join(serviceDir, "service.go")
	serviceContent := `package service

type Service struct {
	Name string ` + "`json:\"name\"`" + `
}
`
	require.NoError(t, os.WriteFile(serviceFile, []byte(serviceContent), 0644))

	config := &Config{
		Input: ConfigInput{
			Dir: tempDir,
			Struct: ConfigInputStruct{
				Explicit:          boolPtr(false),
				IncludeUnexported: boolPtr(false),
			},
			Field: ConfigInputField{
				Explicit:          boolPtr(false),
				IncludeUnexported: boolPtr(false),
			},
		},
		Output: ConfigOutput{
			FileName: "gen.go",
		},
		Elements: []ConfigTag{
			{
				Name: "json",
				Input: ConfigTagInput{
					Mode:        InputModeTypeTagThenField,
					TagPriority: []string{"json"},
				},
				Output: ConfigTagOutput{
					Mode: OutputModeConstant,
				},
			},
		},
	}

	err := Generate(config)
	require.NoError(t, err)

	// Check both output files were created
	modelGen := filepath.Join(modelDir, "gen.go")
	serviceGen := filepath.Join(serviceDir, "gen.go")

	assert.FileExists(t, modelGen)
	assert.FileExists(t, serviceGen)

	modelGenerated, err := os.ReadFile(modelGen)
	require.NoError(t, err)
	assert.Contains(t, string(modelGenerated), `
// Constants for User
const (
	JsonUserName = "name"
)`)

	serviceGenerated, err := os.ReadFile(serviceGen)
	require.NoError(t, err)
	assert.Contains(t, string(serviceGenerated), `
// Constants for Service
const (
	JsonServiceName = "name"
)`)
}

func TestGenerate_SkipEmptyPackages(t *testing.T) {
	tempDir := t.TempDir()

	// Create a package with no structs
	emptyDir := filepath.Join(tempDir, "empty")
	require.NoError(t, os.MkdirAll(emptyDir, 0755))
	emptyFile := filepath.Join(emptyDir, "empty.go")
	emptyContent := `package empty

// No structs here
var x = 1
`
	require.NoError(t, os.WriteFile(emptyFile, []byte(emptyContent), 0644))

	// Create a package with structs
	modelDir := filepath.Join(tempDir, "model")
	require.NoError(t, os.MkdirAll(modelDir, 0755))
	modelFile := filepath.Join(modelDir, "user.go")
	modelContent := `package model

type User struct {
	Name string ` + "`json:\"name\"`" + `
}
`
	require.NoError(t, os.WriteFile(modelFile, []byte(modelContent), 0644))

	config := &Config{
		Input: ConfigInput{
			Dir: tempDir,
			Struct: ConfigInputStruct{
				Explicit:          boolPtr(false),
				IncludeUnexported: boolPtr(false),
			},
			Field: ConfigInputField{
				Explicit:          boolPtr(false),
				IncludeUnexported: boolPtr(false),
			},
		},
		Output: ConfigOutput{
			FileName: "gen.go",
		},
		Elements: []ConfigTag{
			{
				Name: "json",
				Input: ConfigTagInput{
					Mode:        InputModeTypeTagThenField,
					TagPriority: []string{"json"},
				},
				Output: ConfigTagOutput{
					Mode: OutputModeConstant,
				},
			},
		},
	}

	err := Generate(config)
	require.NoError(t, err)

	// Empty package should not have generated file
	emptyGen := filepath.Join(emptyDir, "gen.go")
	assert.NoFileExists(t, emptyGen)

	// Model package should have generated file
	modelGen := filepath.Join(modelDir, "gen.go")
	assert.FileExists(t, modelGen)
}

func TestGenerate_InvalidConfig(t *testing.T) {
	config := &Config{
		Output: ConfigOutput{
			FileName: "test.txt", // Invalid: should be .go
		},
	}

	err := Generate(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create config")
}

func TestGenerate_NestedDirectories(t *testing.T) {
	tempDir := t.TempDir()

	// Create deeply nested package
	nestedDir := filepath.Join(tempDir, "a", "b", "c", "d")
	require.NoError(t, os.MkdirAll(nestedDir, 0755))

	nestedFile := filepath.Join(nestedDir, "nested.go")
	nestedContent := `package d

type Nested struct {
	Value string ` + "`json:\"value\"`" + `
}
`
	require.NoError(t, os.WriteFile(nestedFile, []byte(nestedContent), 0644))

	config := &Config{
		Input: ConfigInput{
			Dir: tempDir,
			Struct: ConfigInputStruct{
				Explicit:          boolPtr(false),
				IncludeUnexported: boolPtr(false),
			},
			Field: ConfigInputField{
				Explicit:          boolPtr(false),
				IncludeUnexported: boolPtr(false),
			},
		},
		Output: ConfigOutput{
			FileName: "generated.go",
		},
		Elements: []ConfigTag{
			{
				Name: "json",
				Input: ConfigTagInput{
					Mode:        InputModeTypeTagThenField,
					TagPriority: []string{"json"},
				},
				Output: ConfigTagOutput{
					Mode: OutputModeConstant,
				},
			},
		},
	}

	err := Generate(config)
	require.NoError(t, err)

	// Check nested directory has generated file
	generatedFile := filepath.Join(nestedDir, "generated.go")
	assert.FileExists(t, generatedFile)

	generated, err := os.ReadFile(generatedFile)
	require.NoError(t, err)
	generatedStr := string(generated)
	expectedOutput := `
// Constants for Nested
const (
	JsonNestedValue = "value"
)`
	assert.Contains(t, generatedStr, expectedOutput)
}

func TestGenerate_WithImportAliases(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "user.go")
	content := `package main

type User struct {
	Name string ` + "`json:\"name\"`" + `
	Age  int    ` + "`json:\"age\"`" + `
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	config := &Config{
		Input: ConfigInput{
			Dir: tempDir,
			Struct: ConfigInputStruct{
				Explicit:          boolPtr(false),
				IncludeUnexported: boolPtr(false),
			},
			Field: ConfigInputField{
				Explicit:          boolPtr(false),
				IncludeUnexported: boolPtr(false),
			},
		},
		Output: ConfigOutput{
			FileName: "aliases_gen.go",
		},
		Elements: []ConfigTag{
			{
				Name: "json",
				Input: ConfigTagInput{
					Mode:        InputModeTypeTagThenField,
					TagPriority: []string{"json"},
				},
				Output: ConfigTagOutput{
					Mode: OutputModeNone,
				},
			},
		},
		Getters: []ConfigGetter{
			{
				Name:    "GetValue",
				Returns: []string{":value"},
				Output: ConfigGetterOutput{
					Format: ConstantFormatPascal,
				},
			},
		},
	}

	err := Generate(config)
	require.NoError(t, err)

	outputFile := filepath.Join(tempDir, "aliases_gen.go")
	assert.FileExists(t, outputFile)

	generated, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	generatedStr := string(generated)
	expectedGetter := `
// GetValueName returns the configured values for User
func (_struct *User) GetValueName() (string) {
	return  _struct.Name
}`
	assert.Contains(t, generatedStr, expectedGetter)
}

func TestGenerate_ComplexScenario(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "user.go")
	content := `package main

type User struct {
	Name    string ` + "`json:\"name\" title:\"Full Name\" field:\"name_field\"`" + `
	Email   string ` + "`json:\"email\" title:\"Email Address\"`" + `
	Age     int    ` + "`json:\"age\"`" + `
	Country string ` + "`json:\"country\" constago:\"include\"`" + `
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	config := &Config{
		Input: ConfigInput{
			Dir: tempDir,
			Struct: ConfigInputStruct{
				Explicit:          boolPtr(false),
				IncludeUnexported: boolPtr(false),
			},
			Field: ConfigInputField{
				Explicit:          boolPtr(false),
				IncludeUnexported: boolPtr(false),
			},
		},
		Output: ConfigOutput{
			FileName: "complex_gen.go",
		},
		Elements: []ConfigTag{
			{
				Name: "json",
				Input: ConfigTagInput{
					Mode:        InputModeTypeTagThenField,
					TagPriority: []string{"json"},
				},
				Output: ConfigTagOutput{
					Mode: OutputModeConstant,
				},
			},
			{
				Name: "title",
				Input: ConfigTagInput{
					Mode:        InputModeTypeTagThenField,
					TagPriority: []string{"title"},
				},
				Output: ConfigTagOutput{
					Mode: OutputModeConstant,
				},
			},
			{
				Name: "field",
				Input: ConfigTagInput{
					Mode:        InputModeTypeTag,
					TagPriority: []string{"field"},
				},
				Output: ConfigTagOutput{
					Mode: OutputModeStruct,
				},
			},
		},
		Getters: []ConfigGetter{
			{
				Name:    "GetAll",
				Returns: []string{"json", "title", "field"},
				Output: ConfigGetterOutput{
					Format: ConstantFormatPascal,
				},
			},
			{
				Name:    "GetValue",
				Returns: []string{":value", "json"},
				Output: ConfigGetterOutput{
					Format: ConstantFormatPascal,
				},
			},
		},
	}

	err := Generate(config)
	require.NoError(t, err)

	outputFile := filepath.Join(tempDir, "complex_gen.go")
	assert.FileExists(t, outputFile)

	generated, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	generatedStr := string(generated)

	expectedBlock := `
// Constants for User
const (
	JsonUserName = "name"
	TitleUserName = "Full Name"
	JsonUserEmail = "email"
	TitleUserEmail = "Email Address"
	JsonUserAge = "age"
	TitleUserAge = "Age"
	JsonUserCountry = "country"
	TitleUserCountry = "Country"
)
// FieldUser contains field constants for User
type FieldUser struct {
	Name string
}

// NewFieldUser creates a new FieldUser instance
func NewFieldUser() *FieldUser {
	return &FieldUser{
		Name: "name_field",
	}
}
// GetAllName returns the configured values for User
func (_struct *User) GetAllName() (string, string, string) {
	return "name", "Full Name", "name_field"
}`
	assert.Contains(t, generatedStr, expectedBlock)
}
