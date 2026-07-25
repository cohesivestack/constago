---
title: go:generate Workflow
description: Pin Constago and regenerate checked-in Go source consistently with go generate.
---

Go's generation convention makes the command discoverable next to the types it
serves.

## Add a directive

Place a directive in any non-generated file in the package or module where you
want developers to start generation:

```go title="generate.go"
package model

//go:generate go run github.com/cohesivestack/constago@v0.0.7 --config ../constago.yaml
```

Paths are resolved from the package directory where `go generate` runs. Set
`input.dir` in the YAML accordingly.

For a repository-wide configuration stored at the module root:

```yaml title="constago.yaml"
input:
  dir: "."
  include: ["**/*.go"]
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
```

Run from the module root:

```bash
go generate ./...
gofmt -w .
go test ./...
```

## Check generated code in CI

Commit generated files, regenerate them in CI, and fail if the working tree
changes:

```bash
go generate ./...
gofmt -w .
go test ./...
git diff --exit-code
```

Pinning `@v0.0.7` makes local and CI output use the same generator release.

## Avoid scanning generated files

Constago does not automatically ignore `output.file_name`. Exclude the current
generated filename and any older generated filenames:

```yaml
input:
  exclude:
    - "**/*_test.go"
    - "**/constago.gen.go"
    - "**/legacy_fields.gen.go"
```
