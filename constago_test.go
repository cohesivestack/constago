package main

import (
	"os"
	"path/filepath"
	"testing"

	constago "github.com/cohesivestack/constago/lib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRootCmd_ConfigFlagsApplied(t *testing.T) {
	// Build the command with a stub run that captures the config
	var captured *constago.Config
	cmd := newRootCmd(func(cfg *constago.Config) error {
		captured = cfg
		return nil
	})

	tmp := t.TempDir()

	args := []string{
		"--input.dir", tmp,
		"--input.include", "**/*.go",
		"--input.exclude", "**/*_test.go",
		"--input.struct.explicit", "true",
		"--input.struct.include_unexported", "true",
		"--input.field.explicit", "true",
		"--input.field.include_unexported", "true",
		"--output.file_name", "gen_out.go",
	}
	cmd.SetArgs(args)

	// Execute and ensure success
	require.NoError(t, cmd.Execute())
	require.NotNil(t, captured)

	// Validate the captured config reflects flags + defaults
	assert.Equal(t, tmp, captured.Input.Dir)
	assert.ElementsMatch(t, []string{"**/*.go"}, captured.Input.Include)
	assert.ElementsMatch(t, []string{"**/*_test.go"}, captured.Input.Exclude)

	// Struct flags
	if assert.NotNil(t, captured.Input.Struct.Explicit) {
		assert.True(t, *captured.Input.Struct.Explicit)
	}
	if assert.NotNil(t, captured.Input.Struct.IncludeUnexported) {
		assert.True(t, *captured.Input.Struct.IncludeUnexported)
	}

	// Field flags
	if assert.NotNil(t, captured.Input.Field.Explicit) {
		assert.True(t, *captured.Input.Field.Explicit)
	}
	if assert.NotNil(t, captured.Input.Field.IncludeUnexported) {
		assert.True(t, *captured.Input.Field.IncludeUnexported)
	}

	assert.Equal(t, "gen_out.go", captured.Output.FileName)
}

func TestCLI_EndToEnd_GeneratesOutput(t *testing.T) {
	tmp := t.TempDir()

	// Minimal Go source to scan
	goFile := filepath.Join(tmp, "user.go")
	src := `package main

type User struct {
    Name string ` + "`json:\"name\"`" + `
}`
	require.NoError(t, os.WriteFile(goFile, []byte(src), 0644))

	// Minimal YAML configuration enabling constants from json tag
	cfgFile := filepath.Join(tmp, "constago.yaml")
	yaml := `output:
  file_name: "out_gen.go"
input:
  dir: "` + tmp + `"
  include:
    - "**/*.go"
  exclude:
    - "**/*_test.go"
elements:
  - name: "json"
    input:
      mode: "tag"
      tag_priority:
        - "json"
    output:
      mode: "constant"
      format:
        holder: "pascal"
        struct: "pascal"
        prefix: "Json"
        suffix: ""
      transform:
        tag_values: false
        value_case: "asIs"
        value_separator: ""
`
	require.NoError(t, os.WriteFile(cfgFile, []byte(yaml), 0644))

	// Build command that runs the real generator
	cmd := newRootCmd(func(cfg *constago.Config) error {
		return constago.Generate(cfg)
	})

	cmd.SetArgs([]string{"--config", cfgFile})

	// Ensure template is resolved from repo root where code_template.tpl lives
	cwd, err := os.Getwd()
	require.NoError(t, err)
	repoRoot := filepath.Dir(cwd)
	require.NoError(t, os.Chdir(repoRoot))
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	require.NoError(t, cmd.Execute())

	// Verify file was generated in the same directory
	out := filepath.Join(tmp, "out_gen.go")
	assert.FileExists(t, out)

	// Spot-check expected content chunk
	data, err := os.ReadFile(out)
	require.NoError(t, err)
	expectedChunk := `
// Constants for User
const (
	JsonUserName = "name"
)`
	assert.Contains(t, string(data), expectedChunk)
}
