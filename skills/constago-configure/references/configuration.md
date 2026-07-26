# Constago configuration reference

Use this reference for Constago's YAML shape and current configuration
semantics.

## Starter configuration

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

getters:
  - name: "metadata"
    returns: ["json", ":value"]
    output:
      prefix: "Metadata"
      format: "pascal"
```

Constago writes `output.file_name` into each selected package containing output.
Always add that filename to `input.exclude`.

## Root, input, and output

| Key | Default or constraint |
| --- | --- |
| `verbose` | `1`; valid values are `0`, `1`, and `2` |
| `input.dir` | `.`; glob root and scan root |
| `input.include` | `["**/*.go"]` |
| `input.exclude` | `["**/*_test.go"]` |
| `input.struct.explicit` | `false` |
| `input.struct.include_unexported` | `false` |
| `input.struct.only` | empty regex allowlist |
| `input.struct.except` | empty regex denylist |
| `input.field.explicit` | `false` |
| `input.field.include_unexported` | `false` |
| `input.field.only` | empty regex allowlist |
| `input.field.except` | empty regex denylist |
| `output.file_name` | `constago.gen.go`; filename only and must end in `.go` |

Include and exclude entries accept recursive globs or `package:NAME` selectors.
Non-package patterns must end in `.go`. Resolve all patterns from `input.dir`.

Selection proceeds through files, structs, then fields:

- Select exported structs and fields by default.
- Set `struct.explicit: true` to require `//constago:include` immediately above
  selected structs.
- Use `//constago:exclude` immediately above a struct to exclude it.
- Set `field.explicit: true` to require a `constago` field tag.
- Use `constago:"include"` and `constago:"exclude"` for field-level selection.
- Apply `only` first and `except` afterward. Both values are Go regular
  expressions.
- Ignore anonymous embedded fields.
- Let explicit exclusion win over inclusion.

## Elements

Each element computes one string value for every selected field.

```yaml
elements:
  - name: "json"
    input:
      mode: "tagThenField"
      tag_priority: ["json"]
    output:
      mode: "constant"
      format:
        holder: "pascal"
        struct: "pascal"
        prefix: "Json"
        suffix: ""
      transform:
        tag_values: false
        value_case: "asIs"
        value_separator: ""
```

Input modes:

- `tag`: use the first present, non-empty tag in `tag_priority`; omit the element
  for a field with no match.
- `field`: derive the value from the Go field name.
- `tagThenField`: use the first matching tag, then fall back to the Go field
  name.

The word `field` in `tag_priority` refers to a literal struct tag called
`field`. It is not a field-name fallback.

Output modes:

- `constant`: emit one package constant per field.
- `struct`: emit a grouped package variable with one field per selected Go
  field.
- `none`: emit no standalone value but retain the element for getters.

Element defaults:

| Key | Default |
| --- | --- |
| `input.mode` | `tagThenField` |
| `input.tag_priority` | `["field", "json", "xml", "yaml", "toml", "sql"]` |
| `output.mode` | `constant` |
| `output.format.holder` | `pascal` |
| `output.format.struct` | `pascal` |
| `output.format.prefix` | element `name` |
| `output.format.suffix` | empty |
| `output.transform.tag_values` | `false` |
| `output.transform.value_case` | `asIs` |
| `output.transform.value_separator` | empty |

Identifier formats are `camel`, `pascal`, `snake`, and `snakeUpper`. Value cases
are `asIs`, `camel`, `pascal`, `upper`, and `lower`.

Transforms change generated string values, not Go identifiers. Field-derived
values are transformed. Tag-derived values are transformed only when
`tag_values` is true.

Struct tag options after the first comma are removed. For example,
`json:"name,omitempty"` supplies `name`. Other tag values are literal, so
`json:"-"` supplies `-`.

## Getters

Each getter emits one method per selected field:

```yaml
getters:
  - name: "metadata"
    returns: ["json", "label", ":value"]
    output:
      prefix: "Metadata"
      suffix: ""
      format: "pascal"
```

- Require at least one `returns` entry.
- Require each entry to name an element or equal `:value`.
- Preserve `returns` order in the generated method signature.
- Return element metadata as strings.
- Return the actual declared Go field type for `:value`.
- Default `output.prefix` to the getter name, `suffix` to empty, and `format` to
  `pascal`.
- Choose unique method names. Constago does not detect collisions with existing
  methods or other getters.

## Static review checklist

- Validate every element name, getter name, tag name, prefix, and suffix as a Go
  identifier where required.
- Validate regex syntax.
- Validate enum spelling and casing.
- Confirm the output is excluded from input.
- Confirm tag priorities match tags that actually occur in selected Go source.
- Check whether unexported generated names from `camel` or `snake` are intended.
- Check potential constant, variable, and method collisions before generation.
