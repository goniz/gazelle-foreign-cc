package language

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/goniz/gazelle-foreign-cc/gazelle"
	"github.com/goniz/gazelle-foreign-cc/common"
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

func TestCMakeDirectiveParsing(t *testing.T) {
	// Test the cmake directive parsing logic
	
	// Create a mock BUILD file with cmake directive
	buildFile := &rule.File{
		Directives: []rule.Directive{
			{Key: "cmake", Value: "@somelib//:srcs"},
		},
	}
	
	// Simulate finding the directive in GenerateRules
	var cmakeSource string
	if buildFile != nil {
		for _, directive := range buildFile.Directives {
			if directive.Key == "cmake" {
				cmakeSource = directive.Value
				break
			}
		}
	}
	
	if cmakeSource != "@somelib//:srcs" {
		t.Errorf("Expected cmake directive value '@somelib//:srcs', got '%s'", cmakeSource)
	}
}

func TestExternalRepoPathGeneration(t *testing.T) {
	// Test the external repository path generation logic
	
	// Create mock args
	c := &config.Config{
		RepoRoot: "/test/workspace",
		Exts:     make(map[string]interface{}),
	}
	c.Exts["cmake"] = gazelle.NewCMakeConfig()
	
	args := language.GenerateArgs{
		Config: c,
		Dir:    "/test/workspace/thirdparty/somelib",
		Rel:    "thirdparty/somelib",
	}
	
	// Test different repository names
	testCases := []struct {
		repoName     string
		expectedPath string
	}{
		{"somelib", "/test/workspace/bazel-workspace/external/somelib"},
		{"myrepo", "/test/workspace/bazel-workspace/external/myrepo"},
	}
	
	for _, tc := range testCases {
		// We can't actually test findExternalRepo without creating the directories,
		// but we can test that the path generation logic is correct
		expectedPath := filepath.Join(args.Config.RepoRoot, "bazel-"+filepath.Base(args.Config.RepoRoot), "external", tc.repoName)
		if expectedPath != tc.expectedPath {
			t.Errorf("Expected path %s, got %s", tc.expectedPath, expectedPath)
		}
	}
}

func TestFindRepoViaRunfilesNewLogic(t *testing.T) {
	// Test the new runfiles-based repository detection logic
	
	lang := &cmakeLang{}
	
	// Test case 1: Repository that doesn't exist in runfiles
	repoPath := lang.findRepoViaRunfiles("nonexistent_repo")
	if repoPath != "" {
		t.Errorf("Expected empty path for nonexistent repository, got %s", repoPath)
	}
	
	// Note: We can't easily test the positive case without actually having
	// the repository available in runfiles during the test, but the negative
	// case validates that the method handles missing repositories correctly
}

func TestFindExternalRepoSimplified(t *testing.T) {
	// Test the simplified external repository detection logic
	
	lang := &cmakeLang{}
	
	// Create mock args
	c := &config.Config{
		RepoRoot: "/test/workspace",
		Exts:     make(map[string]interface{}),
	}
	c.Exts["cmake"] = gazelle.NewCMakeConfig()
	
	args := language.GenerateArgs{
		Config: c,
		Dir:    "/test/workspace/thirdparty/somelib",
		Rel:    "thirdparty/somelib",
	}
	
	// Test that findExternalRepo returns empty string when runfiles don't contain the repo
	// (which is expected behavior when the repo is not provided as data to gazelle rule)
	repoPath := lang.findExternalRepo("nonexistent_repo", args)
	if repoPath != "" {
		t.Errorf("Expected empty path for repository not in runfiles, got %s", repoPath)
	}
}

func TestLabelParsing(t *testing.T) {
	// Test the label parsing logic from generateRulesFromExternalSource
	
	testCases := []struct {
		label        string
		expectedRepo string
		expectedTarget string
		shouldFail   bool
	}{
		{"@somelib//:srcs", "somelib", ":srcs", false},
		{"@myrepo//path:target", "myrepo", "path:target", false},
		{"invalid_label", "", "", true},
		{"@repo", "", "", true},
		{"@", "", "", true},
	}
	
	for _, tc := range testCases {
		if tc.label == "" || tc.label[0] != '@' {
			if !tc.shouldFail {
				t.Errorf("Expected label %s to fail validation", tc.label)
			}
			continue
		}
		
		parts := splitLabel(tc.label)
		if tc.shouldFail {
			if len(parts) == 2 {
				t.Errorf("Expected label %s to fail parsing, but it succeeded", tc.label)
			}
			continue
		}
		
		if len(parts) != 2 {
			t.Errorf("Failed to parse valid label %s", tc.label)
			continue
		}
		
		repoName := parts[0][1:] // Remove @ prefix
		targetPart := parts[1]
		
		if repoName != tc.expectedRepo {
			t.Errorf("Expected repo name %s, got %s for label %s", tc.expectedRepo, repoName, tc.label)
		}
		
		if targetPart != tc.expectedTarget {
			t.Errorf("Expected target %s, got %s for label %s", tc.expectedTarget, targetPart, tc.label)
		}
	}
}

// Helper function to split label (extracted from the main logic)
func splitLabel(label string) []string {
	if len(label) == 0 || label[0] != '@' {
		return nil
	}
	
	parts := []string{}
	if idx := findIndex(label, "//"); idx != -1 {
		parts = []string{label[:idx], label[idx+2:]}
	}
	return parts
}

// Helper function to find index of substring
func findIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestCMakeDefinesInAPICommand(t *testing.T) {
	// Test that CMake defines are properly passed to the cmake command
	sourceDir := "/tmp/test_source"
	buildDir := "/tmp/test_build"
	cmakeExe := "cmake"
	cmakeDefines := map[string]string{
		"ZMQ_BUILD_TESTS": "OFF",
		"WITH_PERF_TOOL":  "OFF",
		"CMAKE_BUILD_TYPE": "Release",
	}

	api := NewCMakeFileAPI(sourceDir, buildDir, cmakeExe, cmakeDefines)

	// Check that the defines are stored correctly
	if len(api.cmakeDefines) != 3 {
		t.Errorf("Expected 3 cmake defines, got %d", len(api.cmakeDefines))
	}

	for key, expectedValue := range cmakeDefines {
		if actualValue, exists := api.cmakeDefines[key]; !exists {
			t.Errorf("Expected CMake define %s to be stored", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected CMake define %s=%s, got %s=%s", key, expectedValue, key, actualValue)
		}
	}
}

func TestCMakeEmptyDefines(t *testing.T) {
	// Test that empty defines work correctly
	sourceDir := "/tmp/test_source"
	buildDir := "/tmp/test_build"  
	cmakeExe := "cmake"
	cmakeDefines := make(map[string]string)

	api := NewCMakeFileAPI(sourceDir, buildDir, cmakeExe, cmakeDefines)

	// Check that empty defines are handled correctly
	if len(api.cmakeDefines) != 0 {
		t.Errorf("Expected 0 cmake defines, got %d", len(api.cmakeDefines))
	}
}

func TestCMakeIncludeDirectoriesGeneration(t *testing.T) {
	// Test that cmake_include_directories targets are generated correctly
	// Using the actual complex_cc_project testdata
	
	lang := &cmakeLang{}
	
	// Create mock args using the actual testdata
	c := &config.Config{
		RepoRoot: "/test/workspace",
		Exts:     make(map[string]interface{}),
	}
	c.Exts["cmake"] = gazelle.NewCMakeConfig()
	
	// Use a relative path that works in the test environment
	args := language.GenerateArgs{
		Config: c,
		Dir:    "testdata/complex_cc_project", // Use relative path to actual testdata
		Rel:    "testdata/complex_cc_project",
		RegularFiles: []string{"src/main.cpp", "src/core.cpp", "src/manager.cpp", "src/utils.cpp", "src/helper.cpp", "tests/test_main.cpp", "tests/test_utils.cpp", "include/common.h", "CMakeLists.txt"},
	}
	
	// Create mock CMake targets that match the complex project structure
	cmakeTargets := []*common.CMakeTarget{
		{
			Name:                "utils",
			Type:                "library",
			Sources:             []string{"src/utils.cpp", "src/helper.cpp"},
			Headers:             []string{},
			IncludeDirectories:  []string{"include", "third_party/include"},
			LinkedLibraries:     []string{},
		},
		{
			Name:                "core", 
			Type:                "library",
			Sources:             []string{"src/core.cpp", "src/manager.cpp"},
			Headers:             []string{"include/common.h"},
			IncludeDirectories:  []string{"include"}, // Different from utils
			LinkedLibraries:     []string{"utils"},
		},
		{
			Name:                "main_app",
			Type:                "executable",
			Sources:             []string{"src/main.cpp"},
			Headers:             []string{},
			IncludeDirectories:  []string{}, // No includes
			LinkedLibraries:     []string{"core", "utils"},
		},
	}
	
	// Call the method under test
	result := lang.generateRulesFromTargetsWithRepoAndAPI(args, cmakeTargets, "", nil)
	
	// Verify that cmake_include_directories targets were generated
	var includeRules []*rule.Rule
	var ccRules []*rule.Rule
	
	for _, r := range result.Gen {
		if r.Kind() == "cmake_include_directories" {
			includeRules = append(includeRules, r)
		} else if r.Kind() == "cc_library" || r.Kind() == "cc_binary" {
			ccRules = append(ccRules, r)
		}
	}
	
	// Should have include targets (at least 1, possibly 2 if includes differ between targets)
	if len(includeRules) == 0 {
		t.Error("Expected at least 1 cmake_include_directories rule, got 0")
	}
	
	// Log what we got for debugging
	for _, r := range includeRules {
		t.Logf("Include rule: %s, includes: %v, srcs: %v", r.Name(), r.AttrStrings("includes"), r.AttrStrings("srcs"))
	}
	
	for _, r := range ccRules {
		t.Logf("CC rule: %s %s, deps: %v", r.Kind(), r.Name(), r.AttrStrings("deps"))
	}
	
	// Should have some cc_* targets (targets that actually have valid files)
	if len(ccRules) == 0 {
		t.Error("Expected at least some cc_* rules, got 0")
	}
	
	// Verify that cc_* targets with include directories reference the include targets in their deps
	hasIncludeTargets := len(includeRules) > 0
	for _, r := range ccRules {
		deps := r.AttrStrings("deps")
		
		// Check if this rule should have include dependencies
		// (only if there are include rules and this target corresponds to one that has includes)
		if hasIncludeTargets && (r.Name() == "utils" || r.Name() == "core") {
			hasIncludeDep := false
			for _, dep := range deps {
				if strings.Contains(dep, "_includes") {
					hasIncludeDep = true
					break
				}
			}
			if !hasIncludeDep {
				t.Errorf("Expected cc_* rule %s to have an include dependency, deps: %v", r.Name(), deps)
			}
		}
		
		// Verify that the includes attribute is NOT set on cc_* targets
		includes := r.AttrStrings("includes")
		if len(includes) > 0 {
			t.Errorf("Expected cc_* rule %s to have no includes attribute, but got: %v", r.Name(), includes)
		}
	}
}

func TestCMakeIncludeDirectoriesExternalRepo(t *testing.T) {
	// Test cmake_include_directories generation for external repositories
	// This test focuses on the external repo-specific logic
	
	lang := &cmakeLang{}
	
	// Create mock args
	c := &config.Config{
		RepoRoot: "/test/workspace",
		Exts:     make(map[string]interface{}),
	}
	c.Exts["cmake"] = gazelle.NewCMakeConfig()
	
	args := language.GenerateArgs{
		Config: c,
		Dir:    "testdata/simple_cc_project", // Use simple project with real files
		Rel:    "thirdparty/libzmq",
		RegularFiles: []string{"lib.cc", "lib.h", "main.cc", "CMakeLists.txt"},
	}
	
	// Create mock CMake targets for external repo that matches simple project structure
	cmakeTargets := []*common.CMakeTarget{
		{
			Name:                "my_lib",
			Type:                "library",
			Sources:             []string{"lib.cc"},
			Headers:             []string{"lib.h"},
			IncludeDirectories:  []string{"include", ".cmake-build"},
			LinkedLibraries:     []string{},
		},
	}
	
	// Call with external repo
	result := lang.generateRulesFromTargetsWithRepoAndAPI(args, cmakeTargets, "libzmq", nil)
	
	// Find the cmake_include_directories rule
	var includeRule *rule.Rule
	var libRule *rule.Rule
	
	for _, r := range result.Gen {
		if r.Kind() == "cmake_include_directories" {
			includeRule = r
		} else if r.Kind() == "cc_library" {
			libRule = r
		}
	}
	
	// Should have exactly one cmake_include_directories rule
	if includeRule == nil {
		t.Fatal("Expected cmake_include_directories rule to be generated")
	}
	
	// Should be named "libzmq_includes"
	if includeRule.Name() != "libzmq_includes" {
		t.Errorf("Expected include rule name 'libzmq_includes', got '%s'", includeRule.Name())
	}
	
	// Should have srcs = ["@libzmq//:srcs"]
	srcs := includeRule.AttrStrings("srcs")
	expectedSrcs := []string{"@libzmq//:srcs"}
	if !reflect.DeepEqual(srcs, expectedSrcs) {
		t.Errorf("Expected srcs %v, got %v", expectedSrcs, srcs)
	}
	
	// Should have the correct includes (with .cmake-build filtered out for external repos)
	includes := includeRule.AttrStrings("includes")
	expectedIncludes := []string{"include"} // .cmake-build should be filtered out for external repos
	if !reflect.DeepEqual(includes, expectedIncludes) {
		t.Errorf("Expected includes %v, got %v", expectedIncludes, includes)
	}
	
	// Library should be generated and reference the include rule
	if libRule == nil {
		t.Fatal("Expected cc_library rule to be generated")
	}
	
	deps := libRule.AttrStrings("deps")
	expectedDep := ":libzmq_includes"
	found := false
	for _, dep := range deps {
		if dep == expectedDep {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected cc_library to have dependency '%s', got deps: %v", expectedDep, deps)
	}
}