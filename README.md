# Constago

> **⚠️ Warning: This project is in early development stage and is not ready for production use.**

Constago is a Go code generator that scans your codebase for struct types and generates constants, accessor structs, and getter functions derived from struct field tags and names. It helps you replace hard‑coded strings and ad‑hoc reflection with type‑safe, discoverable APIs.

What you can generate:

- Constants or accessor structs for:
  - Tag values of struct fields (e.g., `json`, `xml`, `yaml`, custom tags)
  - Field names themselves (with configurable casing and separators)
- Getter functions that return:
  - Tag values
  - Field names
  - The raw field value at runtime (via the special return token `:value`)

Key features:

- Flexible input selection with include/exclude globs and `package:NAME`
- Opt-in/opt-out controls for which structs and fields are processed
- Tag priority resolution (e.g., prefer `json`, fall back to field name, etc.)
- Configurable output modes: none | constant | struct
- Formatting options for identifiers (camel, pascal, snake, snakeUpper), plus optional prefix/suffix
- Value transformations for generated names (case and word separators)
- Single output file per package (default `constago_gen.go`)

## Quick Example

Given a simple model:

```go
// user.go
package model

type User struct {
    Name    string `json:"name" title:"Name"`
    Country string `json:"country" title:"Country"`
}
```

Configure Constago to generate a Title constant set and a Json accessor struct, plus handy getters that can return the raw field value, a tag value, and a constant label:

```yaml
# constago.yaml
input:
  dir: "."
  include:
    - "**/*.go"
elements:
  - name: "title"
    input:
      mode: "tag"
      tag_priority:
        - "title"
    output:
      mode: "constant"
  - name: "json"
    input:
      mode: "tag"
      tag_priority:
        - "json"
    output:
      mode: "struct"
      format:
        prefix: "Json"
getters:
  - name: "Val"
    returns: [":value", "json", "title"]
    output:
      prefix: "V"
      format: "pascal"
```

Run it in your project folder:

```bash
constago
```

Highlights from the generated file (`constago_gen.go`):

```go
// Constants for User
const (
    TitleUserName    = "Name"
    TitleUserCountry = "Country"
)

// JsonUser contains field constants for User
type JsonUser struct {
    Name    string
    Country string
}

func NewJsonUser() *JsonUser {
    return &JsonUser{
        Name:    "name",
        Country: "country",
    }
}

// VName returns (value, json tag, title) for User.Name
func (_struct *User) VName() (string, string, string) {
    return _struct.Name, "name", "Name"
}
```

## Config File

```yaml
input:
  dir: "." # Default "."
  include: # Where to scan for structs (supports globs and package:NAME). Default "**/*.go"
    - "**/*.go"
    - "internal/model/*.go"
    - "package:myapp"
  exclude: # Files to exclude from scanning. Default: "**/*_test.go"
    - "**/*_test.go"
    - "package:examples"
  struct:
    explicit: false # If false, all structs that are in the files matched by the include configuration will be scanned, unless the directive //constago:exclude is placed above the struct. If true, the directive //constago:include must be placed above the struct. Default: false
    include_unexported: false # If true, unexported structs are included, unless this contains the `//constago:include` directive. Default false

  field:
    explicit: false # If true, only fields with a `constago` tag are included. When false, you can use the tag constago="exclude" to exclude specific fields. Default: false.
    include_unexported: false # If false, unexported fields are ignored unless this contains the `constago` tag. Default: false

output:
  file_name: "cosntago_gen.go" # Output file name for generated functions (must end with .go). The files with the generated functions will be created in the same folder used by the source file. Default: "cosntago_gen.go"

elements:
  - name: "title" # required
    input:
      mode: "tagThenField"         # Mode tag | field | tagThenField. Default tagThenField
      tag_priority:                # Order of tags to read the field name from. Default [field, json, xml, yaml, toml, sql]
        - "field"
        - "json"
        - "xml"
        - "yaml"
        - "toml"
        - "sql"
    output:
      mode: "constant"         # Mode none | constant | struct. Default constant
      format:
        holder: "pascal" # The format if an input.field_name.tag_priority is matched. One of: camel | pascal | snake | snakeUpper. Using pascal or snakeUpper will produce exported constants. Default pascal
        struct: "pascal"
        prefix: # Default is the name of the tag
        suffix: # Default not set
      transform:
        tag_values: false # default false. If this is false then transform_value_case and transform_value_separator only applies when the field_name is taken from the struct field name
        value_case: "asIs" # The case type used when transform the field name value. One of: asIs | camel | pascal | upper | lower | sentence. Default: "asIs"
        value_separator: # The separator between words used when transform the field name value. For example you can get snake case, combining lower case with the _ separator

getters:
  - name: "title"
    returns:
      - "title"
      - ":value"
      # Special return tokens supported: ":value"
    output:
      prefix: "Field" # The default value is the name of the getter
      suffix: # Default not set
      format: "pascal" # The format if an input.field_name.tag_priority is matched. One of: camel | pascal | snake | snakeUpper. Using pascal or snakeUpper will produce exported constants. Default pascal
```