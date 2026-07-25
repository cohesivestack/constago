---
title: Generate Getters
description: Generate Constago methods that return tag metadata, field-derived values, and runtime Go field values.
---

Getters add one method per selected field to the field's owning struct.

```yaml
getters:
  - name: "metadata"
    returns: ["json", "label", ":value"]
    output:
      prefix: "Metadata"
      suffix: ""
      format: "pascal"
```

Every name in `returns` must refer to a configured element, except the special
`:value` token.

## Return element values

An element can use any output mode. Getters return its computed string even
when the element uses `output.mode: none`:

```yaml
elements:
  - name: "json"
    input:
      mode: "tag"
      tag_priority: ["json"]
    output:
      mode: "none"

getters:
  - name: "key"
    returns: ["json"]
```

For a `Name string \`json:"name"\`` field, this produces:

```go
func (_struct *User) KeyName() string {
	return "name"
}
```

## Return runtime values

`:value` returns the actual field value with its declared Go type:

```yaml
getters:
  - name: "value"
    returns: [":value"]
```

```go
func (_struct *User) ValueCreatedAt() time.Time {
	return _struct.CreatedAt
}
```

Constago adds imports required by generated method signatures, including types
nested in pointers, slices, maps, channels, functions, and generic type
arguments.

## Return multiple values

The order in `returns` becomes the order in the method signature:

```yaml
elements:
  - name: "json"
    input:
      mode: "tag"
      tag_priority: ["json"]
    output:
      mode: "none"
  - name: "label"
    input:
      mode: "tagThenField"
      tag_priority: ["title"]
    output:
      mode: "none"

getters:
  - name: "metadata"
    returns: ["json", "label", ":value"]
```

```go
func (_struct *User) MetadataName() (string, string, string) {
	return "name", "Name", _struct.Name
}
```

If the field's type is not `string`, only the `:value` position changes type.

## Method names

Getter output builds a method name from:

1. `output.prefix`
2. the Go field name
3. `output.suffix`

The combined name is formatted with `camel`, `pascal`, `snake`, or
`snakeUpper`.

```yaml
output:
  prefix: "Get"
  suffix: "Metadata"
  format: "pascal"
```

For `DisplayName`, the method is `GetDisplayNameMetadata`.

:::caution
`camel` and `snake` produce unexported methods. Constago does not check for
collisions with existing methods or between getter configurations, so choose
unique names.
:::
