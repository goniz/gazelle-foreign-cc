package gazelle

import (
	"flag"
	"log"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

// CMakeConfig holds configuration settings for the CMake Gazelle plugin.
type CMakeConfig struct {
	// Example configuration field: path to CMake executable.
	CMakeExecutable string
	// CMake variables to be passed as -D flags
	CMakeDefines map[string]string
	// Add other CMake-specific configuration fields here.
}

// Constants for directive names
const (
	cmakeExecutableDirective = "cmake_executable"
	cmakeSourceDirective    = "cmake_source"
	cmakeDefineDirective    = "cmake_define"
	// Define other directive names here
)

// NewCMakeConfig creates a new CMakeConfig with default values.
func NewCMakeConfig() *CMakeConfig {
	return &CMakeConfig{
		CMakeExecutable: "cmake", // Default value
		CMakeDefines:    make(map[string]string),
	}
}

// RegisterFlags registers command-line flags for CMake configuration.
// It satisfies the config.Configurer interface.
func (cfg *CMakeConfig) RegisterFlags(fs *flag.FlagSet, cmd string, c *config.Config) {
	// Example of registering a flag.
	// fs.StringVar(&cfg.SomeSetting, "cmake_some_setting", "default_value", "Description of some setting")
}

// CheckFlags validates the configuration settings.
// It satisfies the config.Configurer interface.
func (cfg *CMakeConfig) CheckFlags(fs *flag.FlagSet, c *config.Config) error {
	// Example of validating a flag.
	// if cfg.SomeSetting == "" {
	// 	return fmt.Errorf("-cmake_some_setting must be set")
	// }
	return nil
}

// KnownDirectives returns a list of directive keys that this configurer understands.
// It satisfies the config.Configurer interface.
func (cfg *CMakeConfig) KnownDirectives() []string {
	return []string{
		cmakeExecutableDirective,
		cmakeSourceDirective,
		cmakeDefineDirective,
		// Add other known directives here
	}
}

// Configure parses directives from a BUILD file and updates the configuration.
// It satisfies the config.Configurer interface.
// This function is called for each BUILD file Gazelle processes.
func (cfg *CMakeConfig) Configure(c *config.Config, rel string, f *rule.File) {
	if f == nil { // Only process directives from BUILD files
		return
	}

	for _, directive := range f.Directives {
		switch directive.Key {
		case cmakeExecutableDirective:
			cfg.CMakeExecutable = directive.Value
			log.Printf("Configure: Set CMake executable to %s from directive in %s", cfg.CMakeExecutable, rel)
		case cmakeSourceDirective:
			// The cmake_source directive is handled per-package in GenerateRules, not globally
			log.Printf("Configure: Found cmake_source directive %s in %s (will be processed per-package)", directive.Value, rel)
		case cmakeDefineDirective:
			// Parse cmake_define directive in format "KEY VALUE"
			parts := strings.Fields(directive.Value)
			if len(parts) != 2 {
				log.Printf("Configure: Invalid cmake_define directive format '%s' in %s. Expected format: 'KEY VALUE'", directive.Value, rel)
				continue
			}
			key, value := parts[0], parts[1]
			cfg.CMakeDefines[key] = value
			log.Printf("Configure: Set CMake define %s=%s from directive in %s", key, value, rel)
		// Add cases for other directives here
		default:
			// Gazelle will warn about unknown directives if not in KnownDirectives()
		}
	}
}

// GetCMakeConfig retrieves the CMakeConfig from the global config.Config.
// It initializes it if it doesn't exist.
func GetCMakeConfig(c *config.Config) *CMakeConfig {
	if cfg, ok := c.Exts["cmake"].(*CMakeConfig); ok {
		return cfg
	}
	// If not found, create a new one and store it.
	newCfg := NewCMakeConfig()
	c.Exts["cmake"] = newCfg
	return newCfg
}

// Note: In modern Gazelle, configuration is handled through the Language interface
// methods rather than a separate Configurer registration.
