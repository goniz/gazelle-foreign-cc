package gazelle

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/bazelbuild/rules_go/go/tools/bazel"
)

// Helper function to create a mock language.GenerateArgs
func createMockGenerateArgs(t *testing.T, relDir string, files []string) language.GenerateArgs {
	// Find the workspace root.
	workspaceRoot, err := bazel.Runfile("") // Gets path to the current directory within the runfiles tree
	if err != nil {
		// Fallback for non-Bazel environments (go test)
		// Find the workspace root by looking for go.mod
		wd, err := filepath.Abs(".")
		if err != nil {
			t.Fatalf("Could not get working directory: %v", err)
		}
		// Walk up until we find go.mod
		for {
			if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
				workspaceRoot = wd
				break
			}
			parent := filepath.Dir(wd)
			if parent == wd {
				t.Fatalf("Could not find workspace root (go.mod not found)")
			}
			wd = parent
		}
	}
	// Construct the absolute path to the testdata directory
	absDir := filepath.Join(workspaceRoot, relDir)

	// Create a dummy config
	c := config.New()
	// Initialize CMakeConfig if your GenerateRules or other functions expect it
	// This ensures that GetCMakeConfig(c) doesn't panic or return nil if c.Exts["cmake"] is not set.
	// You might need to register it or manually set it depending on your GetCMakeConfig implementation.
	// Assuming GetCMakeConfig initializes if not present or you have a public constructor/initializer:
	_ = GetCMakeConfig(c) // Ensures c.Exts["cmake"] is populated.

	return language.GenerateArgs{
		Config:       c,
		Dir:          absDir, // Absolute path to the directory being processed
		Rel:          filepath.Base(relDir), // Relative path from repo root (or a common root for testdata)
		RegularFiles: files,
		File:         nil,        // Represents the existing BUILD file, nil if generating anew
	}
}

func TestGenerateRules_SimpleCCProject(t *testing.T) {
	// Define the structure of your test project within testdata
	// The path should be relative to where the test is run from,
	// or use bazel.Runfile to get absolute paths in Bazel test environment.
	// Assuming 'gazelle-foreign-cc' is the workspace name or not needed if paths are relative from it.
	projectRelDir := "testdata/simple_cc_project"

	args := createMockGenerateArgs(t,
		projectRelDir,
		[]string{"main.cc", "lib.cc", "lib.h", "CMakeLists.txt"}, // Files Gazelle sees in the directory
	)

	expectedRuleApp := rule.NewRule("cc_binary", "app")
	expectedRuleApp.SetAttr("srcs", []string{"main.cc"})
	// Our simple parser doesn't link lib.h to app directly, only sources from add_executable

	expectedRuleLib := rule.NewRule("cc_library", "my_lib")
	expectedRuleLib.SetAttr("srcs", []string{"lib.cc"})
	expectedRuleLib.SetAttr("hdrs", []string{"lib.h"})

	expectedRules := []*rule.Rule{expectedRuleApp, expectedRuleLib}

	// Call the function under test
	result := GenerateRules(args)

	// --- Verification ---
	if len(result.Gen) != len(expectedRules) {
		t.Errorf("Expected %d rules, got %d. Generated: %v", len(expectedRules), len(result.Gen), result.Gen)
		for _, r := range result.Gen {
			t.Logf("Generated rule: %s %s, srcs: %v, hdrs: %v", r.Kind(), r.Name(), r.AttrStrings("srcs"), r.AttrStrings("hdrs"))
		}
		return // Avoid further panics if lengths differ
	}

	// Sort both slices by rule name for consistent comparison
	sort.Slice(result.Gen, func(i, j int) bool { return result.Gen[i].Name() < result.Gen[j].Name() })
	sort.Slice(expectedRules, func(i, j int) bool { return expectedRules[i].Name() < expectedRules[j].Name() })

	for i, gotRule := range result.Gen {
		expectedRule := expectedRules[i]
		if gotRule.Kind() != expectedRule.Kind() || gotRule.Name() != expectedRule.Name() {
			t.Errorf("Rule %d: Expected %s %s, got %s %s",
				i, expectedRule.Kind(), expectedRule.Name(), gotRule.Kind(), gotRule.Name())
			continue
		}

		// Compare srcs
		gotSrcs := gotRule.AttrStrings("srcs")
		expectedSrcs := expectedRule.AttrStrings("srcs")
		sort.Strings(gotSrcs) // Sort for consistent comparison
		sort.Strings(expectedSrcs)
		if !reflect.DeepEqual(gotSrcs, expectedSrcs) {
			t.Errorf("Rule %s %s: Expected srcs %v, got %v",
				gotRule.Kind(), gotRule.Name(), expectedSrcs, gotSrcs)
		}

		// Compare hdrs
		gotHdrs := gotRule.AttrStrings("hdrs")
		expectedHdrs := expectedRule.AttrStrings("hdrs")
		sort.Strings(gotHdrs) // Sort for consistent comparison
		sort.Strings(expectedHdrs)
		if !reflect.DeepEqual(gotHdrs, expectedHdrs) {
			t.Errorf("Rule %s %s: Expected hdrs %v, got %v",
				gotRule.Kind(), gotRule.Name(), expectedHdrs, gotHdrs)
		}
		
		// TODO: Add more checks, e.g. for Empty rules if necessary
	}

	// Check Empty rules (important for Gazelle's update mechanism)
	if len(result.Empty) != len(expectedRules) {
		t.Errorf("Expected %d empty rules, got %d.", len(expectedRules), len(result.Empty))
	} else {
		sort.Slice(result.Empty, func(i, j int) bool { return result.Empty[i].Name() < result.Empty[j].Name() })
		for i, gotEmptyRule := range result.Empty {
			expectedRule := expectedRules[i] // Compare against the main expected rules
			if gotEmptyRule.Kind() != expectedRule.Kind() || gotEmptyRule.Name() != expectedRule.Name() {
				t.Errorf("Empty rule %d: Expected %s %s, got %s %s",
					i, expectedRule.Kind(), expectedRule.Name(), gotEmptyRule.Kind(), gotEmptyRule.Name())
			}
		}
	}
}

// TODO: Add more test cases:
// - CMakeLists.txt with no targets
// - CMakeLists.txt with targets but no matching source files in args.RegularFiles
// - CMakeLists.txt with only add_library or only add_executable
// - File names with unusual characters (if your regexes are too simple)
// - Case variations in add_library/add_executable (already handled by (?i) in regex)
