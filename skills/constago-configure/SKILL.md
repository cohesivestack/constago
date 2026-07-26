---
name: constago-configure
description: "Create, edit, review, and troubleshoot Constago YAML configuration, including constago.yaml, Go input selection, elements, getters, output names, value transforms, and exclusions. Use when setting up Constago for a Go project, changing which structs or fields are selected, choosing tag or field metadata, designing the generated API, or resolving configuration validation errors. Do not use merely to execute an existing configuration or to integrate the Go library programmatically."
---

# Configure Constago

Design a configuration from the Go source and the API the user wants to consume.
Keep configuration work separate from generated output and programmatic library
integration.

Read [references/configuration.md](references/configuration.md) before creating
or changing configuration. Treat it as the packaged configuration reference.
When working in the Constago repository itself, prefer the current files under
`docs/src/content/docs/` and `lib/config.go` if they differ from the packaged
reference.

## Workflow

1. Inspect the Go module before proposing YAML.
   - Find `go.mod`, existing `constago.yaml` files, `//go:generate` directives,
     generated filenames, packages, structs, field tags, and Constago include or
     exclude annotations.
   - Preserve an existing configuration's conventions and comments where
     possible.
2. Establish the intended generated API.
   - Determine which packages, structs, and fields belong in scope.
   - Determine whether consumers need package constants, grouped accessors,
     getters, or a combination.
   - Ask for direction only when different API shapes would materially affect
     downstream code and the repository does not reveal a preference.
3. Select inputs narrowly.
   - Set `input.dir` relative to the directory where Constago will run.
   - Add explicit include patterns and exclude tests, fixtures when appropriate,
     and the configured output filename.
   - Use regex filters or explicit annotations only when broad file selection is
     insufficient.
4. Design elements and getters.
   - Use `tag` when a field without the tag should produce no element.
   - Use `field` when values must always come from Go field names.
   - Use `tagThenField` when tag metadata is preferred but field-name fallback is
     acceptable.
   - Use `output.mode: none` for metadata needed only by getters.
   - Keep generated identifiers exported unless package-local access is
     intentional.
5. Create or update the YAML with the smallest explicit configuration that
   communicates intent. Rely on documented defaults only when they are safe and
   clear to future maintainers.
6. Review the result statically.
   - Check enum casing, Go identifier requirements, regexes, getter return names,
     `.go` glob suffixes, and output filename constraints.
   - Check for likely generated-name collisions.
   - Do not run Constago merely to validate configuration because generation
     writes files. If the user also requests generation, continue with the
     `constago-generate` workflow.

## Guardrails

- Never edit a generated `.go` file to compensate for configuration.
- Always exclude the configured generated filename; Constago does not exclude it
  automatically.
- Treat `json:"-"` and similar tag values as literal values. Exclude those
  fields explicitly when they should not generate metadata.
- Remember that `field` inside `tag_priority` means a literal tag named `field`;
  use `input.mode: field` or `tagThenField` for Go field-name values.
- Keep elements and getters in YAML. CLI flags cover only basic input and output
  overrides.
- Report the chosen selectors and generated API shape when handing off the
  configuration.
