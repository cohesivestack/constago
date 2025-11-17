package main

import (
	"errors"
	"fmt"
	"strings"

	constago "github.com/cohesivestack/constago/lib"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	if err := newRootCmd(func(cfg *constago.Config) error {
		err := constago.Generate(cfg)
		if err != nil {
			return err
		}
		return nil
	}).Execute(); err != nil {
		panic(err)
	}
}

// loadConfigFromViper builds a Config using the given viper instance,
// applies the defaults and validations via NewConfig, and returns it.
// It respects the `yaml` struct tags when unmarshalling.
func loadConfigFromViper(v *viper.Viper) (*constago.Config, error) {
	raw := &constago.Config{}

	// Use yaml tags to decode into the structs.
	if err := v.Unmarshal(raw, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "yaml"
	}); err != nil {
		return nil, fmt.Errorf("unable to decode configuration: %w", err)
	}

	// Pass through the normal constructor (defaults + validate)
	cfg, err := constago.NewConfig(raw)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// initViper sets up Viper with config file (optional), env prefix + replacer, and binds all flags.
func initViper(cmd *cobra.Command) (*viper.Viper, error) {
	v := viper.New()

	// ----- Config file (optional) -----
	cfgFile, _ := cmd.Flags().GetString("config")
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file %q: %w", cfgFile, err)
		}
	} else {
		v.SetConfigName("constago")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		if err := v.ReadInConfig(); err != nil {
			var nf viper.ConfigFileNotFoundError
			if !errors.As(err, &nf) {
				return nil, fmt.Errorf("failed to read config: %w", err)
			}
			// If not found, that's fine — flags/env may provide everything.
		}
	}

	// ----- ENV: CONSTAGO_<DOTS_AS_UNDERSCORES> -----
	v.SetEnvPrefix("CONSTAGO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	return v, nil
}

// ---- Only apply CHANGED flags to Viper (preserves tri-state) ----
func applyChangedFlagsToViper(cmd *cobra.Command, v *viper.Viper) error {
	get := func(name string) (any, error) {
		// Decide type by flag.Value.Type()
		f := cmd.Flags().Lookup(name)
		if f == nil {
			return nil, fmt.Errorf("flag %q not found", name)
		}
		switch f.Value.Type() {
		case "bool":
			return cmd.Flags().GetBool(name)
		case "string":
			return cmd.Flags().GetString(name)
		case "stringSlice":
			return cmd.Flags().GetStringSlice(name)
		case "int":
			return cmd.Flags().GetInt(name)
		case "intSlice":
			return cmd.Flags().GetIntSlice(name)
		case "float64":
			return cmd.Flags().GetFloat64(name)
		case "duration":
			return cmd.Flags().GetDuration(name)
		default:
			// For unknown types, try to get as string
			return f.Value.String(), nil
		}
	}

	// Visit only flags that the user actually set on the CLI
	cmd.Flags().Visit(func(f *pflag.Flag) {
		val, err := get(f.Name)
		if err == nil {
			// Only set non-empty values to avoid validation issues
			switch valType := val.(type) {
			case string:
				if valType != "" {
					v.Set(f.Name, val)
				}
			case []string:
				if len(valType) > 0 {
					v.Set(f.Name, val)
				}
			case bool:
				// Always set bool flags since false is a valid value
				v.Set(f.Name, val)
			default:
				v.Set(f.Name, val)
			}
		}
	})
	return nil
}

// newRootCmd creates the Cobra CLI, wires Viper, merges sources, and runs a callback.
func newRootCmd(run func(*constago.Config) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "constago",
		Short: "Generate constants and getters from project structs/tags",
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := initViper(cmd)
			if err != nil {
				return err
			}

			// Important: apply *only* flags the user passed
			if err := applyChangedFlagsToViper(cmd, v); err != nil {
				return err
			}

			cfg, err := loadConfigFromViper(v)
			if err != nil {
				return err
			}
			if run == nil {
				return nil
			}
			return run(cfg)
		},
	}

	// Global
	cmd.Flags().String("config", "", "Path to YAML config file")

	// ---------- INPUT ----------
	cmd.Flags().String("input.dir", "", "Directory to scan (e.g., ./)")
	cmd.Flags().StringSlice("input.include", nil, "Glob patterns to include (comma-separated for ENV)")
	cmd.Flags().StringSlice("input.exclude", nil, "Glob patterns to exclude (comma-separated for ENV)")

	cmd.Flags().Bool("input.struct.explicit", false, "Only include structs explicitly marked")
	cmd.Flags().Bool("input.struct.include_unexported", false, "Include unexported structs when scanning")

	cmd.Flags().String("input.struct.include_only", "", "Regular expression to include structs (whitelist)")
	cmd.Flags().String("input.struct.include_except", "", "Regular expression to exclude structs (blacklist)")

	cmd.Flags().Bool("input.field.explicit", false, "Only include fields explicitly marked")
	cmd.Flags().Bool("input.field.include_unexported", false, "Include unexported fields when scanning")

	cmd.Flags().String("input.field.include_only", "", "Regular expression to include fields (whitelist)")
	cmd.Flags().String("input.field.include_except", "", "Regular expression to exclude fields (blacklist)")

	// ---------- OUTPUT ----------
	cmd.Flags().String("output.file_name", "", "Output file name (e.g., constants_gen.go)")

	// Add help text for simplified configuration
	cmd.Long = `Constago generates constants and getter functions from Go structs.

The tool supports configuration via:
- YAML config file (recommended for all setups)
- Command line flags (for basic input/output overrides)
- Environment variables (CONSTAGO_* prefix)

Elements and getters configuration must be done via YAML config file.
CLI flags only support basic input and output parameters.

Examples:
  constago --config constago.yaml
  constago --input.dir ./src --output.file_name constants.go
  constago --input.include "**/*.go" --input.exclude "**/*_test.go"`

	return cmd
}
