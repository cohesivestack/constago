---
title: Configuration
description: Configure Constago input selection, generated output, elements, getters, formats, and transforms.
---

Constago accepts YAML configuration through `constago.yaml` or the file passed
to `--config`.

## Complete shape

```yaml
verbose: 1

input:
  dir: "."
  include: ["**/*.go"]
  exclude: ["**/*_test.go"]
  struct:
    explicit: false
    include_unexported: false
    only: ""
    except: ""
  field:
    explicit: false
    include_unexported: false
    only: ""
    except: ""

output:
  file_name: "constago.gen.go"

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
        prefix: "json"
        suffix: ""
      transform:
        tag_values: false
        value_case: "asIs"
        value_separator: ""

getters:
  - name: "metadata"
    returns: ["json", ":value"]
    output:
      prefix: "metadata"
      suffix: ""
      format: "pascal"
```

## Root and input defaults

| Key | Default |
| --- | --- |
| `verbose` | `1` |
| `input.dir` | `.` |
| `input.include` | `["**/*.go"]` |
| `input.exclude` | `["**/*_test.go"]` |
| `input.struct.explicit` | `false` |
| `input.struct.include_unexported` | `false` |
| `input.struct.only` | empty |
| `input.struct.except` | empty |
| `input.field.explicit` | `false` |
| `input.field.include_unexported` | `false` |
| `input.field.only` | empty |
| `input.field.except` | empty |
| `output.file_name` | `constago.gen.go` |

`output.file_name` must be a filename ending in `.go`; directory separators are
rejected. Constago writes that filename into each selected package directory.
Add the filename to `input.exclude`; it is not excluded automatically.

## Element defaults

Defaults are applied independently to every entry in `elements`.

| Key | Default |
| --- | --- |
| `input.mode` | `tagThenField` |
| `input.tag_priority` | `["field", "json", "xml", "yaml", "toml", "sql"]` |
| `output.mode` | `constant` |
| `output.format.holder` | `pascal` |
| `output.format.struct` | `pascal` |
| `output.format.prefix` | the element's `name` |
| `output.format.suffix` | empty |
| `output.transform.tag_values` | `false` |
| `output.transform.value_case` | `asIs` |
| `output.transform.value_separator` | empty |

Entries in `tag_priority` are literal struct tag names. The default `field`
entry therefore reads a tag named `field`. Use `input.mode: field` to always
derive values from Go field names, or `tagThenField` to fall back to them.

## Getter defaults

| Key | Default |
| --- | --- |
| `output.prefix` | the getter's `name` |
| `output.suffix` | empty |
| `output.format` | `pascal` |

Every getter needs at least one `returns` entry. Each entry must name a
configured element or use the special `:value` token.

## Valid values

- Input mode: `tag`, `field`, `tagThenField`
- Output mode: `none`, `constant`, `struct`
- Identifier format: `camel`, `pascal`, `snake`, `snakeUpper`
- Value case: `asIs`, `camel`, `pascal`, `upper`, `lower`
- Verbosity: `0`, `1`, `2`

Element names, tag names, prefixes, and suffixes must be valid Go identifiers.
Include and exclude file patterns must end in `.go`, unless they use the
`package:NAME` form.

## Precedence

Explicit command-line flags override values loaded from YAML. Unchanged flag
defaults do not overwrite YAML, which means `--verbose 0` and boolean `false`
overrides work as expected.
