---
title: Getting Started
description: Install Constago and generate type-safe Go constants, field accessors, and getter methods from a struct.
head:
  - tag: title
    content: Constago — Getting Started
---

Constago scans Go structs and generates type-safe names for their fields and
tags. It can produce package constants, grouped accessor values, and methods
that return metadata or the field's runtime value.

:::caution[Early development]
Constago is not ready for production use. Review generated changes before
committing them.
:::

## Requirements

Constago v0.1.0 requires Go 1.23 or later.

## Install the command

```bash
go install github.com/cohesivestack/constago@v0.1.0
```

Make sure the Go binary directory is on your `PATH`, then verify the command:

```bash
constago --help
```

You can also run a pinned version without installing it:

```bash
go run github.com/cohesivestack/constago@v0.1.0 --help
```

## Create a struct

```go title="user.go"
package model

type User struct {
	Name    string `json:"name" title:"Name"`
	Country string `json:"country" title:"Country"`
}
```

## Configure Constago

Create `constago.yaml` in the directory where you will run the command:

```yaml title="constago.yaml"
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

Keep the generated filename in `input.exclude`. Otherwise, a later run can scan
the previous generated file because the default `**/*.go` include matches it.

## Generate code

Run Constago from the directory containing the configuration file:

```bash
constago
```

The command discovers `constago.yaml` automatically. To use another path:

```bash
constago --config ./config/constago.yaml
```

For the `User` struct, the generated file contains:

```go title="constago.gen.go"
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

Format the generated file and compile the package:

```bash
gofmt -w constago.gen.go
go test ./...
```

## Choose what to generate next

- Read [Configuration](/configuration/) for every key and default.
- Use [Select Files & Types](/input-selection/) to control scanning.
- See [Generate Elements](/elements/) for constants, grouped accessors, and
  transforms.
- See [Generate Getters](/getters/) for metadata and runtime-value methods.
- Import Constago directly with the [Library API](/library/).
