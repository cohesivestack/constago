# Constago

Generate compiler-visible Go code from struct field names and tags.

> [!WARNING]
> Constago is in early development and is not yet recommended for production use. Review generated code before committing it.

[Website and documentation](https://constago.cohesivestack.com)

Constago is a build-time code generator. It scans selected Go structs and turns
field names and tags such as `json`, `xml`, `yaml`, or your own tags into:

- package constants, such as `JsonUserName`;
- grouped field accessors, such as `TitleUser.Name`; and
- methods that return field metadata, the field's runtime value, or both.

This is useful anywhere your application needs stable field identifiers—for
example, validation errors, forms, filters, API payloads, or database mappings.
Instead of repeating string literals such as `"display_name"` throughout the
codebase, you can use generated Go identifiers that are discoverable by your
editor and checked by the compiler.

Constago does not serialize data, validate values, or act as an ORM. It derives
metadata from your structs and generates ordinary Go source code; the generated
API does not require runtime reflection.

## Quick start

Constago requires Go 1.23 or later.

Install the command:

```bash
go install github.com/cohesivestack/constago@latest
```

Given this model:

```go
// user.go
package model

type User struct {
	Name    string `json:"name" title:"Name"`
	Country string `json:"country" title:"Country"`
}
```

Create `constago.yaml` in the directory where you will run Constago:

```yaml
input:
  dir: "."
  include:
    - "**/*.go"
  exclude:
    - "**/*_test.go"
    - "**/constago.gen.go"

output:
  file_name: "constago.gen.go"

elements:
  - name: "json"
    input:
      mode: "tag"
      tag_priority: ["json"]
    output:
      mode: "constant"
      format:
        prefix: "Json"

  - name: "title"
    input:
      mode: "tag"
      tag_priority: ["title"]
    output:
      mode: "struct"
      format:
        prefix: "Title"

getters:
  - name: "metadata"
    returns: ["json", "title", ":value"]
    output:
      prefix: "Metadata"
      format: "pascal"
```

Run:

```bash
constago
```

Constago writes `constago.gen.go` in the selected package. The generated code
includes:

```go
const (
	JsonUserName    = "name"
	JsonUserCountry = "country"
)

var TitleUser = struct {
	Name    string
	Country string
}{
	Name:    "Name",
	Country: "Country",
}

func (_struct *User) MetadataName() (string, string, string) {
	return "name", "Name", _struct.Name
}

func (_struct *User) MetadataCountry() (string, string, string) {
	return "country", "Country", _struct.Country
}
```

You can then refer to metadata as `JsonUserName` or `TitleUser.Name`, and call
`user.MetadataName()` when you need the configured metadata together with the
current field value.

## What you can configure

### Elements

An element defines one string value for every selected field.

- `input.mode: tag` reads the first available tag in `tag_priority`.
- `input.mode: field` derives the value from the Go field name.
- `input.mode: tagThenField` reads tags first and falls back to the field name.
- `output.mode: constant` emits one package constant per field.
- `output.mode: struct` emits a grouped package variable.
- `output.mode: none` emits no standalone value but keeps the element available
  to generated getters.

Generated identifiers support `camel`, `pascal`, `snake`, and `snakeUpper`
formats, optional prefixes and suffixes, and configurable value transformations.

### Getters

A getter generates one method per selected field. Its `returns` list can contain
configured element names and the special `:value` token:

```yaml
getters:
  - name: "key"
    returns: ["json", ":value"]
```

The metadata values are strings. `:value` returns the actual field value using
its declared Go type, including types that require imports.

### Input selection

You can select source files with include/exclude globs or `package:NAME`, filter
struct and field names with the `only` and `except` regular-expression options,
include unexported declarations, or opt individual declarations in and out:

```go
//constago:include
type User struct {
	Name     string `json:"name"`
	Password string `json:"password" constago:"exclude"`
}
```

See [Select Files & Types](https://constago.cohesivestack.com/input-selection/)
for the complete selection rules.

## Configuration and CLI

Without `--config`, the command looks for `constago.yaml` in the current
directory:

```bash
constago
constago --config ./config/constago.yaml
constago --help
```

Basic input and output settings can also be supplied through command-line flags
or `CONSTAGO_*` environment variables. Elements and getters are configured in
YAML.

For all keys, defaults, and valid values, see the
[configuration reference](https://constago.cohesivestack.com/configuration/).
The documentation also covers the
[CLI](https://constago.cohesivestack.com/cli/),
[elements](https://constago.cohesivestack.com/elements/),
[getters](https://constago.cohesivestack.com/getters/), and
[library API](https://constago.cohesivestack.com/library/).

## Generation notes

- Constago creates the configured output filename in each selected package
  directory.
- Add that filename to `input.exclude`; generated files are not excluded
  automatically.
- Struct tag options after the first comma are removed. For example,
  `json:"name,omitempty"` becomes `name`.
- Tag values are otherwise literal: `json:"-"` generates `-` unless you exclude
  that field.
- Format the output and compile your packages after generation:

```bash
gofmt -w constago.gen.go
go test ./...
```

For reproducible generation in a project, see the
[`go generate` guide](https://constago.cohesivestack.com/guides/go-generate/).

## Development

Run the test suite:

```bash
go test ./...
```

## License

Copyright © 2026 Carlos Forero

Constago is developed and maintained by [Cohesive Stack LLC](https://www.cohesivestack.com) and released under the [MIT License](LICENSE).
