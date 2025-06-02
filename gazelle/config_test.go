package gazelle

import (
	"testing"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

func TestCMakeDirectiveKnown(t *testing.T) {
	cfg := NewCMakeConfig()
	directives := cfg.KnownDirectives()
	
	// Check that cmake directive is in known directives
	found := false
	for _, d := range directives {
		if d == "cmake" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("cmake directive not found in KnownDirectives")
	}
}

func TestCMakeExecutableDirective(t *testing.T) {
	cfg := NewCMakeConfig()
	c := &config.Config{
		Exts: make(map[string]interface{}),
	}
	c.Exts["cmake"] = cfg

	// Create a mock BUILD file with the cmake_executable directive
	f := &rule.File{
		Directives: []rule.Directive{
			{Key: "cmake_executable", Value: "/usr/bin/cmake"},
		},
	}

	// Configure should parse the directive
	cfg.Configure(c, "test/package", f)

	// Verify the directive was parsed correctly
	if cfg.CMakeExecutable != "/usr/bin/cmake" {
		t.Errorf("Expected CMakeExecutable to be '/usr/bin/cmake', got '%s'", cfg.CMakeExecutable)
	}
}