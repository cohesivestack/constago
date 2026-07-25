package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	constago "github.com/cohesivestack/constago/lib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDocsGettingStarted verifies the source and configuration published on the
// documentation home page with the real CLI command.
func TestDocsGettingStarted(t *testing.T) {
	tmp := t.TempDir()

	source := `package model

type User struct {
	Name    string ` + "`json:\"name\" title:\"Name\"`" + `
	Country string ` + "`json:\"country\" title:\"Country\"`" + `
}
`
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "user.go"), []byte(source), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/model\n\ngo 1.23\n"), 0o644))

	config := `input:
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
`
	configPath := filepath.Join(tmp, "constago.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(config), 0o644))

	previousDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(previousDir))
	})

	cmd := newRootCmd(func(cfg *constago.Config) error {
		return constago.Generate(cfg)
	})
	cmd.SetArgs([]string{"--config", configPath, "--verbose", "0"})
	require.NoError(t, cmd.Execute())

	repeatedCmd := newRootCmd(func(cfg *constago.Config) error {
		return constago.Generate(cfg)
	})
	repeatedCmd.SetArgs([]string{"--config", configPath, "--verbose", "0"})
	require.NoError(t, repeatedCmd.Execute())

	generatedPath := filepath.Join(tmp, "constago.gen.go")
	gofmt := exec.Command("gofmt", "-w", generatedPath)
	require.NoError(t, gofmt.Run())

	generated, err := os.ReadFile(generatedPath)
	require.NoError(t, err)
	output := string(generated)
	assert.Contains(t, output, `JsonUserName    = "name"`)
	assert.Contains(t, output, `JsonUserCountry = "country"`)
	assert.Contains(t, output, `var TitleUser = struct {`)
	assert.Contains(t, output, `func (_struct *User) MetadataName() (string, string, string)`)
	assert.Contains(t, output, `return "country", "Country", _struct.Country`)

	goTest := exec.Command("go", "test", "./...")
	goTest.Dir = tmp
	result, err := goTest.CombinedOutput()
	require.NoError(t, err, strings.TrimSpace(string(result)))
}

// TestDocsLibraryExample verifies the programmatic configuration documented on
// the Library API page and compiles the generated package.
func TestDocsLibraryExample(t *testing.T) {
	tmp := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(tmp, "product.go"),
		[]byte("package product\n\ntype Product struct {\n\tSKU string `json:\"sku\"`\n}\n"),
		0o644,
	))
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/product\n\ngo 1.23\n"), 0o644))

	cfg := &constago.Config{
		Verbose: intPointer(0),
		Input: constago.ConfigInput{
			Dir:     tmp,
			Include: []string{"**/*.go"},
			Exclude: []string{"**/*_test.go", "**/constago.gen.go"},
		},
		Output: constago.ConfigOutput{
			FileName: "constago.gen.go",
		},
		Elements: []constago.ConfigTag{
			{
				Name: "json",
				Input: constago.ConfigTagInput{
					Mode:        constago.InputModeTypeTag,
					TagPriority: []string{"json"},
				},
				Output: constago.ConfigTagOutput{
					Mode: constago.OutputModeConstant,
				},
			},
		},
	}

	require.NoError(t, constago.Generate(cfg))
	require.NoError(t, constago.Generate(cfg))

	generatedPath := filepath.Join(tmp, "constago.gen.go")
	gofmt := exec.Command("gofmt", "-w", generatedPath)
	require.NoError(t, gofmt.Run())

	generated, err := os.ReadFile(generatedPath)
	require.NoError(t, err)
	assert.Contains(t, string(generated), `JsonProductSku = "sku"`)

	goTest := exec.Command("go", "test", "./...")
	goTest.Dir = tmp
	result, err := goTest.CombinedOutput()
	require.NoError(t, err, strings.TrimSpace(string(result)))
}

// TestDocsCLIEnvironmentOverride verifies the documented environment-variable
// form against a key that is present in the YAML configuration.
func TestDocsCLIEnvironmentOverride(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "constago.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(`output:
  file_name: "yaml.gen.go"
`), 0o644))
	t.Setenv("CONSTAGO_OUTPUT_FILE_NAME", "environment.gen.go")

	var captured *constago.Config
	cmd := newRootCmd(func(cfg *constago.Config) error {
		captured = cfg
		return nil
	})
	cmd.SetArgs([]string{"--config", configPath})

	require.NoError(t, cmd.Execute())
	require.NotNil(t, captured)
	assert.Equal(t, "environment.gen.go", captured.Output.FileName)
}

func intPointer(value int) *int {
	return &value
}
