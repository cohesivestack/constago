---
title: Generate Elements
description: Generate Constago constants, grouped field accessors, or getter-only values from tags and Go field names.
---

An element defines one value derived for every selected struct field.

```yaml
elements:
  - name: "json"
    input:
      mode: "tag"
      tag_priority: ["json"]
    output:
      mode: "constant"
```

## Input modes

### `tag`

Uses the first non-empty tag in `tag_priority`. If no listed tag exists, the
element is omitted for that field.

```yaml
input:
  mode: "tag"
  tag_priority: ["db", "json"]
```

For `json:"display_name,omitempty"`, Constago uses `display_name`; tag options
after the first comma are removed.

### `field`

Always derives the value from the Go field name. `tag_priority` receives its
normal default when omitted, but it is not consulted in this mode.

```yaml
input:
  mode: "field"
  tag_priority: ["field"]
```

### `tagThenField`

Uses the first tag match, then falls back to the field name.

```yaml
input:
  mode: "tagThenField"
  tag_priority: ["json"]
```

:::caution[Tag values are literal]
Constago does not apply encoding-package semantics beyond removing comma
options. For example, `json:"-"` generates the literal value `-`; it does not
skip the field.
:::

## Output modes

Given:

```go
type User struct {
	DisplayName string `json:"display_name"`
}
```

### `constant`

```yaml
output:
  mode: "constant"
  format:
    prefix: "Json"
```

Generates:

```go
const JsonUserDisplayName = "display_name"
```

### `struct`

This mode generates an initialized package variable with a grouped anonymous
struct:

```yaml
output:
  mode: "struct"
  format:
    prefix: "Json"
```

```go
var JsonUser = struct {
	DisplayName string
}{
	DisplayName: "display_name",
}
```

Use it as `JsonUser.DisplayName`.

### `none`

No standalone constant or grouped variable is emitted. The value remains
available to getters that list the element in `returns`.

## Identifier formatting

The element's `format` controls generated names:

```yaml
format:
  holder: "pascal"
  struct: "pascal"
  prefix: "Json"
  suffix: "Key"
```

- `holder` formats fields inside grouped accessor variables.
- `struct` formats top-level constant names and grouped accessor variable names.
- `prefix` and `suffix` wrap the struct and field portions.
- Formats are `camel`, `pascal`, `snake`, and `snakeUpper`.

With `prefix: Json`, `suffix: Key`, and `struct: snakeUpper`, the constant for
`User.DisplayName` is `JSON_USER_DISPLAY_NAME_KEY`.

## Transform values

Transforms affect the generated string value, not its Go identifier:

```yaml
transform:
  tag_values: true
  value_case: "upper"
  value_separator: "_"
```

- `tag_values: false` leaves tag-derived values unchanged.
- Field-name fallbacks are always transformed.
- `value_case` accepts `asIs`, `camel`, `pascal`, `upper`, or `lower`.
- `value_separator` joins detected words with the configured string.

For a `DisplayName` field, `lower` plus `_` produces `display_name`.
