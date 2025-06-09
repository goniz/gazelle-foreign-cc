package gazelle

import (
	"log"
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
		Rel:          relDir, // Relative path from repo root
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

	// Check Empty rules - we don't generate empty rules currently since they interfere with deps generation
	// This may change in the future if we need empty rules for Gazelle's update mechanism
	if len(result.Empty) != 0 {
		t.Errorf("Expected 0 empty rules, got %d.", len(result.Empty))
	}
}

func TestGenerateRules_DepsGeneration(t *testing.T) {
	// Test that target_link_libraries generates correct deps attributes
	projectRelDir := "testdata/simple_cc_project"

	args := createMockGenerateArgs(t,
		projectRelDir,
		[]string{"main.cc", "lib.cc", "lib.h", "CMakeLists.txt"},
	)

	result := GenerateRules(args)

	// Find the "app" rule
	var appRule *rule.Rule
	for _, r := range result.Gen {
		if r.Name() == "app" {
			appRule = r
			break
		}
	}

	if appRule == nil {
		t.Fatal("Expected to find 'app' rule")
	}

	// Verify that the app rule has deps attribute set to the library
	deps := appRule.AttrStrings("deps")
	expectedDeps := []string{":my_lib"}
	if !reflect.DeepEqual(deps, expectedDeps) {
		t.Errorf("Expected deps %v for rule 'app', got %v", expectedDeps, deps)
	}
}

func TestGenerateRules_ComplexCCProject_DepsGeneration(t *testing.T) {
	// Test that complex target_link_libraries generates correct deps attributes
	projectRelDir := "testdata/complex_cc_project"

	args := createMockGenerateArgs(t,
		projectRelDir,
		[]string{"src/main.cpp", "src/core.cpp", "src/manager.cpp", "src/utils.cpp", "src/helper.cpp", "tests/test_main.cpp", "tests/test_utils.cpp", "CMakeLists.txt"},
	)

	result := GenerateRules(args)

	// Create a map of rules by name for easier testing
	rulesByName := make(map[string]*rule.Rule)
	for _, r := range result.Gen {
		rulesByName[r.Name()] = r
	}

	// Test core library dependencies
	coreRule := rulesByName["core"]
	if coreRule == nil {
		t.Fatal("Expected to find 'core' rule")
	}
	coreDeps := coreRule.AttrStrings("deps")
	expectedCoreDeps := []string{":utils"}
	if !reflect.DeepEqual(coreDeps, expectedCoreDeps) {
		t.Errorf("Expected deps %v for rule 'core', got %v", expectedCoreDeps, coreDeps)
	}

	// Test main_app dependencies (should link to both core and utils)
	mainAppRule := rulesByName["main_app"]
	if mainAppRule == nil {
		t.Fatal("Expected to find 'main_app' rule")
	}
	mainAppDeps := mainAppRule.AttrStrings("deps")
	expectedMainAppDeps := []string{":core", ":utils"}
	// Sort both to ensure consistent comparison
	sort.Strings(mainAppDeps)
	sort.Strings(expectedMainAppDeps)
	if !reflect.DeepEqual(mainAppDeps, expectedMainAppDeps) {
		t.Errorf("Expected deps %v for rule 'main_app', got %v", expectedMainAppDeps, mainAppDeps)
	}

	// Test test_runner dependencies 
	testRunnerRule := rulesByName["test_runner"]
	if testRunnerRule == nil {
		t.Fatal("Expected to find 'test_runner' rule")
	}
	testRunnerDeps := testRunnerRule.AttrStrings("deps")
	expectedTestRunnerDeps := []string{":utils"}
	if !reflect.DeepEqual(testRunnerDeps, expectedTestRunnerDeps) {
		t.Errorf("Expected deps %v for rule 'test_runner', got %v", expectedTestRunnerDeps, testRunnerDeps)
	}
}

// TODO: Add more test cases:
// - CMakeLists.txt with no targets
// - CMakeLists.txt with targets but no matching source files in args.RegularFiles
// - CMakeLists.txt with only add_library or only add_executable
// - File names with unusual characters (if your regexes are too simple)
// - Case variations in add_library/add_executable (already handled by (?i) in regex)

func TestGenerateRules_ConfigureFile(t *testing.T) {
	// Test the configure_file parsing functionality
	relDir := "testdata/configure_file_example"
	
	// Files in the test directory
	files := []string{
		"CMakeLists.txt",
		"config.h.in",
		"src/lib.cpp",
		"src/main.cpp",
	}
	
	args := createMockGenerateArgs(t, relDir, files)
	
	result := GenerateRules(args)
	
	// Verify that rules were generated
	if len(result.Gen) == 0 {
		t.Fatal("Expected rules to be generated, but got none")
	}
	
	// Create a map for easier lookup
	rulesByName := make(map[string]*rule.Rule)
	for _, r := range result.Gen {
		rulesByName[r.Name()] = r
	}
	
	// Check that cmake_configure_file rule was generated
	configRule := rulesByName["config_h"]
	if configRule == nil {
		t.Fatal("Expected to find 'config_h' cmake_configure_file rule")
	}
	
	if configRule.Kind() != "cmake_configure_file" {
		t.Errorf("Expected rule kind 'cmake_configure_file', got '%s'", configRule.Kind())
	}
	
	// Check attributes
	out := configRule.AttrString("out")
	if out != "config.h" {
		t.Errorf("Expected out 'config.h', got '%s'", out)
	}
	
	// Check new attributes
	cmakeBinary := configRule.AttrString("cmake_binary")
	if cmakeBinary != "//:cmake" {
		t.Errorf("Expected cmake_binary '//:cmake', got '%s'", cmakeBinary)
	}
	
	cmakeSourceDir := configRule.AttrString("cmake_source_dir")
	if cmakeSourceDir != "." {
		t.Errorf("Expected cmake_source_dir '.', got '%s'", cmakeSourceDir)
	}
	
	generatedFilePath := configRule.AttrString("generated_file_path")
	// generated_file_path should be optional when it equals the out attribute
	// If not set explicitly, the rule should default it to the out path
	if generatedFilePath != "" && generatedFilePath != "config.h" {
		t.Errorf("Expected generated_file_path to be empty or 'config.h', got '%s'", generatedFilePath)
	}
	
	// Check that cmake_source_files includes CMakeLists.txt and the input file
	cmakeSourceFiles := configRule.AttrStrings("cmake_source_files")
	expectedFiles := []string{"CMakeLists.txt", "config.h.in"}
	if len(cmakeSourceFiles) != len(expectedFiles) {
		t.Errorf("Expected cmake_source_files %v, got %v", expectedFiles, cmakeSourceFiles)
	} else {
		for i, expected := range expectedFiles {
			if i >= len(cmakeSourceFiles) || cmakeSourceFiles[i] != expected {
				t.Errorf("Expected cmake_source_files[%d] '%s', got '%s'", i, expected, cmakeSourceFiles[i])
				break
			}
		}
	}
	
	// Check that defines were set (stored as a private attribute)
	if !hasDefines(configRule) {
		t.Error("Expected defines to be set, but rule does not have defines attribute")
	}
	
	// Check that regular cc_library and cc_binary rules were also generated
	libRule := rulesByName["mylib"]
	if libRule == nil {
		t.Error("Expected to find 'mylib' cc_library rule")
	} else if libRule.Kind() != "cc_library" {
		t.Errorf("Expected mylib to be cc_library, got %s", libRule.Kind())
	}
	
	appRule := rulesByName["app"]
	if appRule == nil {
		t.Error("Expected to find 'app' cc_binary rule")
	} else if appRule.Kind() != "cc_binary" {
		t.Errorf("Expected app to be cc_binary, got %s", appRule.Kind())
	}
	
	log.Printf("ConfigureFile test completed successfully. Generated %d rules.", len(result.Gen))
}

// Helper function to check if a rule has defines attribute set
func hasDefines(r *rule.Rule) bool {
	return r.Attr("defines") != nil
}
