# Constago generation reference

Use this reference to run, verify, and troubleshoot Constago generation.

## Execution forms

Run an installed command:

```bash
constago --config constago.yaml
```

Run a released version without a global install:

```bash
go run github.com/cohesivestack/constago@v0.1.0 --config constago.yaml
```

Treat the version above as an example. Preserve the project's existing version
or select the requested released version instead of assuming `v0.1.0`.

Inspect available flags from the selected executable:

```bash
constago --help
```

Useful flags include:

```text
--config
--verbose
--input.dir
--input.include
--input.exclude
--input.struct.explicit
--input.struct.include_unexported
--input.struct.only
--input.struct.except
--input.field.explicit
--input.field.include_unexported
--input.field.only
--input.field.except
--output.file_name
```

Quote glob arguments so the shell does not expand them. Define elements and
getters in YAML rather than CLI flags.

Environment variables use the `CONSTAGO_` prefix and replace dots with
underscores, for example:

```bash
CONSTAGO_VERBOSE=2 \
CONSTAGO_OUTPUT_FILE_NAME=fields.gen.go \
constago --config constago.yaml
```

Prefer YAML or flags for predictable one-off behavior. Environment overrides
are most reliable for keys already represented in the YAML.

## Reproducible go:generate

Add a directive to a non-generated file:

```go
package model

//go:generate go run github.com/cohesivestack/constago@v0.1.0 --config ../constago.yaml
```

Pin the version that the project has selected. Paths are resolved from the
package directory in which `go generate` executes, not necessarily the module
root. Make `input.dir` agree with that working directory.

Run the repository workflow:

```bash
go generate ./...
gofmt -w .
go test ./...
git diff --exit-code
```

Use `gofmt -w .` only when that broad formatting scope is already the
repository's convention. During an agent change, prefer formatting just the
generated files to avoid unrelated rewrites.

## Safe verification

Before generation:

1. Capture `git status --short`.
2. Locate all existing files matching `output.file_name`.
3. Confirm generated filenames are excluded from the scan.
4. Confirm the command's working directory and config path.

After generation:

1. Compare `git status --short` with the pre-run state.
2. Identify every created or changed generated file; Constago can write one in
   each selected package.
3. Run `gofmt` on those files.
4. Inspect diffs for package names, imports, constants, grouped accessors, and
   getters.
5. Run focused package tests and `go test ./...` when practical.
6. In CI, regenerate, format, test, and fail on a non-empty generated diff.

Constago does not run `gofmt`, compile packages, or automatically exclude its
own output.

## Troubleshooting

### No file generated

Run with details:

```bash
constago --verbose 2 --config constago.yaml
```

Then check:

- `input.dir` relative to the process working directory;
- include patterns and package selectors;
- exclusions;
- struct and field explicit modes;
- exported and unexported selection;
- `only` and `except` regexes;
- tags required by `input.mode: tag`; and
- whether at least one selected field can produce an element or getter.

### Field or element missing

- `tag` mode omits the element when none of its priority tags exists.
- `tagThenField` falls back to the Go field name.
- `field` mode always derives from the Go field name.
- `field` in `tag_priority` means a literal tag named `field`.
- `json:"-"` is a literal value, not an automatic exclusion.

### Generated code fails to compile

Check:

- collisions with existing constants, variables, or methods;
- collisions between element or getter definitions;
- unexported names produced by `camel` or `snake`;
- types unavailable under current build tags;
- stale files under an older generated filename; and
- required imports for field types.

Never patch generated output as the durable fix. Change source or configuration,
regenerate, format, and test again.
