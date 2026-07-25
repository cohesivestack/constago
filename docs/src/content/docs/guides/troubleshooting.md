---
title: Troubleshooting
description: Diagnose missing Constago output, invalid configuration, naming problems, and generated-code compilation failures.
---

## No file was generated

Constago only writes a package file when at least one selected struct produces
output.

Check:

1. `input.dir` is relative to the process working directory.
2. At least one `input.include` pattern matches a `.go` file.
3. No exclusion removes the file or package.
4. Struct and field explicit mode is not filtering everything.
5. At least one element has a usable value, or a getter can be produced.

Run with details:

```bash
constago --verbose 2 --config constago.yaml
```

## A tagged field is missing

With `input.mode: tag`, the field produces no value when none of the configured
tags exists or when the selected tag value is empty.

Use `tagThenField` to fall back to the Go field name:

```yaml
input:
  mode: "tagThenField"
  tag_priority: ["json"]
```

## `field` in tag priority did not use the Go field name

Inside `tag_priority`, `field` means a literal struct tag named `field`. To
always use the Go field name, select field mode:

```yaml
input:
  mode: "field"
  tag_priority: ["field"]
```

To prefer a tag and fall back to the Go field name, use `tagThenField`:

```yaml
input:
  mode: "tagThenField"
  tag_priority: ["json"]
```

## JSON `-` appeared in generated output

Constago reads tag text rather than interpreting `encoding/json` rules. Exclude
the field explicitly:

```go
Secret string `json:"-" constago:"exclude"`
```

## Generated code does not compile

Run:

```bash
gofmt -w .
go test ./...
```

Common causes include:

- A generated identifier collides with existing code.
- Two element or getter configurations produce the same name.
- A `camel` or `snake` name is inaccessible from another package.
- A field type refers to code unavailable under the active build tags.
- An old generated file is still present under a different filename.

Constago writes syntactically structured Go but does not run `gofmt` or the Go
compiler itself.

## Configuration validation failed

Check that:

- File globs end in `.go`.
- Package selectors use `package:ValidGoIdentifier`.
- Output filenames end in `.go` and contain no directory separator.
- Every element and getter has a valid Go identifier name.
- Every getter return names an element or is exactly `:value`.
- Regex values compile.
- Enum strings match the casing shown in [Configuration](/configuration/).
