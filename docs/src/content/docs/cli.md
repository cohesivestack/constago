---
title: CLI Reference
description: Run the Constago command with YAML configuration and command-line overrides.
---

The `constago` command loads configuration, scans Go source files, and writes one
generated file in every package containing selected structs.

```bash
constago [flags]
```

## Configuration lookup

With no `--config` flag, Constago looks for `constago.yaml` in the current
working directory. A missing default file is allowed; explicit flags and built-in
defaults are still applied.

Use a specific file with:

```bash
constago --config ./config/constago.yaml
```

Only changed flags override YAML values. Because elements and getters are
nested collections, define them in YAML.

## Environment overrides

Environment variables use the `CONSTAGO_` prefix and replace dots in
configuration keys with underscores. For example:

```bash
CONSTAGO_VERBOSE=2 \
CONSTAGO_OUTPUT_FILE_NAME=fields.gen.go \
constago --config constago.yaml
```

Environment overrides are most reliable for keys already represented in the
YAML file. Use YAML for `elements` and `getters`, and use flags when you need a
guaranteed one-off override.

## Flags

| Flag | Value | Purpose |
| --- | --- | --- |
| `--config` | path | YAML configuration file |
| `--verbose` | `0`, `1`, or `2` | Silent, steps, or steps plus element details |
| `--input.dir` | path | Root directory to scan |
| `--input.include` | string list | Include Go globs or `package:NAME` selectors |
| `--input.exclude` | string list | Exclude Go globs or `package:NAME` selectors |
| `--input.struct.explicit` | boolean | Require `//constago:include` on structs |
| `--input.struct.include_unexported` | boolean | Include unexported structs |
| `--input.struct.only` | regex | Include only matching struct names |
| `--input.struct.except` | regex | Exclude matching struct names |
| `--input.field.explicit` | boolean | Require a `constago` tag on fields |
| `--input.field.include_unexported` | boolean | Include unexported fields |
| `--input.field.only` | regex | Include only matching field names |
| `--input.field.except` | regex | Exclude matching field names |
| `--output.file_name` | filename | Generated Go filename, without directories |

Pass multiple include or exclude values as a comma-separated list:

```bash
constago \
  --input.dir . \
  --input.include '**/*.go,package:model' \
  --input.exclude '**/*_test.go,package:fixtures' \
  --output.file_name fields.gen.go
```

Quote glob patterns so your shell does not expand them before Constago receives
them.

## Verbosity

- `0` produces no progress output.
- `1` reports scanning and generation steps. This is the default.
- `2` also reports selected generated elements.

## Exit behavior

Configuration, scanning, or generation failures cause the command to return an
error. The current executable reports that error as a panic, so CI should rely
on the non-zero exit status rather than parse its text.

## Inspect help

The installed binary is the source of truth for flags:

```bash
constago --help
```
