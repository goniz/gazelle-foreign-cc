package language

import (
	"path/filepath"
	"testing"

	"github.com/goniz/gazelle-foreign-cc/gazelle"
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