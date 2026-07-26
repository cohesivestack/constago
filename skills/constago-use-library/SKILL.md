---
name: constago-use-library
description: "Integrate Constago programmatically in Go through github.com/cohesivestack/constago/lib. Use when importing the Constago package, loading YAML with LoadConfig, building Config values in Go, invoking Generate from custom tooling, applying defaults with NewConfig, inspecting scans with NewModelBuilder, or testing a library-based integration. Do not use for ordinary CLI generation or configuration-only work."
---

# Use Constago as a Go Library

Build programmatic Constago integrations with the public package whose import
path ends in `/lib` and whose declared package name is `constago`.

Read [references/library-api.md](references/library-api.md) before modifying Go
code. When working in the Constago repository itself, inspect the current
exported declarations and tests under `lib/` if they differ from the packaged
reference.

## Workflow

1. Inspect the consuming module.
   - Read `go.mod`, the relevant package, existing generation entry points,
     error-handling conventions, and tests.
   - Preserve an existing Constago version and import style.
2. Choose the smallest API surface for the requested outcome.
   - Use `LoadConfig` followed by `Generate` to drive generation from YAML.
   - Construct `Config` in Go followed by `Generate` when configuration must be
     compiled or assembled programmatically.
   - Use `NewConfig` and `NewModelBuilder(...).Build()` when callers need to
     validate defaults or inspect the scan model without writing generated
     files.
3. Add or update the dependency using the project's normal Go module workflow.
   - Do not silently upgrade an existing version.
   - Prefer a released, explicit version for reproducible tools.
4. Import the library explicitly when clarity helps:

   ```go
   import constago "github.com/cohesivestack/constago/lib"
   ```

5. Implement explicit error handling. Propagate or wrap errors in libraries and
   use the application's established logging or exit behavior in commands.
6. Verify side effects and paths.
   - `Generate` writes the configured filename into every selected package that
     produces output.
   - `input.dir` and relative YAML paths depend on the process working
     directory.
   - Exclude generated filenames from subsequent scans.
7. Format changed Go files, run focused tests, then run `go test ./...` when
   practical.

## API selection

- Prefer `LoadConfig` when configuration belongs to operators or a repository
  file.
- Prefer a Go `Config` when configuration is derived, embedded, or needs compiler
  checks.
- Prefer `NewModelBuilder` when the task is analysis, preview, reporting, or
  custom rendering rather than writing Constago's standard generated output.
- Use exported enum constants instead of raw strings in Go configuration.

## Guardrails

- Do not import the repository root command package; import `/lib`.
- Do not pass `nil` to `NewConfig`.
- Remember that `NewConfig` mutates the supplied `*Config` while applying
  defaults.
- Do not manually reproduce Constago's defaulting logic in consumer code.
- Do not use `Generate` when the user requests a side-effect-free inspection;
  build the model instead.
- Never edit generated output as the primary implementation.
