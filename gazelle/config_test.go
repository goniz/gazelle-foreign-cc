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

	// Configure should parse the directives
	cfg.Configure(c, "test/package", f)

	// Verify the directives were parsed correctly
	expectedDefines := map[string]string{
		"ZMQ_BUILD_TESTS":   "OFF",
		"WITH_PERF_TOOL":    "OFF",
		"CMAKE_BUILD_TYPE":  "Release",
	}

	for key, expectedValue := range expectedDefines {
		if actualValue, exists := cfg.CMakeDefines[key]; !exists {
			t.Errorf("Expected CMake define %s to be set", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected CMake define %s=%s, got %s=%s", key, expectedValue, key, actualValue)
		}
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

	// Configure should handle invalid formats gracefully
	cfg.Configure(c, "test/package", f)

	// No defines should be set for invalid formats
	if len(cfg.CMakeDefines) != 0 {
		t.Errorf("Expected no CMake defines to be set for invalid formats, got %v", cfg.CMakeDefines)
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