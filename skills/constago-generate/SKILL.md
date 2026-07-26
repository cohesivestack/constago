---
name: constago-generate
description: "Run and verify Constago code generation from an existing configuration using the Constago CLI, go run, or go generate. Use when generating or regenerating Go source, setting up a reproducible go:generate or CI workflow, formatting and testing generated output, inspecting generation diffs, or troubleshooting missing and invalid generated code. Do not use primarily to design constago.yaml or to embed Constago as a Go library."
---

# Generate with Constago

Execute an existing Constago setup reproducibly and verify the generated Go
source without hand-editing it.

Read [references/generation.md](references/generation.md) before running or
changing a generation workflow. When working in the Constago repository itself,
prefer the current CLI implementation and documentation if they differ from the
packaged reference.

## Workflow

1. Inspect the project workflow.
   - Read `go.mod`, `constago.yaml`, existing `//go:generate` directives,
     build scripts, CI configuration, generated files, and repository guidance.
   - Record the working tree state so generated changes can be distinguished
     from the user's existing changes.
2. Choose the existing execution path.
   - Prefer the repository's documented command or `go generate` directive.
   - Preserve an already pinned Constago version.
   - Do not silently install a global binary or upgrade Constago.
   - If no workflow exists, prefer `go run
     github.com/cohesivestack/constago@<version>` for a reproducible run and use
     an explicit released version when committing a directive.
3. Preflight the configuration.
   - Resolve the configuration path and `input.dir` from the command's working
     directory.
   - Confirm that `input.exclude` covers the configured output filename.
   - If configuration design must change materially, apply the
     `constago-configure` workflow before generation.
4. Run the narrowest appropriate command from the correct directory.
   - Capture the full failure output and exit status.
   - Do not work around a failure by modifying generated files.
5. Identify files created or changed by this run.
   - Constago can write the configured filename into more than one selected
     package.
   - Format only generated files unless the repository explicitly formats a
     broader scope.
6. Verify in proportion to the change.
   - Inspect generated diffs for the expected structs, elements, getters, and
     imports.
   - Run focused package tests, followed by `go test ./...` when practical.
   - Run the repository's normal lint, build, or generated-code drift check when
     one exists.
7. Report the command, generated paths, verification results, and any remaining
   warnings.

## Troubleshooting order

1. Confirm the process working directory and config path.
2. Run with verbosity level `2`.
3. Confirm include patterns match `.go` files and excludes do not remove the
   intended package.
4. Check explicit struct and field selection, regex filters, tags, and
   annotations.
5. Check for old generated filenames and identifier or method collisions.
6. Format and compile the generated packages to distinguish generator failures
   from Go compilation failures.

## Guardrails

- Preserve unrelated working-tree changes.
- Treat generated files as outputs: change source or configuration, then
  regenerate.
- Ask before introducing a materially different versioning or CI policy.
- Do not assume Constago runs `gofmt` or compiles the result.
- Do not claim success until the generated diff has been inspected and relevant
  Go tests have passed, or clearly state why verification could not run.
