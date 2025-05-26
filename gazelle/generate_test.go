package gazelle

import (
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
func createMockGenerateArgs(t *testing.T, relDir string, files []string, genFiles []string) language.GenerateArgs {
	// Find the workspace root.
	workspaceRoot, err := bazel.Runfile("") // Gets path to the current directory within the runfiles tree
	if err != nil {
		t.Fatalf("Could not find workspace root: %v", err)
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

	// Combine all files into RegularFiles - CMakeLists.txt should be treated as a regular file
	allFiles := append(files, genFiles...)

	return language.GenerateArgs{
		Config:       c,
		Dir:          absDir, // Absolute path to the directory being processed
		Rel:          filepath.Base(relDir), // Relative path from repo root (or a common root for testdata)
		RegularFiles: allFiles,
		GenFiles:     []string{}, // Generated files
		File:         nil,        // Represents the existing BUILD file, nil if generating anew
	}
}

func TestBasic(t *testing.T) {
	// Simple test to verify the test infrastructure works
	t.Log("Basic test passed")
}

func TestGenerateRules_LibZMQProject(t *testing.T) {
	// Test the libzmq project which has multiple libraries and executables
	projectRelDir := "testdata/libzmq_project"

	args := createMockGenerateArgs(t,
		projectRelDir,
		[]string{
			"src/zmq.cpp", "src/socket.cpp", "src/context.cpp", "src/message.cpp",
			"src/poller.cpp", "src/address.cpp", "src/tcp_connecter.cpp", "src/tcp_listener.cpp",
			"include/zmq.h", "include/zmq_utils.h",
			"tests/test_basic.cpp", "perf/inproc_lat.cpp",
		},
		[]string{"CMakeLists.txt"},
	)

	// Expected rules based on the CMakeLists.txt
	expectedRuleZmq := rule.NewRule("cc_library", "zmq")
	expectedRuleZmq.SetAttr("srcs", []string{
		"src/zmq.cpp", "src/socket.cpp", "src/context.cpp", "src/message.cpp",
		"src/poller.cpp", "src/address.cpp", "src/tcp_connecter.cpp", "src/tcp_listener.cpp",
	})
	expectedRuleZmq.SetAttr("hdrs", []string{"include/zmq.h", "include/zmq_utils.h"})

	expectedRuleZmqStatic := rule.NewRule("cc_library", "zmq-static")
	expectedRuleZmqStatic.SetAttr("srcs", []string{
		"src/zmq.cpp", "src/socket.cpp", "src/context.cpp", "src/message.cpp",
		"src/poller.cpp", "src/address.cpp", "src/tcp_connecter.cpp", "src/tcp_listener.cpp",
	})

	expectedRuleZmqTest := rule.NewRule("cc_binary", "zmq_test")
	expectedRuleZmqTest.SetAttr("srcs", []string{"tests/test_basic.cpp"})

	expectedRulePerfTest := rule.NewRule("cc_binary", "perf_inproc_lat")
	expectedRulePerfTest.SetAttr("srcs", []string{"perf/inproc_lat.cpp"})

	expectedRules := []*rule.Rule{expectedRuleZmq, expectedRuleZmqStatic, expectedRuleZmqTest, expectedRulePerfTest}

	// Call the function under test
	result := GenerateRules(args)

	// --- Verification ---
	if len(result.Gen) != len(expectedRules) {
		t.Errorf("Expected %d rules, got %d. Generated: %v", len(expectedRules), len(result.Gen), result.Gen)
		for _, r := range result.Gen {
			t.Logf("Generated rule: %s %s, srcs: %v, hdrs: %v", r.Kind(), r.Name(), r.AttrStrings("srcs"), r.AttrStrings("hdrs"))
		}
		return
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
		sort.Strings(gotSrcs)
		sort.Strings(expectedSrcs)
		if !reflect.DeepEqual(gotSrcs, expectedSrcs) {
			t.Errorf("Rule %s %s: Expected srcs %v, got %v",
				gotRule.Kind(), gotRule.Name(), expectedSrcs, gotSrcs)
		}

		// Compare hdrs
		gotHdrs := gotRule.AttrStrings("hdrs")
		expectedHdrs := expectedRule.AttrStrings("hdrs")
		sort.Strings(gotHdrs)
		sort.Strings(expectedHdrs)
		if !reflect.DeepEqual(gotHdrs, expectedHdrs) {
			t.Errorf("Rule %s %s: Expected hdrs %v, got %v",
				gotRule.Kind(), gotRule.Name(), expectedHdrs, gotHdrs)
		}
	}

	// Check Empty rules
	if len(result.Empty) != len(expectedRules) {
		t.Errorf("Expected %d empty rules, got %d.", len(expectedRules), len(result.Empty))
	} else {
		sort.Slice(result.Empty, func(i, j int) bool { return result.Empty[i].Name() < result.Empty[j].Name() })
		for i, gotEmptyRule := range result.Empty {
			expectedRule := expectedRules[i]
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
