---
title: Select Files & Types
description: Select files, packages, structs, and fields for Constago generation with globs, regexes, directives, and tags.
---

Constago filters input in stages: files, structs, then fields.

## Select files

`input.dir` is the root for glob matching. Include patterns add files, and
exclude patterns remove them.

```yaml
input:
  dir: "."
  include:
    - "**/*.go"
    - "internal/model/*.go"
    - "package:domain"
  exclude:
    - "**/*_test.go"
    - "**/testdata/*.go"
    - "package:fixtures"
```

Patterns support recursive `**` globs. A `package:NAME` selector walks
`input.dir`, parses Go package clauses, and selects files whose declared package
name equals `NAME`.

Exclude the configured output filename explicitly so later runs do not scan
previously generated code:

```yaml
input:
  exclude:
    - "**/*_test.go"
    - "**/constago.gen.go"
```

## Select structs

By default, exported structs are selected. Unexported structs are skipped unless
`include_unexported` is true or the struct has an include directive.

```yaml
input:
  struct:
    explicit: false
    include_unexported: false
    only: "^(User|Account)$"
    except: "DTO$"
```

- `only` is a regular-expression allowlist.
- `except` is a regular-expression denylist applied after `only`.
- `explicit: true` requires an include directive.

Put directives immediately above a type declaration:

```go
//constago:include
type internalRecord struct {
	ID string
}

// constago:exclude
type GeneratedDTO struct {
	ID string
}
```

An include directive can select an unexported struct even when
`include_unexported` is false. An exclude directive always wins. A struct with
both directives is skipped and recorded as a scan error.

## Select fields

By default, exported named fields are selected. Use the `constago` struct tag
for per-field control:

```go
type User struct {
	ID       string `json:"id"`
	Password string `json:"password" constago:"exclude"`
	internal string `json:"internal" constago:"include"`
}
```

```yaml
input:
  field:
    explicit: false
    include_unexported: false
    only: "^(ID|Name|Email)$"
    except: "Password$"
```

`constago:"exclude"` always excludes a field. `constago:"include"` includes it
before exported-name and regex checks.

With `field.explicit: true`, fields need a `constago` tag. In practice, use
`constago:"include"` on selected fields:

```go
type User struct {
	ID   string `json:"id" constago:"include"`
	Name string `json:"name"`
}
```

Anonymous embedded fields are ignored and do not produce elements or getters.
