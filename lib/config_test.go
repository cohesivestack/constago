package constago

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/cohesivestack/valgo"
	"github.com/stretchr/testify/assert"
)

func TestConfigLoad(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		expectError bool
		checkConfig func(*testing.T, *Config)
	}{
		{
			name: "valid minimal config",
			yamlContent: `
output:
  file_name: "test_gen.go"
input:
  dir: "."
  include:
    - "**/*.go"
elements:
  - name: "field"
    input:
      mode: "tagThenField"
      tag_priority:
        - "json"
        - "field"
    output:
      mode: "constant"
      format:
        holder: "pascal"
        struct: "pascal"
        prefix: "Field"
        suffix: "Const"
      transform:
        tag_values: false
        value_case: "asIs"
        value_separator: "_"
getters:
  - name: "validator"
    returns:
      - "field"
    output:
      prefix: "Get"
      suffix: "Validator"
      format: "pascal"
`,
			expectError: false,
			checkConfig: func(t *testing.T, config *Config) {
				assert.Equal(t, "test_gen.go", config.Output.FileName)
				assert.Equal(t, ".", config.Input.Dir)
				assert.Equal(t, []string{"**/*.go"}, config.Input.Include)
				assert.Len(t, config.Elements, 1)
				assert.Equal(t, "field", config.Elements[0].Name)
				assert.Len(t, config.Getters, 1)
				assert.Equal(t, "validator", config.Getters[0].Name)
			},
		},
		{
			name: "valid full config",
			yamlContent: `
output:
  file_name: "validators_gen.go"
input:
  dir: "."
  include:
    - "**/*.go"
    - "internal/model/*.go"
    - "package:myapp"
  exclude:
    - "**/*_test.go"
    - "package:examples"
  struct:
    explicit: true
    include_unexported: true
  field:
    explicit: true
    include_unexported: true
elements:
  - name: "field"
    input:
      mode: "tag"
      tag_priority:
        - "json"
        - "field"
    output:
      mode: "constant"
      format:
        holder: "pascal"
        struct: "pascal"
        prefix: "Field"
        suffix: "Const"
      transform:
        tag_values: true
        value_case: "upper"
        value_separator: "_"
  - name: "title"
    input:
      mode: "field"
      tag_priority:
        - "title"
        - "label"
    output:
      mode: "constant"
      format:
        holder: "snake"
        struct: "snake"
        prefix: "Title"
        suffix: "Label"
      transform:
        tag_values: true
        value_case: "pascal"
        value_separator: "_"
getters:
  - name: "validator"
    returns:
      - "field"
      - "title"
    output:
      prefix: "Get"
      suffix: "Validator"
      format: "pascal"
`,
			expectError: false,
			checkConfig: func(t *testing.T, config *Config) {
				assert.Equal(t, "validators_gen.go", config.Output.FileName)
				assert.Equal(t, ".", config.Input.Dir)

				// Check input config
				assert.Equal(t, []string{"**/*.go", "internal/model/*.go", "package:myapp"}, config.Input.Include)
				assert.Equal(t, []string{"**/*_test.go", "package:examples"}, config.Input.Exclude)
				assert.True(t, *config.Input.Struct.Explicit)
				assert.True(t, *config.Input.Struct.IncludeUnexported)
				assert.True(t, *config.Input.Field.Explicit)
				assert.True(t, *config.Input.Field.IncludeUnexported)

				// Check elements
				assert.Len(t, config.Elements, 2)

				// Check field element
				fieldElement := config.Elements[0]
				assert.Equal(t, "field", fieldElement.Name)
				assert.Equal(t, InputModeTypeTag, fieldElement.Input.Mode)
				assert.Equal(t, []string{"json", "field"}, fieldElement.Input.TagPriority)
				assert.Equal(t, OutputModeConstant, fieldElement.Output.Mode)
				assert.Equal(t, ConstantFormatPascal, fieldElement.Output.Format.Holder)
				assert.Equal(t, ConstantFormatPascal, fieldElement.Output.Format.Struct)
				assert.Equal(t, "Field", fieldElement.Output.Format.Prefix)
				assert.Equal(t, "Const", fieldElement.Output.Format.Suffix)
				assert.True(t, *fieldElement.Output.Transform.TagValues)
				assert.Equal(t, TransformCaseUpper, fieldElement.Output.Transform.ValueCase)
				assert.Equal(t, "_", fieldElement.Output.Transform.ValueSeparator)

				// Check title element
				titleElement := config.Elements[1]
				assert.Equal(t, "title", titleElement.Name)
				assert.Equal(t, InputModeTypeField, titleElement.Input.Mode)
				assert.Equal(t, []string{"title", "label"}, titleElement.Input.TagPriority)
				assert.Equal(t, OutputModeConstant, titleElement.Output.Mode)
				assert.Equal(t, ConstantFormatSnake, titleElement.Output.Format.Holder)
				assert.Equal(t, ConstantFormatSnake, titleElement.Output.Format.Struct)
				assert.Equal(t, "Title", titleElement.Output.Format.Prefix)
				assert.Equal(t, "Label", titleElement.Output.Format.Suffix)
				assert.True(t, *titleElement.Output.Transform.TagValues)
				assert.Equal(t, TransformCasePascal, titleElement.Output.Transform.ValueCase)
				assert.Equal(t, "_", titleElement.Output.Transform.ValueSeparator)

				// Check getters
				assert.Len(t, config.Getters, 1)
				validatorGetter := config.Getters[0]
				assert.Equal(t, "validator", validatorGetter.Name)
				assert.Equal(t, []string{"field", "title"}, validatorGetter.Returns)
				assert.Equal(t, "Get", validatorGetter.Output.Prefix)
				assert.Equal(t, "Validator", validatorGetter.Output.Suffix)
				assert.Equal(t, ConstantFormatPascal, validatorGetter.Output.Format)
			},
		},
		{
			name:        "invalid yaml",
			yamlContent: `invalid: yaml: content:`,
			expectError: true,
		},
		{
			name:        "file not found",
			yamlContent: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filename string
			if tt.yamlContent != "" {
				filename = "test_config.yaml"
				err := os.WriteFile(filename, []byte(tt.yamlContent), 0644)
				assert.NoError(t, err)
				defer os.Remove(filename)
			} else {
				filename = "nonexistent.yaml"
			}

			config, err := LoadConfig(filename)

			originalError := errors.Unwrap(err)

			var valgoError *valgo.Error
			if errors.As(originalError, &valgoError) {
				errorStruct, _ := valgoError.MarshalJSONIndent("", "  ")
				fmt.Printf("failed to read config file: %s\n", string(errorStruct))
			}

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				if tt.checkConfig != nil {
					tt.checkConfig(t, config)
				}
			}
		})
	}
}

func TestConfigSetDefaults(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected *Config
	}{
		{
			name:   "empty config sets all defaults",
			config: &Config{},
			expected: &Config{
				Input: ConfigInput{
					Dir:     ".",
					Include: []string{"**/*.go"},
					Exclude: []string{"**/*_test.go"},
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
					FileName: "constago_gen.go",
				},
			},
		},
		{
			name: "partial config preserves existing values",
			config: &Config{
				Input: ConfigInput{
					Dir:     ".",
					Include: []string{"custom/*.go"},
				},
				Output: ConfigOutput{
					FileName: "custom.go",
				},
				Elements: []ConfigTag{
					{
						Name: "field",
						Input: ConfigTagInput{
							Mode: InputModeTypeTag,
						},
					},
				},
				Getters: []ConfigGetter{
					{
						Name:    "validator",
						Returns: []string{"field"},
					},
				},
			},
			expected: &Config{
				Output: ConfigOutput{
					FileName: "custom.go", // preserved
				},
				Input: ConfigInput{
					Include: []string{"custom/*.go"},  // preserved
					Exclude: []string{"**/*_test.go"}, // default
					Dir:     ".",                      // default
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
						Name: "field",
						Input: ConfigTagInput{
							Mode:        InputModeTypeTag,                                        // preserved
							TagPriority: []string{"field", "json", "xml", "yaml", "toml", "sql"}, // default
						},
						Output: ConfigTagOutput{
							Mode: OutputModeConstant,
							Format: ConfigTagOutputFormat{
								Holder: ConstantFormatPascal,
								Struct: ConstantFormatPascal,
								Prefix: "field", // default based on element name
								Suffix: "",
							},
							Transform: ConfigTagOutputTransform{
								TagValues:      boolPtr(false),
								ValueCase:      TransformCaseAsIs,
								ValueSeparator: "",
							},
						},
					},
				},
				Getters: []ConfigGetter{
					{
						Name:    "validator",
						Returns: []string{"field"},
						Output: ConfigGetterOutput{
							Prefix: "validator", // default based on getter name
							Format: ConstantFormatPascal,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.setDefaults()

			// Check input defaults
			assert.Equal(t, tt.expected.Input.Dir, tt.config.Input.Dir)
			assert.Equal(t, tt.expected.Input.Include, tt.config.Input.Include)
			assert.Equal(t, tt.expected.Input.Exclude, tt.config.Input.Exclude)
			assert.Equal(t, tt.expected.Input.Struct.Explicit, tt.config.Input.Struct.Explicit)
			assert.Equal(t, tt.expected.Input.Struct.IncludeUnexported, tt.config.Input.Struct.IncludeUnexported)
			assert.Equal(t, tt.expected.Input.Field.Explicit, tt.config.Input.Field.Explicit)
			assert.Equal(t, tt.expected.Input.Field.IncludeUnexported, tt.config.Input.Field.IncludeUnexported)

			// Check output defaults
			assert.Equal(t, tt.expected.Output.FileName, tt.config.Output.FileName)

			// Check elements defaults
			for i, expectedElement := range tt.expected.Elements {
				element := tt.config.Elements[i]
				assert.Equal(t, expectedElement.Input.Mode, element.Input.Mode)
				assert.Equal(t, expectedElement.Input.TagPriority, element.Input.TagPriority)
				assert.Equal(t, expectedElement.Output.Mode, element.Output.Mode)
				assert.Equal(t, expectedElement.Output.Format.Holder, element.Output.Format.Holder)
				assert.Equal(t, expectedElement.Output.Format.Struct, element.Output.Format.Struct)
				assert.Equal(t, expectedElement.Output.Format.Prefix, element.Output.Format.Prefix)
				assert.Equal(t, expectedElement.Output.Format.Suffix, element.Output.Format.Suffix)
				assert.Equal(t, expectedElement.Output.Transform.TagValues, element.Output.Transform.TagValues)
				assert.Equal(t, expectedElement.Output.Transform.ValueCase, element.Output.Transform.ValueCase)
				assert.Equal(t, expectedElement.Output.Transform.ValueSeparator, element.Output.Transform.ValueSeparator)
			}

			// Check getters defaults
			for i, expectedGetter := range tt.expected.Getters {
				getter := tt.config.Getters[i]
				assert.Equal(t, expectedGetter.Output.Prefix, getter.Output.Prefix)
				assert.Equal(t, expectedGetter.Output.Format, getter.Output.Format)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		errorContains map[string][]string
	}{
		{
			name: "valid config",
			config: &Config{
				Output: ConfigOutput{
					FileName: "test.go",
				},
				Input: ConfigInput{
					Dir:     ".",
					Include: []string{"**/*.go", "model/*.go"},
					Exclude: []string{"**/*_test.go"},
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
						Name: "field",
						Input: ConfigTagInput{
							Mode:        InputModeTypeTagThenField,
							TagPriority: []string{"json", "field"},
						},
						Output: ConfigTagOutput{
							Mode: OutputModeConstant,
							Format: ConfigTagOutputFormat{
								Holder: ConstantFormatCamel,
								Struct: ConstantFormatCamel,
							},
							Transform: ConfigTagOutputTransform{
								TagValues:      boolPtr(false),
								ValueCase:      TransformCaseAsIs,
								ValueSeparator: "",
							},
						},
					},
				},
				Getters: []ConfigGetter{
					{
						Name:    "validator",
						Returns: []string{"field"},
						Output: ConfigGetterOutput{
							Format: ConstantFormatPascal,
						},
					},
				},
			},
		},
		{
			name: "invalid output filename - wrong extension",
			config: &Config{
				Output: ConfigOutput{
					FileName: "test.txt", // should end with .go
				},
			},
			errorContains: map[string][]string{
				"output.file_name": {"File name must be a valid Go filename"},
			},
		},
		{
			name: "invalid output filename - with directory path",
			config: &Config{
				Output: ConfigOutput{
					FileName: "path/test.go", // should not contain directory path
				},
			},
			errorContains: map[string][]string{
				"output.file_name": {"File name must be a valid Go filename"},
			},
		},
		{
			name: "invalid source pattern - no valid pattern",
			config: &Config{
				Output: ConfigOutput{
					FileName: "test.go",
				},
				Input: ConfigInput{
					Include: []string{"invalid-pattern"}, // doesn't contain .go, **, or start with package:
				},
			},
			errorContains: map[string][]string{
				"input.include[0]": {"Source pattern must be a valid source pattern"},
			},
		},
		{
			name: "invalid element name - empty",
			config: &Config{
				Output: ConfigOutput{
					FileName: "test.go",
				},
				Input: ConfigInput{
					Include: []string{"**/*.go"},
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
						Name: "", // empty name
					},
				},
			},
			errorContains: map[string][]string{
				"elements[0].name": {"Name can't be blank"},
			},
		},
		{
			name: "invalid element name - not valid Go identifier",
			config: &Config{
				Output: ConfigOutput{
					FileName: "test.go",
				},
				Input: ConfigInput{
					Include: []string{"**/*.go"},
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
						Name: "123invalid", // invalid identifier
					},
				},
			},
			errorContains: map[string][]string{
				"elements[0].name": {"\"123invalid\" is not a valid Go identifier"},
			},
		},
		{
			name: "invalid element input mode",
			config: &Config{
				Output: ConfigOutput{
					FileName: "test.go",
				},
				Input: ConfigInput{
					Include: []string{"**/*.go"},
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
						Name: "field",
						Input: ConfigTagInput{
							Mode:        "invalid", // should be tag, field, or tagThenField
							TagPriority: []string{"json"},
						},
					},
				},
			},
			errorContains: map[string][]string{
				"elements[0].input.mode": {"\"invalid\" is not a valid Mode, must be tag, field, or tagThenField"},
			},
		},
		{
			name: "invalid element tag priority - empty",
			config: &Config{
				Output: ConfigOutput{
					FileName: "test.go",
				},
				Input: ConfigInput{
					Include: []string{"**/*.go"},
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
						Name: "field",
						Input: ConfigTagInput{
							Mode:        InputModeTypeTagThenField,
							TagPriority: []string{}, // empty - should have at least one
						},
					},
				},
			},
			errorContains: map[string][]string{
				"elements[0].input.tag_priority": {"Tag priority must have at least one element"},
			},
		},
		{
			name: "invalid element tag priority - invalid identifier",
			config: &Config{
				Output: ConfigOutput{
					FileName: "test.go",
				},
				Input: ConfigInput{
					Include: []string{"**/*.go"},
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
						Name: "field",
						Input: ConfigTagInput{
							Mode:        InputModeTypeTagThenField,
							TagPriority: []string{"json", "123invalid"}, // invalid identifier
						},
					},
				},
			},
			errorContains: map[string][]string{
				"elements[0].input.tag_priority[1]": {"\"123invalid\" is not a valid Go identifier"},
			},
		},
		{
			name: "invalid element constant format",
			config: &Config{
				Output: ConfigOutput{
					FileName: "test.go",
				},
				Input: ConfigInput{
					Include: []string{"**/*.go"},
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
						Name: "field",
						Input: ConfigTagInput{
							Mode:        InputModeTypeTagThenField,
							TagPriority: []string{"json"},
						},
						Output: ConfigTagOutput{
							Mode: OutputModeConstant,
							Format: ConfigTagOutputFormat{
								Holder: "invalid", // not in valid list
								Struct: ConstantFormatPascal,
							},
							Transform: ConfigTagOutputTransform{
								TagValues:      boolPtr(false),
								ValueCase:      TransformCaseAsIs,
								ValueSeparator: "",
							},
						},
					},
				},
			},
			errorContains: map[string][]string{
				"elements[0].output.format.holder": {"\"invalid\" is not a valid Holder, must be camel, pascal, snake, snakeUpper"},
			},
		},
		{
			name: "invalid element transform value case",
			config: &Config{
				Output: ConfigOutput{
					FileName: "test.go",
				},
				Input: ConfigInput{
					Include: []string{"**/*.go"},
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
						Name: "field",
						Input: ConfigTagInput{
							Mode:        InputModeTypeTagThenField,
							TagPriority: []string{"json"},
						},
						Output: ConfigTagOutput{
							Mode: OutputModeConstant,
							Format: ConfigTagOutputFormat{
								Holder: ConstantFormatPascal,
								Struct: ConstantFormatPascal,
							},
							Transform: ConfigTagOutputTransform{
								TagValues:      boolPtr(false),
								ValueCase:      "invalid", // not in valid list
								ValueSeparator: "",
							},
						},
					},
				},
			},
			errorContains: map[string][]string{
				"elements[0].output.transform.value_case": {"\"invalid\" is not a valid Value case, must be asIs, camel, pascal, upper, lower, title, sentence"},
			},
		},
		{
			name: "invalid getter name - empty",
			config: &Config{
				Output: ConfigOutput{
					FileName: "test.go",
				},
				Input: ConfigInput{
					Include: []string{"**/*.go"},
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
						Name: "field",
					},
				},
				Getters: []ConfigGetter{
					{
						Name: "", // empty name
					},
				},
			},
			errorContains: map[string][]string{
				"getters[0].name": {"Name can't be blank"},
			},
		},
		{
			name: "invalid getter name - not valid Go identifier",
			config: &Config{
				Output: ConfigOutput{
					FileName: "test.go",
				},
				Input: ConfigInput{
					Include: []string{"**/*.go"},
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
						Name: "field",
					},
				},
				Getters: []ConfigGetter{
					{
						Name: "123invalid", // invalid identifier
					},
				},
			},
			errorContains: map[string][]string{
				"getters[0].name": {"\"123invalid\" is not a valid Go identifier"},
			},
		},
		{
			name: "invalid getter returns - empty",
			config: &Config{
				Output: ConfigOutput{
					FileName: "test.go",
				},
				Input: ConfigInput{
					Include: []string{"**/*.go"},
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
						Name: "field",
					},
				},
				Getters: []ConfigGetter{
					{
						Name:    "validator",
						Returns: []string{}, // empty - should have at least one
					},
				},
			},
			errorContains: map[string][]string{
				"getters[0].returns": {"Returns must have at least one element"},
			},
		},
		{
			name: "invalid getter returns - non-existent element",
			config: &Config{
				Output: ConfigOutput{
					FileName: "test.go",
				},
				Input: ConfigInput{
					Include: []string{"**/*.go"},
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
						Name: "field",
					},
				},
				Getters: []ConfigGetter{
					{
						Name:    "validator",
						Returns: []string{"nonexistent"}, // element doesn't exist
					},
				},
			},
			errorContains: map[string][]string{
				"getters[0].returns[0]": {"Return is not valid"},
			},
		},
		{
			name: "invalid getter format",
			config: &Config{
				Output: ConfigOutput{
					FileName: "test.go",
				},
				Input: ConfigInput{
					Include: []string{"**/*.go"},
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
						Name: "field",
					},
				},
				Getters: []ConfigGetter{
					{
						Name:    "validator",
						Returns: []string{"field"},
						Output: ConfigGetterOutput{
							Format: "invalid", // not in valid list
						},
					},
				},
			},
			errorContains: map[string][]string{
				"getters[0].output.format": {"\"invalid\" is not a valid Format, must be camel, pascal, snake, snakeUpper"},
			},
		},
		{
			name: "missing struct validation",
			config: &Config{
				Output: ConfigOutput{
					FileName: "test.go",
				},
				Input: ConfigInput{
					Include: []string{"**/*.go"},
					Field: ConfigInputField{
						Explicit:          boolPtr(false),
						IncludeUnexported: boolPtr(false),
					},
				},
			},
			errorContains: map[string][]string{
				"input.struct.explicit":           {"Explicit must not be nil"},
				"input.struct.include_unexported": {"Include unexported must not be nil"},
			},
		},
		{
			name: "missing field validation",
			config: &Config{
				Output: ConfigOutput{
					FileName: "test.go",
				},
				Input: ConfigInput{
					Include: []string{"**/*.go"},
					Struct: ConfigInputStruct{
						Explicit:          boolPtr(false),
						IncludeUnexported: boolPtr(false),
					},
				},
			},
			errorContains: map[string][]string{
				"input.field.explicit":           {"Explicit must not be nil"},
				"input.field.include_unexported": {"Include unexported must not be nil"},
			},
		},
		{
			name: "valid config with all source types",
			config: &Config{
				Output: ConfigOutput{
					FileName: "test.go",
				},
				Input: ConfigInput{
					Dir: ".",
					Include: []string{
						"**/*.go",          // glob pattern
						"model/*.go",       // glob pattern
						"package:myapp",    // package reference
						"package:my_app",   // package with underscore
						"package:MyApp",    // package with uppercase
						"package:MyApp123", // package with numbers
					},
					Exclude: []string{
						"model/internal.go", // go file
						"**/test/*.go",      // glob pattern
						"package:test",      // package reference
					},
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
						Name: "field",
						Input: struct {
							Mode        InputModeType `yaml:"mode"`
							TagPriority []string      `yaml:"tag_priority"`
						}{
							Mode:        InputModeTypeTagThenField,
							TagPriority: []string{"json", "field"},
						},
						Output: ConfigTagOutput{
							Mode: OutputModeConstant,
							Format: ConfigTagOutputFormat{
								Holder: ConstantFormatCamel,
								Struct: ConstantFormatCamel,
							},
							Transform: ConfigTagOutputTransform{
								TagValues:      boolPtr(false),
								ValueCase:      TransformCaseAsIs,
								ValueSeparator: "",
							},
						},
					},
				},
				Getters: []ConfigGetter{
					{
						Name:    "validator",
						Returns: []string{"field"},
						Output: ConfigGetterOutput{
							Format: ConstantFormatPascal,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()

			expectError := len(tt.errorContains) > 0
			var valgoError *valgo.Error
			if expectError && assert.Error(t, err) && assert.True(t, errors.As(err, &valgoError)) {
				for field, expected := range tt.errorContains {
					if assert.Contains(t, valgoError.Errors(), field) {
						for _, expected := range expected {
							assert.Contains(t, valgoError.Errors()[field].Messages(), expected)
						}
					}
				}
			} else {
				// Check if it's a nil pointer to valgo.Error (which means validation passed)
				if valgoErr, ok := err.(*valgo.Error); ok && valgoErr == nil {
					// Validation passed, this is expected
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		checkConfig func(*testing.T, *Config)
	}{
		{
			name: "valid config with defaults applied",
			config: &Config{
				Output: ConfigOutput{
					FileName: "test.go",
				},
				Input: ConfigInput{
					Include: []string{"**/*.go"},
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
						Name: "field",
					},
				},
				Getters: []ConfigGetter{
					{
						Name:    "validator",
						Returns: []string{"field"},
					},
				},
			},
			expectError: false,
			checkConfig: func(t *testing.T, config *Config) {
				// Check that defaults were applied
				assert.Equal(t, ".", config.Input.Dir)
				assert.Equal(t, []string{"**/*.go"}, config.Input.Include)
				assert.Equal(t, []string{"**/*_test.go"}, config.Input.Exclude)
				assert.False(t, *config.Input.Struct.Explicit)
				assert.False(t, *config.Input.Struct.IncludeUnexported)
				assert.False(t, *config.Input.Field.Explicit)
				assert.False(t, *config.Input.Field.IncludeUnexported)

				// Check element defaults
				assert.Equal(t, InputModeTypeTagThenField, config.Elements[0].Input.Mode)
				assert.Equal(t, []string{"field", "json", "xml", "yaml", "toml", "sql"}, config.Elements[0].Input.TagPriority)
				assert.Equal(t, OutputModeConstant, config.Elements[0].Output.Mode)
				assert.Equal(t, "field", config.Elements[0].Output.Format.Prefix)
				assert.Equal(t, ConstantFormatPascal, config.Elements[0].Output.Format.Holder)
				assert.Equal(t, ConstantFormatPascal, config.Elements[0].Output.Format.Struct)
				assert.False(t, *config.Elements[0].Output.Transform.TagValues)
				assert.Equal(t, TransformCaseAsIs, config.Elements[0].Output.Transform.ValueCase)

				// Check getter defaults
				assert.Equal(t, "validator", config.Getters[0].Output.Prefix)
				assert.Equal(t, ConstantFormatPascal, config.Getters[0].Output.Format)
			},
		},
		{
			name: "invalid config - validation fails",
			config: &Config{
				Output: ConfigOutput{
					FileName: "test.txt", // invalid extension
				},
			},
			expectError: true,
		},
		{
			name:        "empty config - all defaults applied",
			config:      &Config{},
			expectError: false,
			checkConfig: func(t *testing.T, config *Config) {
				// Check that all defaults were applied
				assert.Equal(t, ".", config.Input.Dir)
				assert.Equal(t, []string{"**/*.go"}, config.Input.Include)
				assert.Equal(t, []string{"**/*_test.go"}, config.Input.Exclude)
				assert.Equal(t, "constago_gen.go", config.Output.FileName)
				assert.False(t, *config.Input.Struct.Explicit)
				assert.False(t, *config.Input.Struct.IncludeUnexported)
				assert.False(t, *config.Input.Field.Explicit)
				assert.False(t, *config.Input.Field.IncludeUnexported)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewConfig(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				if tt.checkConfig != nil {
					tt.checkConfig(t, config)
				}
			}
		})
	}
}
