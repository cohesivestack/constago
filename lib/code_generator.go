package constago

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const templateName = "code_template.tpl"

//go:embed code_template.tpl
var codeTemplate string

type generator struct {
	model *Model
}

func Generate(config *Config) error {
	cfg, err := NewConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	// Build the model using the model builder
	builder := NewModelBuilder(cfg)
	model, err := builder.Build()
	if err != nil {
		return fmt.Errorf("failed to build model: %w", err)
	}

	g := &generator{model: model}

	// Parse the template
	tmpl, err := template.New(templateName).Parse(codeTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Generate code for each package
	for _, pkg := range g.model.Packages {
		if len(pkg.Structs) == 0 {
			continue // Skip packages with no structs to generate
		}

		// Create output directory if it doesn't exist
		outputDir := pkg.Path
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
		}

		fileName := filepath.Join(outputDir, cfg.Output.FileName)

		templateData := struct {
			Config  *Config
			Package *PackageModel
		}{
			Config:  cfg,
			Package: pkg,
		}

		output, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", fileName, err)
		}
		defer output.Close()

		err = tmpl.Execute(output, templateData)
		if err != nil {
			return fmt.Errorf("failed to execute template for %s: %w", fileName, err)
		}
	}

	return nil
}
