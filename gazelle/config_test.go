package gazelle

import (
	"testing"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

func TestCMakeSourceDirectiveKnown(t *testing.T) {
	cfg := NewCMakeConfig()
	directives := cfg.KnownDirectives()
	
	// Check that cmake_source directive is in known directives
	found := false
	for _, d := range directives {
		if d == "cmake_source" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("cmake_source directive not found in KnownDirectives")
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

func TestCMakeSourceDirective(t *testing.T) {
	cfg := NewCMakeConfig()
	c := &config.Config{
		Exts: make(map[string]interface{}),
	}
	c.Exts["cmake"] = cfg

	// Create a mock BUILD file with the cmake_source directive
	f := &rule.File{
		Directives: []rule.Directive{
			{Key: "cmake_source", Value: "@somelib"},
		},
	}

	// Configure should parse the directive
	cfg.Configure(c, "test/package", f)

	// The directive should be processed without error
	// Note: cmake_source directives are handled per-package, not stored in config
}

func TestCMakeDefineDirective(t *testing.T) {
	cfg := NewCMakeConfig()
	c := &config.Config{
		Exts: make(map[string]interface{}),
	}
	c.Exts["cmake"] = cfg

	// Create a mock BUILD file with cmake_define directives
	f := &rule.File{
		Directives: []rule.Directive{
			{Key: "cmake_define", Value: "ZMQ_BUILD_TESTS OFF"},
			{Key: "cmake_define", Value: "WITH_PERF_TOOL OFF"},
			{Key: "cmake_define", Value: "CMAKE_BUILD_TYPE Release"},
		},
	}

	// Configure should parse the directives without error
	cfg.Configure(c, "test/package", f)

	// Note: cmake_define directives are no longer stored globally in cfg.CMakeDefines
	// They are processed per-package in GenerateRules to ensure proper scoping.
	// This test verifies that Configure() processes the directives without error.
	
	// The CMakeDefines map should be empty since directives are processed per-package
	if len(cfg.CMakeDefines) != 0 {
		t.Errorf("Expected CMakeDefines to be empty (defines are processed per-package), got %v", cfg.CMakeDefines)
	}
}

func TestCMakeDefineDirectiveInvalidFormat(t *testing.T) {
	cfg := NewCMakeConfig()
	c := &config.Config{
		Exts: make(map[string]interface{}),
	}
	c.Exts["cmake"] = cfg

	// Create a mock BUILD file with invalid cmake_define directive
	f := &rule.File{
		Directives: []rule.Directive{
			{Key: "cmake_define", Value: "INVALID_FORMAT"},
			{Key: "cmake_define", Value: "TOO MANY PARTS HERE"},
		},
	}

	// Configure should handle invalid formats gracefully without panicking
	cfg.Configure(c, "test/package", f)

	// Note: cmake_define directives are no longer stored globally in cfg.CMakeDefines
	// They are processed per-package in GenerateRules where invalid formats are handled.
	// This test verifies that Configure() handles invalid directives without error.
	
	// The CMakeDefines map should be empty since directives are processed per-package
	if len(cfg.CMakeDefines) != 0 {
		t.Errorf("Expected CMakeDefines to be empty (defines are processed per-package), got %v", cfg.CMakeDefines)
	}
}

func TestCMakeDefineDirectiveKnown(t *testing.T) {
	cfg := NewCMakeConfig()
	directives := cfg.KnownDirectives()
	
	// Check that cmake_define directive is in known directives
	found := false
	for _, d := range directives {
		if d == "cmake_define" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("cmake_define directive not found in KnownDirectives")
	}
}