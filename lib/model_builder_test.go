package constago

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelBuilderFindFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files across packages/dirs
	testFiles := []string{
		"main.go",
		"model/user.go",
		"model/user_test.go",
		"logic/logic.go",
		"logic/logic_test.go",
		"utils/helper.go",
		"utils/helper_test.go",
		"internal/config.go",
	}

	for _, rel := range testFiles {
		filePath := filepath.Join(tempDir, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0755))
		parts := strings.Split(rel, "/")
		pkg := parts[0]
		if len(parts) == 1 {
			pkg = "main"
		}
		content := fmt.Sprintf("package %s\n\n// dummy\n", pkg)
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))
	}

	tests := []struct {
		name          string
		config        *Config
		expectedFiles []string
		expectError   bool
	}{
		{
			name: "include all go files",
			config: &Config{
				Input: ConfigInput{
					Dir:     tempDir,
					Include: []string{"**/*.go"},
					Exclude: []string{},
					Struct: ConfigInputStruct{
						Explicit:          boolPtr(false),
						IncludeUnexported: boolPtr(false),
					},
					Field: ConfigInputField{
						Explicit:          boolPtr(false),
						IncludeUnexported: boolPtr(false),
					},
				},
			},
			expectedFiles: []string{
				filepath.Join(tempDir, "main.go"),
				filepath.Join(tempDir, "model/user.go"),
				filepath.Join(tempDir, "model/user_test.go"),
				filepath.Join(tempDir, "logic/logic.go"),
				filepath.Join(tempDir, "logic/logic_test.go"),
				filepath.Join(tempDir, "utils/helper.go"),
				filepath.Join(tempDir, "utils/helper_test.go"),
				filepath.Join(tempDir, "internal/config.go"),
			},
		},
		{
			name: "exclude patterns and package",
			config: &Config{
				Input: ConfigInput{
					Dir:     tempDir,
					Include: []string{"**/*.go"},
					Exclude: []string{"**/*_test.go", "internal/*.go", "utils/helper.go", "package:logic"},
					Struct: struct {
						Explicit          *bool `yaml:"explicit"`
						IncludeUnexported *bool `yaml:"include_unexported"`
					}{boolPtr(false), boolPtr(false)},
					Field: struct {
						Explicit          *bool `yaml:"explicit"`
						IncludeUnexported *bool `yaml:"include_unexported"`
					}{boolPtr(false), boolPtr(false)},
				},
			},
			expectedFiles: []string{
				filepath.Join(tempDir, "main.go"),
				filepath.Join(tempDir, "model/user.go"),
			},
		},
		{
			name: "invalid glob pattern",
			config: &Config{
				Input: ConfigInput{
					Dir:     tempDir,
					Include: []string{"[invalid"},
					Exclude: []string{},
					Struct: struct {
						Explicit          *bool `yaml:"explicit"`
						IncludeUnexported *bool `yaml:"include_unexported"`
					}{boolPtr(false), boolPtr(false)},
					Field: struct {
						Explicit          *bool `yaml:"explicit"`
						IncludeUnexported *bool `yaml:"include_unexported"`
					}{boolPtr(false), boolPtr(false)},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &modelBuilder{config: tt.config}
			files, err := b.findFiles()
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.expectedFiles, files)
		})
	}
}

func TestModelBuilderFindStructs(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test Go file with structs
	testFile := filepath.Join(tempDir, "user.go")
	content := `package main

type User struct {
	Name string ` + "`json:\"name\" title:\"Name\"`" + `
	Age  int    ` + "`json:\"age\" title:\"Age\"`" + `
	Email string ` + "`json:\"email\" title:\"Email\"`" + `
}

type Company struct {
	Name string ` + "`json:\"name\" title:\"Name\" constago:\"exclude\"`" + `
}

//constago:exclude
type Admin struct {
	User
	Role string ` + "`json:\"role\" title:\"Role\"`" + `
}

//constago:include
type Config struct {
	Host     string ` + "`json:\"host\" title:\"Host\"`" + `
	Port     int    ` + "`json:\"port\" title:\"Port\"`" + `
}

type user struct {
	Name string ` + "`json:\"name\" title:\"Name\"`" + `
	Age  int    ` + "`json:\"age\" title:\"Age\"`" + `
	Email string ` + "`json:\"email\" title:\"Email\"`" + `
}

type company struct {
	Name string ` + "`json:\"name\" title:\"Name\" constago:\"exclude\"`" + `
}

//constago:exclude
type admin struct {
	User
	Role string ` + "`json:\"role\" title:\"Role\"`" + `
}

//constago:include
type config struct {
	Host     string ` + "`json:\"host\" title:\"Host\"`" + `
	Port     int    ` + "`json:\"port\" title:\"Port\"`" + `
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	buildConfig := func() (*Config, error) {
		return NewConfig(&Config{
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
			Elements: []ConfigTag{
				{
					Name: "json",
					Input: ConfigTagInput{
						Mode:        InputModeTypeTagThenField,
						TagPriority: []string{"json"},
					},
				},
				{
					Name: "title",
					Input: ConfigTagInput{
						Mode:        InputModeTypeTagThenField,
						TagPriority: []string{"title"},
					},
				},
			},
		})
	}

	tests := []struct {
		name            string
		setConfig       func(*Config)
		expectedStructs []string
	}{
		{
			name: "not explicit and not include unexported",
			setConfig: func(baseConfig *Config) {
				baseConfig.Input.Struct.Explicit = boolPtr(false)
				baseConfig.Input.Struct.IncludeUnexported = boolPtr(false)
			},
			expectedStructs: []string{"User", "Config", "config"},
		},
		{
			name: "not explicit and include unexported",
			setConfig: func(baseConfig *Config) {
				baseConfig.Input.Struct.Explicit = boolPtr(false)
				baseConfig.Input.Struct.IncludeUnexported = boolPtr(true)
			},
			expectedStructs: []string{"User", "Config", "user", "config"},
		},
		{
			name: "explicit and exclude unexported",
			setConfig: func(baseConfig *Config) {
				baseConfig.Input.Struct.Explicit = boolPtr(true)
				baseConfig.Input.Struct.IncludeUnexported = boolPtr(false)
			},
			expectedStructs: []string{"Config", "config"},
		},
		{
			name: "explicit and include unexported",
			setConfig: func(baseConfig *Config) {
				baseConfig.Input.Struct.Explicit = boolPtr(true)
				baseConfig.Input.Struct.IncludeUnexported = boolPtr(true)
			},
			expectedStructs: []string{"Config", "config"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseConfig, err := buildConfig()
			require.NoError(t, err)

			tt.setConfig(baseConfig)
			builder := NewModelBuilder(baseConfig)

			err = builder.scanFiles()
			require.NoError(t, err)

			assert.Equal(t, 1, builder.model.FilesScanned)

			assert.Equal(t, len(tt.expectedStructs), builder.model.StructsFound)
			var structsFound []string
			for _, structInfo := range builder.model.Packages[tempDir].Structs {
				structsFound = append(structsFound, structInfo.Name)
			}
			assert.ElementsMatch(t, structsFound, tt.expectedStructs)
		})
	}
}

func TestModelBuilderBuildConstants(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test Go file with structs
	testFile := filepath.Join(tempDir, "user.go")
	content := `package main

type User struct {
	Name string ` + "`json:\"name\" title:\"Name\"`" + `
	Age  int ` + "`json:\"age\" title:\"Age\" constago:\"exclude\"`" + `
	Country string ` + "`json:\"country\" title:\"Country\" constago:\"include\"`" + `
	email string ` + "`json:\"email\" title:\"Email\"`" + `
	phone string ` + "`json:\"phone\" title:\"Phone\" constago:\"exclude\"`" + `
	address string ` + "`json:\"address\" title:\"Address\" constago:\"include\"`" + `
}

type Admin struct {
	User
	Role string ` + "`json:\"role\" title:\"Role\"`" + `
}

type Company struct {
	Name string ` + "`json:\"name\" title:\"Name\"`" + `
	Industry string ` + "`json:\"industry\" title:\"Industry\" constago:\"include\"`" + `
	Country string ` + "`json:\"country\" title:\"Country\" constago:\"exclude\"`" + `
	email string ` + "`json:\"email\" title:\"Email\" constago:\"exclude\"`" + `
	phone string ` + "`json:\"phone\" title:\"Phone\" constago:\"include\"`" + `
	address string ` + "`json:\"address\" title:\"Address\"`" + `
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	buildConfig := func() (*Config, error) {
		return NewConfig(&Config{
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
		})
	}

	tests := []struct {
		name              string
		setConfig         func(*Config)
		expectedConstants map[string]map[string]string
	}{
		{
			name: "not explicit and not include unexported",
			setConfig: func(baseConfig *Config) {
				baseConfig.Input.Field.Explicit = boolPtr(false)
				baseConfig.Input.Field.IncludeUnexported = boolPtr(false)
			},
			expectedConstants: map[string]map[string]string{
				"User": {
					"JsonUserName":     "name",
					"TitleUserName":    "Name",
					"JsonUserCountry":  "country",
					"TitleUserCountry": "Country",
					"JsonUserAddress":  "address",
					"TitleUserAddress": "Address",
				},
				"Admin": {
					"JsonAdminRole":  "role",
					"TitleAdminRole": "Role",
				},
				"Company": {
					"JsonCompanyName":      "name",
					"TitleCompanyName":     "Name",
					"JsonCompanyIndustry":  "industry",
					"TitleCompanyIndustry": "Industry",
					"JsonCompanyPhone":     "phone",
					"TitleCompanyPhone":    "Phone",
				},
			},
		},
		{
			name: "explicit and not include unexported",
			setConfig: func(baseConfig *Config) {
				baseConfig.Input.Field.Explicit = boolPtr(true)
				baseConfig.Input.Field.IncludeUnexported = boolPtr(false)
			},
			expectedConstants: map[string]map[string]string{
				"User": {
					"JsonUserCountry":  "country",
					"TitleUserCountry": "Country",
					"JsonUserAddress":  "address",
					"TitleUserAddress": "Address",
				},
				"Company": {
					"JsonCompanyIndustry":  "industry",
					"TitleCompanyIndustry": "Industry",
					"JsonCompanyPhone":     "phone",
					"TitleCompanyPhone":    "Phone",
				},
			},
		},
		{
			name: "not explicit and include unexported",
			setConfig: func(baseConfig *Config) {
				baseConfig.Input.Field.Explicit = boolPtr(false)
				baseConfig.Input.Field.IncludeUnexported = boolPtr(true)
			},
			expectedConstants: map[string]map[string]string{
				"User": {
					"JsonUserName":     "name",
					"TitleUserName":    "Name",
					"JsonUserCountry":  "country",
					"TitleUserCountry": "Country",
					"JsonUserAddress":  "address",
					"TitleUserAddress": "Address",
					"JsonUserEmail":    "email",
					"TitleUserEmail":   "Email",
				},
				"Admin": {
					"JsonAdminRole":  "role",
					"TitleAdminRole": "Role",
				},
				"Company": {
					"JsonCompanyName":      "name",
					"TitleCompanyName":     "Name",
					"JsonCompanyIndustry":  "industry",
					"TitleCompanyIndustry": "Industry",
					"JsonCompanyPhone":     "phone",
					"TitleCompanyPhone":    "Phone",
					"JsonCompanyAddress":   "address",
					"TitleCompanyAddress":  "Address",
				},
			},
		},
		{
			name: "explicit and include unexported",
			setConfig: func(baseConfig *Config) {
				baseConfig.Input.Field.Explicit = boolPtr(true)
				baseConfig.Input.Field.IncludeUnexported = boolPtr(true)
			},
			expectedConstants: map[string]map[string]string{
				"User": {
					"JsonUserCountry":  "country",
					"TitleUserCountry": "Country",
					"JsonUserAddress":  "address",
					"TitleUserAddress": "Address",
				},
				"Company": {
					"JsonCompanyIndustry":  "industry",
					"TitleCompanyIndustry": "Industry",
					"JsonCompanyPhone":     "phone",
					"TitleCompanyPhone":    "Phone",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseConfig, err := buildConfig()
			require.NoError(t, err)
			tt.setConfig(baseConfig)

			scanner := NewModelBuilder(baseConfig)
			require.NoError(t, err)

			err = scanner.scanFile(testFile)
			require.NoError(t, err)

			assert.Len(t, scanner.model.Packages, 1)

			assert.Equal(t, 1, scanner.model.FilesScanned)
			assert.Equal(t, len(tt.expectedConstants), scanner.model.StructsFound)
			assert.Equal(t, len(tt.expectedConstants), len(scanner.model.Packages[tempDir].Structs))

			for _, structModel := range scanner.model.Packages[tempDir].Structs {
				assert.Len(t, structModel.Constants, len(tt.expectedConstants[structModel.Name]))

				assert.Len(t, structModel.Structs, 0)
				assert.Len(t, structModel.Getters, 0)

				for _, constant := range structModel.Constants {
					assert.Equal(t, tt.expectedConstants[structModel.Name][constant.Name], constant.Value)
				}

			}
		})
	}
}

func TestModelBuilderBuildStructs(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test Go file with structs
	testFile := filepath.Join(tempDir, "user.go")
	content := `package main

type User struct {
	Name string ` + "`json:\"name\" title:\"Name\"`" + `
	Age  int ` + "`json:\"age\" title:\"Age\" constago:\"exclude\"`" + `
	Country string ` + "`json:\"country\" title:\"Country\" constago:\"include\"`" + `
	email string ` + "`json:\"email\" title:\"Email\"`" + `
	phone string ` + "`json:\"phone\" title:\"Phone\" constago:\"exclude\"`" + `
	address string ` + "`json:\"address\" title:\"Address\" constago:\"include\"`" + `
}

type Admin struct {
	User
	Role string ` + "`json:\"role\" title:\"Role\"`" + `
}

type Company struct {
	Name string ` + "`json:\"name\" title:\"Name\"`" + `
	Industry string ` + "`json:\"industry\" title:\"Industry\" constago:\"include\"`" + `
	Country string ` + "`json:\"country\" title:\"Country\" constago:\"exclude\"`" + `
	email string ` + "`json:\"email\" title:\"Email\" constago:\"exclude\"`" + `
	phone string ` + "`json:\"phone\" title:\"Phone\" constago:\"include\"`" + `
	address string ` + "`json:\"address\" title:\"Address\"`" + `
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	buildConfig := func() (*Config, error) {
		return NewConfig(&Config{
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
				{
					Name: "title",
					Input: ConfigTagInput{
						Mode:        InputModeTypeTagThenField,
						TagPriority: []string{"title"},
					},
					Output: ConfigTagOutput{
						Mode: OutputModeStruct,
					},
				},
			},
		})
	}

	tests := []struct {
		name            string
		setConfig       func(*Config)
		expectedStructs map[string]map[string]map[string]string
	}{
		{
			name: "not explicit and not include unexported",
			setConfig: func(baseConfig *Config) {
				baseConfig.Input.Field.Explicit = boolPtr(false)
				baseConfig.Input.Field.IncludeUnexported = boolPtr(false)
			},
			expectedStructs: map[string]map[string]map[string]string{
				"User": {
					"JsonUser": {
						"Name":    "name",
						"Country": "country",
						"Address": "address",
					},
					"TitleUser": {
						"Name":    "Name",
						"Country": "Country",
						"Address": "Address",
					},
				},
				"Admin": {
					"JsonAdmin": {
						"Role": "role",
					},
					"TitleAdmin": {
						"Role": "Role",
					},
				},
				"Company": {
					"JsonCompany": {
						"Name":     "name",
						"Industry": "industry",
						"Phone":    "phone",
					},
					"TitleCompany": {
						"Name":     "Name",
						"Industry": "Industry",
						"Phone":    "Phone",
					},
				},
			},
		},
		{
			name: "explicit and not include unexported",
			setConfig: func(baseConfig *Config) {
				baseConfig.Input.Field.Explicit = boolPtr(true)
				baseConfig.Input.Field.IncludeUnexported = boolPtr(false)
			},
			expectedStructs: map[string]map[string]map[string]string{
				"User": {
					"JsonUser": {
						"Country": "country",
						"Address": "address",
					},
					"TitleUser": {
						"Country": "Country",
						"Address": "Address",
					},
				},
				"Company": {
					"JsonCompany": {
						"Industry": "industry",
						"Phone":    "phone",
					},
					"TitleCompany": {
						"Industry": "Industry",
						"Phone":    "Phone",
					},
				},
			},
		},
		{
			name: "not explicit and include unexported",
			setConfig: func(baseConfig *Config) {
				baseConfig.Input.Field.Explicit = boolPtr(false)
				baseConfig.Input.Field.IncludeUnexported = boolPtr(true)
			},
			expectedStructs: map[string]map[string]map[string]string{
				"User": {
					"JsonUser": {
						"Name":    "name",
						"Country": "country",
						"Address": "address",
						"Email":   "email",
					},
					"TitleUser": {
						"Name":    "Name",
						"Country": "Country",
						"Address": "Address",
						"Email":   "Email",
					},
				},
				"Admin": {
					"JsonAdmin": {
						"Role": "role",
					},
					"TitleAdmin": {
						"Role": "Role",
					},
				},
				"Company": {
					"JsonCompany": {
						"Name":     "name",
						"Industry": "industry",
						"Phone":    "phone",
						"Address":  "address",
					},
					"TitleCompany": {
						"Name":     "Name",
						"Industry": "Industry",
						"Phone":    "Phone",
						"Address":  "Address",
					},
				},
			},
		},
		{
			name: "explicit and include unexported",
			setConfig: func(baseConfig *Config) {
				baseConfig.Input.Field.Explicit = boolPtr(true)
				baseConfig.Input.Field.IncludeUnexported = boolPtr(true)
			},
			expectedStructs: map[string]map[string]map[string]string{
				"User": {
					"JsonUser": {
						"Country": "country",
						"Address": "address",
					},
					"TitleUser": {
						"Country": "Country",
						"Address": "Address",
					},
				},
				"Company": {
					"JsonCompany": {
						"Industry": "industry",
						"Phone":    "phone",
					},
					"TitleCompany": {
						"Industry": "Industry",
						"Phone":    "Phone",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseConfig, err := buildConfig()
			require.NoError(t, err)
			tt.setConfig(baseConfig)

			scanner := NewModelBuilder(baseConfig)
			require.NoError(t, err)

			err = scanner.scanFile(testFile)
			require.NoError(t, err)

			assert.Len(t, scanner.model.Packages, 1)

			assert.Equal(t, 1, scanner.model.FilesScanned)
			assert.Equal(t, len(tt.expectedStructs), scanner.model.StructsFound)
			assert.Equal(t, len(tt.expectedStructs), len(scanner.model.Packages[tempDir].Structs))

			for _, structModel := range scanner.model.Packages[tempDir].Structs {
				assert.Len(t, structModel.Structs, len(tt.expectedStructs[structModel.Name]))

				assert.Len(t, structModel.Constants, 0)
				assert.Len(t, structModel.Getters, 0)

				for _, _struct := range structModel.Structs {

					assert.Len(t, _struct.Fields, len(tt.expectedStructs[structModel.Name][_struct.Name]))

					for _, field := range _struct.Fields {
						assert.Equal(t, tt.expectedStructs[structModel.Name][_struct.Name][field.Name], field.Value)
					}
				}

			}
		})
	}
}

func TestModelBuilderBuildGetters(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test Go file with structs
	testFile := filepath.Join(tempDir, "user.go")
	content := `package main

type User struct {
	Name string ` + "`json:\"name\" title:\"Name\" field:\"field_name\"`" + `
	Age  int ` + "`json:\"age\" title:\"Age\"`" + `
	Country string ` + "`json:\"country\" title:\"Country\" field:\"field_country\" constago:\"include\"`" + `
	email string ` + "`json:\"email\" title:\"Email\" field:\"field_email\"`" + `
	phone string ` + "`json:\"phone\" title:\"Phone\" field:\"field_phone\" constago:\"exclude\"`" + `
	address string ` + "`json:\"address\" title:\"Address\" field:\"field_address\" constago:\"include\"`" + `
}

type Admin struct {
	User
	Role string ` + "`json:\"role\" title:\"Role\"`" + `
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	buildConfig := func() (*Config, error) {
		return NewConfig(&Config{
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
						Format: ConfigTagOutputFormat{
							Prefix: "Field",
						},
					},
				},
			},
		})
	}

	tests := []struct {
		name            string
		setConfig       func(*Config)
		expectedGetters map[string]map[string][]ReturnOutput
	}{
		{
			name: "getter with one return",
			setConfig: func(baseConfig *Config) {
				baseConfig.Getters = []ConfigGetter{
					{
						Name:    "Val",
						Returns: []string{"json"},
						Output: ConfigGetterOutput{
							Prefix: "V",
							Format: ConstantFormatPascal,
						},
					},
				}
			},
			expectedGetters: map[string]map[string][]ReturnOutput{
				"User": {
					"VName": {
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "name",
							},
						},
					},
					"VAge": {
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "age",
							},
						},
					},
					"VCountry": {
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "country",
							},
						},
					},
					"VAddress": {
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "address",
							},
						},
					},
				},
				"Admin": {
					"VRole": {
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "role",
							},
						},
					},
				},
			},
		},
		{
			name: "getter with multiple returns",
			setConfig: func(baseConfig *Config) {
				baseConfig.Getters = []ConfigGetter{
					{
						Name:    "Val",
						Returns: []string{"json", "title", "field"},
						Output: ConfigGetterOutput{
							Prefix: "V",
							Format: ConstantFormatPascal,
						},
					},
				}
			},
			expectedGetters: map[string]map[string][]ReturnOutput{
				"User": {
					"VName": {
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "name",
							},
						},
						{
							Constant: &ConstantOutput{
								Name:  "TitleUserName",
								Value: "Name",
							},
						},
						{
							Field: &FieldOutput{
								StructName: "FieldUser",
								Name:       "Name",
								Value:      "field_name",
							},
						},
					},
					"VCountry": {
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "country",
							},
						},
						{
							Constant: &ConstantOutput{
								Name:  "TitleUserCountry",
								Value: "Country",
							},
						},
						{
							Field: &FieldOutput{
								StructName: "FieldUser",
								Name:       "Country",
								Value:      "field_country",
							},
						},
					},
					"VAddress": {
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "address",
							},
						},
						{
							Constant: &ConstantOutput{
								Name:  "TitleUserAddress",
								Value: "Address",
							},
						},
						{
							Field: &FieldOutput{
								StructName: "FieldUser",
								Name:       "Address",
								Value:      "field_address",
							},
						},
					},
				},
			},
		},
		{
			name: "getter with value return",
			setConfig: func(baseConfig *Config) {
				baseConfig.Getters = []ConfigGetter{
					{
						Name:    "Val",
						Returns: []string{":value", "json"},
						Output: ConfigGetterOutput{
							Prefix: "V",
							Format: ConstantFormatPascal,
						},
					},
				}
			},
			expectedGetters: map[string]map[string][]ReturnOutput{
				"User": {
					"VName": {
						{
							Value: &ValueOutput{
								FieldName: "Name",
								TypeName:  "string",
								TypePackage: &TypePackageOutput{
									Path: "",
									Name: "main",
								},
							},
						},
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "name",
							},
						},
					},
					"VAge": {
						{
							Value: &ValueOutput{
								FieldName: "Age",
								TypeName:  "int",
								TypePackage: &TypePackageOutput{
									Path: "",
									Name: "main",
								},
							},
						},
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "age",
							},
						},
					},
					"VCountry": {
						{
							Value: &ValueOutput{
								FieldName: "Country",
								TypeName:  "string",
								TypePackage: &TypePackageOutput{
									Path: "",
									Name: "main",
								},
							},
						},
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "country",
							},
						},
					},
					"VAddress": {
						{
							Value: &ValueOutput{
								FieldName: "address",
								TypeName:  "string",
								TypePackage: &TypePackageOutput{
									Path: "",
									Name: "main",
								},
							},
						},
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "address",
							},
						},
					},
				},
				"Admin": {
					"VRole": {
						{
							Value: &ValueOutput{
								FieldName: "Role",
								TypeName:  "string",
								TypePackage: &TypePackageOutput{
									Path: "",
									Name: "main",
								},
							},
						},
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "role",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseConfig, err := buildConfig()
			require.NoError(t, err)
			tt.setConfig(baseConfig)

			scanner := NewModelBuilder(baseConfig)
			require.NoError(t, err)

			err = scanner.scanFile(testFile)
			require.NoError(t, err)

			assert.Len(t, scanner.model.Packages, 1)

			assert.Equal(t, 1, scanner.model.FilesScanned)

			var structsWithGetters int

			for _, structModel := range scanner.model.Packages[tempDir].Structs {
				if len(structModel.Getters) > 0 {
					structsWithGetters++
				} else {
					continue
				}
				assert.Len(t, structModel.Getters, len(tt.expectedGetters[structModel.Name]))

				for _, getter := range structModel.Getters {
					assert.Len(t, getter.Returns, len(tt.expectedGetters[structModel.Name][getter.Name]))

					for i, returnOutput := range getter.Returns {
						assert.EqualValues(t, tt.expectedGetters[structModel.Name][getter.Name][i].Constant, returnOutput.Constant)
						assert.EqualValues(t, tt.expectedGetters[structModel.Name][getter.Name][i].None, returnOutput.None)
						assert.EqualValues(t, tt.expectedGetters[structModel.Name][getter.Name][i].Value, returnOutput.Value)
						assert.EqualValues(t, tt.expectedGetters[structModel.Name][getter.Name][i].Field, returnOutput.Field)
						if tt.expectedGetters[structModel.Name][getter.Name][i].Value != nil {
							assert.EqualValues(t, tt.expectedGetters[structModel.Name][getter.Name][i].Value.TypePackage, returnOutput.Value.TypePackage)
							assert.EqualValues(t, tt.expectedGetters[structModel.Name][getter.Name][i].Value.TypeName, returnOutput.Value.TypeName)
						}
					}
				}
			}
			assert.Equal(t, len(tt.expectedGetters), structsWithGetters)
		})
	}
}

func TestModelBuilderBuildGetterWithPackageTypeReturn(t *testing.T) {
	tempDir := t.TempDir()

	// Create a temporary module and dependency packages
	// go.mod
	goMod := "module github.com/example\n\ngo 1.22\n\nrequire (\ngopkg.in/yaml.v3 v3.0.1\ngithub.com/gofrs/uuid/v5 v5.4.0\n)\n"
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0644))

	// Run go mod tidy to generate go.sum for external dependencies
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = tempDir
	require.NoError(t, cmd.Run())

	// github.com/example/strings
	stringsDir := filepath.Join(tempDir, "strings")
	require.NoError(t, os.MkdirAll(stringsDir, 0755))
	stringsSrc := "package strings\n\n// String is a sample exported type\ntype String struct{}\n"
	require.NoError(t, os.WriteFile(filepath.Join(stringsDir, "strings.go"), []byte(stringsSrc), 0644))

	// github.com/example/integers
	integersDir := filepath.Join(tempDir, "integers")
	require.NoError(t, os.MkdirAll(integersDir, 0755))
	integersSrc := "package integers\n\n// Integer is a sample exported type\ntype Integer struct{}\n"
	require.NoError(t, os.WriteFile(filepath.Join(integersDir, "integers.go"), []byte(integersSrc), 0644))

	// github.com/example/booleans (aliased as binary in import)
	booleansDir := filepath.Join(tempDir, "booleans")
	require.NoError(t, os.MkdirAll(booleansDir, 0755))
	booleansSrc := "package booleans\n\n// Boolean is a sample exported type\ntype Boolean struct{}\n"
	require.NoError(t, os.WriteFile(filepath.Join(booleansDir, "booleans.go"), []byte(booleansSrc), 0644))

	// github.com/example/booleans (aliased as binary in import)
	floatsDir := filepath.Join(tempDir, "floats", "v1")
	require.NoError(t, os.MkdirAll(floatsDir, 0755))
	floatsSrc := "package floats\n\n// Float is a sample exported type\ntype Float struct{}\n"
	require.NoError(t, os.WriteFile(filepath.Join(floatsDir, "floats.go"), []byte(floatsSrc), 0644))

	// Create a test Go file with structs that import the above packages
	testFile := filepath.Join(tempDir, "user.go")
	content := `package main

import (
  "github.com/example/strings"
  "github.com/example/integers"
	binary "github.com/example/booleans"
  "github.com/example/floats/v1"
	"gopkg.in/yaml.v3"
	"github.com/gofrs/uuid/v5"
)

type User struct {
	ID uuid.UUID ` + "`json:\"id\" title:\"ID\"`" + `
	Name strings.String ` + "`json:\"name\" title:\"Name\" field:\"field_name\"`" + `
	Age  integers.Integer ` + "`json:\"age\" title:\"Age\"`" + `
	Country strings.String ` + "`json:\"country\" title:\"Country\" field:\"field_country\" constago:\"include\"`" + `
	email integers.Integer ` + "`json:\"email\" title:\"Email\" field:\"field_email\"`" + `
	Phone string ` + "`json:\"phone\" title:\"Phone\" field:\"field_phone\"`" + `
	address strings.String ` + "`json:\"address\" title:\"Address\" field:\"field_address\" constago:\"include\"`" + `
	Enabled binary.Boolean ` + "`json:\"enabled\" title:\"Enabled\"`" + `
	Height floats.Float ` + "`json:\"height\" title:\"Height\"`" + `
	Node yaml.Node ` + "`json:\"node\" title:\"Node\"`" + `
}

type Admin struct {
	User
	Role strings.String ` + "`json:\"role\" title:\"Role\"`" + `
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	buildConfig := func() (*Config, error) {
		return NewConfig(&Config{
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
						Format: ConfigTagOutputFormat{
							Prefix: "Field",
						},
					},
				},
			},
		})
	}

	tests := []struct {
		name            string
		setConfig       func(*Config)
		expectedGetters map[string]map[string][]ReturnOutput
	}{
		{
			name: "getter with package type return",
			setConfig: func(baseConfig *Config) {
				baseConfig.Getters = []ConfigGetter{
					{
						Name:    "Val",
						Returns: []string{":value", "json"},
						Output: ConfigGetterOutput{
							Prefix: "V",
							Format: ConstantFormatPascal,
						},
					},
				}
			},
			expectedGetters: map[string]map[string][]ReturnOutput{
				"User": {
					"VId": {
						{
							Value: &ValueOutput{
								FieldName: "ID",
								TypeName:  "uuid.UUID",
								TypePackage: &TypePackageOutput{
									Path:  "github.com/gofrs/uuid/v5",
									Name:  "uuid",
									Alias: "",
								},
							},
						},
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "id",
							},
						},
					},
					"VName": {
						{
							Value: &ValueOutput{
								FieldName: "Name",
								TypeName:  "strings.String",
								TypePackage: &TypePackageOutput{
									Path:  "github.com/example/strings",
									Name:  "strings",
									Alias: "",
								},
							},
						},
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "name",
							},
						},
					},
					"VAge": {
						{
							Value: &ValueOutput{
								FieldName: "Age",
								TypeName:  "integers.Integer",
								TypePackage: &TypePackageOutput{
									Path:  "github.com/example/integers",
									Name:  "integers",
									Alias: "",
								},
							},
						},
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "age",
							},
						},
					},
					"VCountry": {
						{
							Value: &ValueOutput{
								FieldName: "Country",
								TypeName:  "strings.String",
								TypePackage: &TypePackageOutput{
									Path:  "github.com/example/strings",
									Name:  "strings",
									Alias: "",
								},
							},
						},
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "country",
							},
						},
					},
					"VPhone": {
						{
							Value: &ValueOutput{
								FieldName: "Phone",
								TypeName:  "string",
								TypePackage: &TypePackageOutput{
									Path:  "",
									Name:  "main",
									Alias: "",
								},
							},
						},
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "phone",
							},
						},
					},
					"VAddress": {
						{
							Value: &ValueOutput{
								FieldName: "address",
								TypeName:  "strings.String",
								TypePackage: &TypePackageOutput{
									Path:  "github.com/example/strings",
									Name:  "strings",
									Alias: "",
								},
							},
						},
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "address",
							},
						},
					},
					"VEnabled": {
						{
							Value: &ValueOutput{
								FieldName: "Enabled",
								TypeName:  "binary.Boolean",
								TypePackage: &TypePackageOutput{
									Path:  "github.com/example/booleans",
									Name:  "booleans",
									Alias: "",
								},
							},
						},
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "enabled",
							},
						},
					},
					"VHeight": {
						{
							Value: &ValueOutput{
								FieldName: "Height",
								TypeName:  "floats.Float",
								TypePackage: &TypePackageOutput{
									Path:  "github.com/example/floats/v1",
									Name:  "floats",
									Alias: "",
								},
							},
						},
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "height",
							},
						},
					},
					"VNode": {
						{
							Value: &ValueOutput{
								FieldName: "Node",
								TypeName:  "yaml.Node",
								TypePackage: &TypePackageOutput{
									Path:  "gopkg.in/yaml.v3",
									Name:  "yaml",
									Alias: "",
								},
							},
						},
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "node",
							},
						},
					},
				},
				"Admin": {
					"VRole": {
						{
							Value: &ValueOutput{
								FieldName: "Role",
								TypeName:  "strings.String",
								TypePackage: &TypePackageOutput{
									Path:  "github.com/example/strings",
									Name:  "strings",
									Alias: "",
								},
							},
						},
						{
							None: &NoneOutput{
								Name:  "json",
								Value: "role",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseConfig, err := buildConfig()
			require.NoError(t, err)
			tt.setConfig(baseConfig)

			scanner := NewModelBuilder(baseConfig)
			require.NoError(t, err)

			err = scanner.scanFile(testFile)
			require.NoError(t, err)

			assert.Len(t, scanner.model.Packages, 1)

			assert.Equal(t, 1, scanner.model.FilesScanned)

			var structsWithGetters int

			for _, structModel := range scanner.model.Packages[tempDir].Structs {
				if len(structModel.Getters) > 0 {
					structsWithGetters++
				} else {
					continue
				}
				assert.Len(t, structModel.Getters, len(tt.expectedGetters[structModel.Name]))

				for _, getter := range structModel.Getters {
					assert.Len(t, getter.Returns, len(tt.expectedGetters[structModel.Name][getter.Name]))

					for i, returnOutput := range getter.Returns {
						assert.EqualValues(t, tt.expectedGetters[structModel.Name][getter.Name][i].Constant, returnOutput.Constant)
						assert.EqualValues(t, tt.expectedGetters[structModel.Name][getter.Name][i].None, returnOutput.None)
						assert.EqualValues(t, tt.expectedGetters[structModel.Name][getter.Name][i].Value, returnOutput.Value)
						assert.EqualValues(t, tt.expectedGetters[structModel.Name][getter.Name][i].Field, returnOutput.Field)
						if tt.expectedGetters[structModel.Name][getter.Name][i].Value != nil {
							assert.EqualValues(t, tt.expectedGetters[structModel.Name][getter.Name][i].Value.TypePackage, returnOutput.Value.TypePackage)
							assert.EqualValues(t, tt.expectedGetters[structModel.Name][getter.Name][i].Value.TypeName, returnOutput.Value.TypeName)
							assert.EqualValues(t, tt.expectedGetters[structModel.Name][getter.Name][i].Value.TypePackage.Alias, returnOutput.Value.TypePackage.Alias)
						}
					}
				}
			}
			assert.Equal(t, len(tt.expectedGetters), structsWithGetters)
		})
	}
}

func TestModelBuilderBuildConstantsWithTransform(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test Go file with structs
	testFile := filepath.Join(tempDir, "user.go")
	content := `package main

type User struct {
	FirstName string ` + "`json:\"first_name\"`" + `
	LastName string ` + "`json:\"last_name\"`" + `
	Age  int ` + "`json:\"age\"`" + `
	Country string ` + "`json:\"country\"`" + `
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	buildConfig := func() (*Config, error) {
		return NewConfig(&Config{
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
						TagPriority: []string{"json"},
					},
					Output: ConfigTagOutput{
						Mode: OutputModeConstant,
						Transform: ConfigTagOutputTransform{
							TagValues:      boolPtr(true),
							ValueCase:      TransformCasePascal,
							ValueSeparator: " ",
						},
					},
				},
			},
		})
	}

	tests := []struct {
		name              string
		setConfig         func(*Config)
		expectedConstants map[string]map[string]string
	}{
		{
			name:      "Transform value case and separator",
			setConfig: func(baseConfig *Config) {},
			expectedConstants: map[string]map[string]string{
				"User": {
					"JsonUserFirstName":  "first_name",
					"TitleUserFirstName": "First Name",
					"JsonUserLastName":   "last_name",
					"TitleUserLastName":  "Last Name",
					"JsonUserAge":        "age",
					"TitleUserAge":       "Age",
					"JsonUserCountry":    "country",
					"TitleUserCountry":   "Country",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseConfig, err := buildConfig()
			require.NoError(t, err)
			tt.setConfig(baseConfig)

			scanner := NewModelBuilder(baseConfig)
			require.NoError(t, err)

			err = scanner.scanFile(testFile)
			require.NoError(t, err)

			assert.Len(t, scanner.model.Packages, 1)

			assert.Equal(t, 1, scanner.model.FilesScanned)
			assert.Equal(t, len(tt.expectedConstants), scanner.model.StructsFound)
			assert.Equal(t, len(tt.expectedConstants), len(scanner.model.Packages[tempDir].Structs))

			for _, structModel := range scanner.model.Packages[tempDir].Structs {
				assert.Len(t, structModel.Constants, len(tt.expectedConstants[structModel.Name]))

				assert.Len(t, structModel.Structs, 0)
				assert.Len(t, structModel.Getters, 0)

				for _, constant := range structModel.Constants {
					assert.Equal(t, tt.expectedConstants[structModel.Name][constant.Name], constant.Value, fmt.Sprintf("Value expected for %s.%s", structModel.Name, constant.Name))
				}
			}
		})
	}
}
