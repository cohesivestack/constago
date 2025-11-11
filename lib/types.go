package constago

const validSourceErrorMessage = "{{title}} must be a valid source pattern"
const validIncludeErrorMessage = "{{title}} must have at least one element"
const validGoIdentifierErrorMessage = "\"{{value}}\" is not a valid Go identifier"

// InputModeType
type InputModeType string

const (
	InputModeTypeTagThenField InputModeType = "tagThenField"
	InputModeTypeField        InputModeType = "field"
	InputModeTypeTag          InputModeType = "tag"
)

var validNameOrTitleModes = []InputModeType{
	InputModeTypeTagThenField,
	InputModeTypeField,
	InputModeTypeTag,
}

// OutputModeType
type OutputModeType string

const (
	OutputModeNone     OutputModeType = "none"
	OutputModeStruct   OutputModeType = "struct"
	OutputModeConstant OutputModeType = "constant"
)

var validOutputModes = []OutputModeType{
	OutputModeNone,
	OutputModeStruct,
	OutputModeConstant,
}

const validOutputModesErrorMessage = "\"{{value}}\" is not a valid {{title}}, must be none, struct, constant"

const validNameOrTitleModesErrorMessage = "\"{{value}}\" is not a valid {{title}}, must be tag, field, or tagThenField"

const validRegexErrorMessage = "{{title}} must be a valid regular expression"

// ConstantFormatType
type ConstantFormatType string

const (
	ConstantFormatCamel      ConstantFormatType = "camel"
	ConstantFormatPascal     ConstantFormatType = "pascal"
	ConstantFormatSnake      ConstantFormatType = "snake"
	ConstantFormatSnakeUpper ConstantFormatType = "snakeUpper"
)

var validConstantFormats = []ConstantFormatType{
	ConstantFormatCamel,
	ConstantFormatPascal,
	ConstantFormatSnake,
	ConstantFormatSnakeUpper,
}

const validConstantFormatsErrorMessage = "\"{{value}}\" is not a valid {{title}}, must be camel, pascal, snake, snakeUpper"

// TransformCaseType
type TransformCaseType string

const (
	TransformCaseAsIs   TransformCaseType = "asIs"
	TransformCaseCamel  TransformCaseType = "camel"
	TransformCasePascal TransformCaseType = "pascal"
	TransformCaseUpper  TransformCaseType = "upper"
	TransformCaseLower  TransformCaseType = "lower"
)

var validTransformCases = []TransformCaseType{
	TransformCaseAsIs,
	TransformCaseCamel,
	TransformCasePascal,
	TransformCaseUpper,
	TransformCaseLower,
}

const validTransformCasesErrorMessage = "\"{{value}}\" is not a valid {{title}}, must be asIs, camel, pascal, upper, lower, title, sentence"
