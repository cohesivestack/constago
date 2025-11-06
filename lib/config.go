package constago

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	v "github.com/cohesivestack/valgo"
	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure for the Constago generator
type Config struct {
	Input    ConfigInput    `yaml:"input"`
	Output   ConfigOutput   `yaml:"output"`
	Elements []ConfigTag    `yaml:"elements"`
	Getters  []ConfigGetter `yaml:"getters"`
}

func (c *Config) validate() error {
	val := v.
		In("input", c.Input.validate()).
		In("output", c.Output.validate()).
		Do(func(val *v.Validation) {
			for i, element := range c.Elements {
				val.InRow("elements", i, element.validate())
			}
		}).
		Do(func(val *v.Validation) {
			elements := make([]string, len(c.Elements))
			for i, element := range c.Elements {
				elements[i] = element.Name
			}
			for i, getter := range c.Getters {
				val.InRow("getters", i, getter.validate(val.IsValid("elements"), elements))
			}
		})

	// Return proper nil interface when validation passes
	if val.Valid() {
		return nil
	}
	return val.ToValgoError()
}

// config.input
type ConfigInput struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`

	Dir string `yaml:"dir"`

	Struct ConfigInputStruct `yaml:"struct"`
	Field  ConfigInputField  `yaml:"field"`
}

type ConfigInputStruct struct {
	Explicit          *bool `yaml:"explicit"`
	IncludeUnexported *bool `yaml:"include_unexported"`
}

func (c *ConfigInputStruct) isExplicit() bool {
	return c.Explicit != nil && *c.Explicit
}

func (c *ConfigInputStruct) isIncludeUnexported() bool {
	return c.IncludeUnexported != nil && *c.IncludeUnexported
}

type ConfigInputField struct {
	Explicit          *bool `yaml:"explicit"`
	IncludeUnexported *bool `yaml:"include_unexported"`
}

func (c *ConfigInputField) isExplicit() bool {
	return c.Explicit != nil && *c.Explicit
}

func (c *ConfigInputField) isIncludeUnexported() bool {
	return c.IncludeUnexported != nil && *c.IncludeUnexported
}

func (c *ConfigInput) validate() *v.Validation {
	isValidSourcePatterns := func(val *v.Validation, field string, sources []string) {
		for i, source := range sources {
			val.InCell(field, i, v.Is(v.String(source, "", "Source pattern").Not().Blank().Passing(isValidSource, validSourceErrorMessage)))
		}
	}

	return v.
		In("struct",
			v.Is(
				v.BoolP(c.Struct.Explicit, "explicit").Not().Nil(),
				v.BoolP(c.Struct.IncludeUnexported, "include_unexported").Not().Nil(),
			),
		).
		Do(func(val *v.Validation) {
			isValidSourcePatterns(val, "include", c.Include)
			isValidSourcePatterns(val, "exclude", c.Exclude)
		}).
		In("field",
			v.Is(
				v.BoolP(c.Field.Explicit, "explicit").Not().Nil(),
				v.BoolP(c.Field.IncludeUnexported, "include_unexported").Not().Nil(),
			),
		)
}

// config.output
type ConfigOutput struct {
	FileName string `yaml:"file_name"`
}

func (c *ConfigOutput) validate() *v.Validation {
	return v.Is(
		v.String(c.FileName, "file_name").Not().Blank().MatchingTo(regexp.MustCompile(`^[^/\\]*\.go$`), "{{title}} must be a valid Go filename"),
	)
}

// config.tags[i]
type ConfigTag struct {
	Name string `yaml:"name"`

	Input  ConfigTagInput  `yaml:"input"`
	Output ConfigTagOutput `yaml:"output"`
}

type ConfigTagInput struct {
	Mode        InputModeType `yaml:"mode"`
	TagPriority []string      `yaml:"tag_priority"`
}

type ConfigTagOutput struct {
	Mode      OutputModeType           `yaml:"mode"`
	Format    ConfigTagOutputFormat    `yaml:"format"`
	Transform ConfigTagOutputTransform `yaml:"transform"`
}

type ConfigTagOutputFormat struct {
	Holder ConstantFormatType `yaml:"holder"`
	Struct ConstantFormatType `yaml:"struct"`
	Prefix string             `yaml:"prefix"`
	Suffix string             `yaml:"suffix"`
}

type ConfigTagOutputTransform struct {
	TagValues      *bool             `yaml:"tag_values"`
	ValueCase      TransformCaseType `yaml:"value_case"`
	ValueSeparator string            `yaml:"value_separator"`
}

func (c *ConfigTag) validate() *v.Validation {
	return v.
		Is(v.String(c.Name, "name").Not().Blank().Passing(isValidGoIdentifier, validGoIdentifierErrorMessage)).
		In("input", v.
			Is(
				v.String(c.Input.Mode, "mode").Not().Blank().InSlice(validNameOrTitleModes, validNameOrTitleModesErrorMessage),
				v.Int(len(c.Input.TagPriority), "tag_priority").Not().LessThan(1, validIncludeErrorMessage),
			).
			Do(func(val *v.Validation) {
				for i, tag := range c.Input.TagPriority {
					val.InCell("tag_priority", i, v.Is(v.String(tag, "", "Tag priority").Passing(isValidGoIdentifier, validGoIdentifierErrorMessage)))
				}
			}),
		).
		In("output", v.
			Is(v.String(c.Output.Mode, "mode").Not().Blank().InSlice(validOutputModes, validOutputModesErrorMessage)).
			In("format", v.Is(
				v.String(c.Output.Format.Holder, "holder").Not().Blank().InSlice(validConstantFormats, validConstantFormatsErrorMessage),
				v.String(c.Output.Format.Struct, "struct").Not().Blank().InSlice(validConstantFormats, validConstantFormatsErrorMessage),
				v.String(c.Output.Format.Prefix, "prefix").Empty().Or().Passing(isValidGoIdentifier, validGoIdentifierErrorMessage),
				v.String(c.Output.Format.Suffix, "suffix").Empty().Or().Passing(isValidGoIdentifier, validGoIdentifierErrorMessage),
			)).
			In("transform", v.Is(
				v.BoolP(c.Output.Transform.TagValues, "tag_values").Not().Nil(),
				v.String(c.Output.Transform.ValueCase, "value_case").Not().Blank().InSlice(validTransformCases, validTransformCasesErrorMessage),
				v.String(c.Output.Transform.ValueSeparator, "value_separator").Empty().Or().Passing(isValidGoIdentifier, validGoIdentifierErrorMessage),
			)),
		)
}

// config.getters[i]
type ConfigGetter struct {
	Name    string             `yaml:"name"`
	Returns []string           `yaml:"returns"`
	Output  ConfigGetterOutput `yaml:"output"`
}

type ConfigGetterOutput struct {
	Prefix string             `yaml:"prefix"`
	Suffix string             `yaml:"suffix"`
	Format ConstantFormatType `yaml:"format"`
}

func (c *ConfigGetter) validate(validElements bool, elements []string) *v.Validation {
	return v.
		Is(v.String(c.Name, "name").Not().Blank().Passing(isValidGoIdentifier, validGoIdentifierErrorMessage)).
		Is(v.Int(len(c.Returns), "returns").Not().LessThan(1, validIncludeErrorMessage)).
		When(validElements, func(val *v.Validation) {
			_elements := append(elements, ":value")
			for i, element := range c.Returns {
				// Special returns like :value don't need to be valid Go identifiers
				if strings.HasPrefix(element, ":") {
					val.InCell("returns", i,
						v.Is(v.String(element, "", "Return").
							Not().Blank().
							InSlice(_elements)),
					)
				} else {
					val.InCell("returns", i,
						v.Is(v.String(element, "", "Return").
							Not().Blank().
							Passing(isValidGoIdentifier, validGoIdentifierErrorMessage).
							InSlice(_elements)),
					)
				}
			}
		}).
		In("output", v.
			Is(
				v.String(c.Output.Prefix, "prefix").Empty().Or().Passing(isValidGoIdentifier, validGoIdentifierErrorMessage),
				v.String(c.Output.Suffix, "suffix").Empty().Or().Passing(isValidGoIdentifier, validGoIdentifierErrorMessage),
				v.String(c.Output.Format, "format").Not().Blank().InSlice(validConstantFormats, validConstantFormatsErrorMessage),
			),
		)
}

// LoadConfig loads and parses the configuration from a YAML file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Set defaults
	config, err = NewConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}

	return config, nil
}

func NewConfig(config *Config) (*Config, error) {
	// Set defaults
	config.setDefaults()

	// Validate configuration
	if err := config.validate(); err != nil {
		switch err.(type) {
		case *v.Error:
			out, _ := (err).(*v.Error).MarshalJSONPretty()
			return nil, fmt.Errorf("config validation failed: %s", string(out))
		default:
			return nil, fmt.Errorf("config validation failed: %w", err)
		}
	}

	return config, nil
}

// setDefaults sets default values for configuration fields
func (config *Config) setDefaults() {
	// Input defaults
	if isStringBlank(config.Input.Dir) {
		config.Input.Dir = "."
	}
	if len(config.Input.Include) == 0 {
		config.Input.Include = []string{"**/*.go"}
	}
	if len(config.Input.Exclude) == 0 {
		config.Input.Exclude = []string{"**/*_test.go"}
	}
	if config.Input.Struct.Explicit == nil {
		config.Input.Struct.Explicit = boolPtr(false)
	}
	if config.Input.Struct.IncludeUnexported == nil {
		config.Input.Struct.IncludeUnexported = boolPtr(false)
	}
	if config.Input.Field.Explicit == nil {
		config.Input.Field.Explicit = boolPtr(false)
	}
	if config.Input.Field.IncludeUnexported == nil {
		config.Input.Field.IncludeUnexported = boolPtr(false)
	}

	// Output defaults
	if isStringBlank(config.Output.FileName) {
		config.Output.FileName = "constago_gen.go"
	}

	for i := range config.Elements {
		element := &config.Elements[i]

		if element.Input.Mode == "" {
			element.Input.Mode = InputModeTypeTagThenField
		}
		if len(element.Input.TagPriority) == 0 {
			element.Input.TagPriority = []string{"field", "json", "xml", "yaml", "toml", "sql"}
		}
		if element.Output.Mode == "" {
			element.Output.Mode = OutputModeConstant
		}
		if element.Output.Format.Holder == "" {
			element.Output.Format.Holder = ConstantFormatPascal
		}
		if element.Output.Format.Struct == "" {
			element.Output.Format.Struct = ConstantFormatPascal
		}
		if isStringBlank(element.Output.Format.Prefix) {
			element.Output.Format.Prefix = element.Name
		}
		if isStringBlank(element.Output.Format.Suffix) {
			element.Output.Format.Suffix = ""
		}
		if element.Output.Transform.TagValues == nil {
			element.Output.Transform.TagValues = boolPtr(false)
		}
		if element.Output.Transform.ValueCase == "" {
			element.Output.Transform.ValueCase = TransformCaseAsIs
		}
		if element.Output.Transform.ValueSeparator == "" {
			element.Output.Transform.ValueSeparator = ""
		}
	}

	for i := range config.Getters {
		getter := &config.Getters[i]

		if isStringBlank(getter.Output.Prefix) {
			getter.Output.Prefix = getter.Name
		}
		if isStringBlank(getter.Output.Format) {
			getter.Output.Format = ConstantFormatPascal
		}
	}
}
