package gazelle

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
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



func TestGenerateRules_RealLibZMQ(t *testing.T) {
	// Test with real libzmq source code
	// This test uses the actual libzmq CMakeLists.txt content and realistic source structure
	
	// Create a temporary directory to simulate the real libzmq project
	tmpDir, err := os.MkdirTemp("", "real_libzmq_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Use the actual libzmq CMakeLists.txt content (simplified but realistic)
	// This is based on the real libzmq-4.3.5 CMakeLists.txt structure
	realCMakeContent := `# CMake build script for ZeroMQ
project(ZeroMQ)
cmake_minimum_required(VERSION 2.8.12)

include(CheckIncludeFiles)
include(CheckCCompilerFlag)
include(CheckCXXCompilerFlag)
include(CheckLibraryExists)
include(CheckCSourceCompiles)
include(CheckCSourceRuns)
include(CMakeDependentOption)
include(CheckCXXSymbolExists)
include(CheckStructHasMember)
include(CheckTypeSize)
include(FindThreads)
include(GNUInstallDirs)

# Set lists to empty beforehand
set(sources)
set(cxx-sources)
set(html-docs)
set(target_outputs)

# Source files (subset of real libzmq sources)
set(sources
    src/address.cpp
    src/client.cpp
    src/clock.cpp
    src/ctx.cpp
    src/curve_client.cpp
    src/curve_server.cpp
    src/dealer.cpp
    src/devpoll.cpp
    src/dgram.cpp
    src/dist.cpp
    src/epoll.cpp
    src/err.cpp
    src/fq.cpp
    src/io_object.cpp
    src/io_thread.cpp
    src/ip.cpp
    src/ipc_address.cpp
    src/ipc_connecter.cpp
    src/ipc_listener.cpp
    src/kqueue.cpp
    src/lb.cpp
    src/mailbox.cpp
    src/mailbox_safe.cpp
    src/mechanism.cpp
    src/metadata.cpp
    src/msg.cpp
    src/mtrie.cpp
    src/object.cpp
    src/options.cpp
    src/own.cpp
    src/pair.cpp
    src/pipe.cpp
    src/plain_client.cpp
    src/plain_server.cpp
    src/poll.cpp
    src/poller_base.cpp
    src/proxy.cpp
    src/pub.cpp
    src/pull.cpp
    src/push.cpp
    src/random.cpp
    src/raw_decoder.cpp
    src/raw_encoder.cpp
    src/reaper.cpp
    src/rep.cpp
    src/req.cpp
    src/router.cpp
    src/select.cpp
    src/server.cpp
    src/session_base.cpp
    src/signaler.cpp
    src/socket_base.cpp
    src/socks.cpp
    src/socks_connecter.cpp
    src/stream.cpp
    src/stream_engine.cpp
    src/sub.cpp
    src/tcp.cpp
    src/tcp_address.cpp
    src/tcp_connecter.cpp
    src/tcp_listener.cpp
    src/thread.cpp
    src/trie.cpp
    src/v1_decoder.cpp
    src/v1_encoder.cpp
    src/v2_decoder.cpp
    src/v2_encoder.cpp
    src/xpub.cpp
    src/xsub.cpp
    src/zmq.cpp
    src/zmq_utils.cpp
)

set(public_headers
    include/zmq.h
    include/zmq_utils.h
)

# Build shared library
add_library(libzmq SHARED 
    src/address.cpp
    src/client.cpp
    src/clock.cpp
    src/ctx.cpp
    src/curve_client.cpp
    src/curve_server.cpp
    src/dealer.cpp
    src/devpoll.cpp
    src/dgram.cpp
    src/dist.cpp
    src/epoll.cpp
    src/err.cpp
    src/fq.cpp
    src/io_object.cpp
    src/io_thread.cpp
    src/ip.cpp
    src/ipc_address.cpp
    src/ipc_connecter.cpp
    src/ipc_listener.cpp
    src/kqueue.cpp
    src/lb.cpp
    src/mailbox.cpp
    src/mailbox_safe.cpp
    src/mechanism.cpp
    src/metadata.cpp
    src/msg.cpp
    src/mtrie.cpp
    src/object.cpp
    src/options.cpp
    src/own.cpp
    src/pair.cpp
    src/pipe.cpp
    src/plain_client.cpp
    src/plain_server.cpp
    src/poll.cpp
    src/poller_base.cpp
    src/proxy.cpp
    src/pub.cpp
    src/pull.cpp
    src/push.cpp
    src/random.cpp
    src/raw_decoder.cpp
    src/raw_encoder.cpp
    src/reaper.cpp
    src/rep.cpp
    src/req.cpp
    src/router.cpp
    src/select.cpp
    src/server.cpp
    src/session_base.cpp
    src/signaler.cpp
    src/socket_base.cpp
    src/socks.cpp
    src/socks_connecter.cpp
    src/stream.cpp
    src/stream_engine.cpp
    src/sub.cpp
    src/tcp.cpp
    src/tcp_address.cpp
    src/tcp_connecter.cpp
    src/tcp_listener.cpp
    src/thread.cpp
    src/trie.cpp
    src/v1_decoder.cpp
    src/v1_encoder.cpp
    src/v2_decoder.cpp
    src/v2_encoder.cpp
    src/xpub.cpp
    src/xsub.cpp
    src/zmq.cpp
    src/zmq_utils.cpp
    include/zmq.h
    include/zmq_utils.h
)
set_target_properties(libzmq PROPERTIES
    OUTPUT_NAME "zmq"
)

# Build static library
add_library(libzmq-static STATIC 
    src/address.cpp
    src/client.cpp
    src/clock.cpp
    src/ctx.cpp
    src/curve_client.cpp
    src/curve_server.cpp
    src/dealer.cpp
    src/devpoll.cpp
    src/dgram.cpp
    src/dist.cpp
    src/epoll.cpp
    src/err.cpp
    src/fq.cpp
    src/io_object.cpp
    src/io_thread.cpp
    src/ip.cpp
    src/ipc_address.cpp
    src/ipc_connecter.cpp
    src/ipc_listener.cpp
    src/kqueue.cpp
    src/lb.cpp
    src/mailbox.cpp
    src/mailbox_safe.cpp
    src/mechanism.cpp
    src/metadata.cpp
    src/msg.cpp
    src/mtrie.cpp
    src/object.cpp
    src/options.cpp
    src/own.cpp
    src/pair.cpp
    src/pipe.cpp
    src/plain_client.cpp
    src/plain_server.cpp
    src/poll.cpp
    src/poller_base.cpp
    src/proxy.cpp
    src/pub.cpp
    src/pull.cpp
    src/push.cpp
    src/random.cpp
    src/raw_decoder.cpp
    src/raw_encoder.cpp
    src/reaper.cpp
    src/rep.cpp
    src/req.cpp
    src/router.cpp
    src/select.cpp
    src/server.cpp
    src/session_base.cpp
    src/signaler.cpp
    src/socket_base.cpp
    src/socks.cpp
    src/socks_connecter.cpp
    src/stream.cpp
    src/stream_engine.cpp
    src/sub.cpp
    src/tcp.cpp
    src/tcp_address.cpp
    src/tcp_connecter.cpp
    src/tcp_listener.cpp
    src/thread.cpp
    src/trie.cpp
    src/v1_decoder.cpp
    src/v1_encoder.cpp
    src/v2_decoder.cpp
    src/v2_encoder.cpp
    src/xpub.cpp
    src/xsub.cpp
    src/zmq.cpp
    src/zmq_utils.cpp
)
set_target_properties(libzmq-static PROPERTIES
    OUTPUT_NAME "zmq"
)

# Performance tools
add_executable(inproc_lat perf/inproc_lat.cpp)
target_link_libraries(inproc_lat libzmq)

add_executable(inproc_thr perf/inproc_thr.cpp)
target_link_libraries(inproc_thr libzmq)

add_executable(local_lat perf/local_lat.cpp)
target_link_libraries(local_lat libzmq)

add_executable(local_thr perf/local_thr.cpp)
target_link_libraries(local_thr libzmq)

add_executable(remote_lat perf/remote_lat.cpp)
target_link_libraries(remote_lat libzmq)

add_executable(remote_thr perf/remote_thr.cpp)
target_link_libraries(remote_thr libzmq)

# Tests
enable_testing()

add_executable(test_system tests/test_system.cpp)
target_link_libraries(test_system libzmq)
add_test(NAME test_system COMMAND test_system)

add_executable(test_pair_inproc tests/test_pair_inproc.cpp)
target_link_libraries(test_pair_inproc libzmq)
add_test(NAME test_pair_inproc COMMAND test_pair_inproc)

add_executable(test_pair_tcp tests/test_pair_tcp.cpp)
target_link_libraries(test_pair_tcp libzmq)
add_test(NAME test_pair_tcp COMMAND test_pair_tcp)

add_executable(test_reqrep_inproc tests/test_reqrep_inproc.cpp)
target_link_libraries(test_reqrep_inproc libzmq)
add_test(NAME test_reqrep_inproc COMMAND test_reqrep_inproc)

add_executable(test_reqrep_tcp tests/test_reqrep_tcp.cpp)
target_link_libraries(test_reqrep_tcp libzmq)
add_test(NAME test_reqrep_tcp COMMAND test_reqrep_tcp)

add_executable(test_hwm tests/test_hwm.cpp)
target_link_libraries(test_hwm libzmq)
add_test(NAME test_hwm COMMAND test_hwm)

add_executable(test_hwm_pubsub tests/test_hwm_pubsub.cpp)
target_link_libraries(test_hwm_pubsub libzmq)
add_test(NAME test_hwm_pubsub COMMAND test_hwm_pubsub)

add_executable(test_reqrep_device tests/test_reqrep_device.cpp)
target_link_libraries(test_reqrep_device libzmq)
add_test(NAME test_reqrep_device COMMAND test_reqrep_device)

add_executable(test_sub_forward tests/test_sub_forward.cpp)
target_link_libraries(test_sub_forward libzmq)
add_test(NAME test_sub_forward COMMAND test_sub_forward)

add_executable(test_invalid_rep tests/test_invalid_rep.cpp)
target_link_libraries(test_invalid_rep libzmq)
add_test(NAME test_invalid_rep COMMAND test_invalid_rep)

add_executable(test_msg_flags tests/test_msg_flags.cpp)
target_link_libraries(test_msg_flags libzmq)
add_test(NAME test_msg_flags COMMAND test_msg_flags)

add_executable(test_msg_ffn tests/test_msg_ffn.cpp)
target_link_libraries(test_msg_ffn libzmq)
add_test(NAME test_msg_ffn COMMAND test_msg_ffn)
`

	// Write the CMakeLists.txt file
	cmakeFile := filepath.Join(tmpDir, "CMakeLists.txt")
	if err := os.WriteFile(cmakeFile, []byte(realCMakeContent), 0644); err != nil {
		t.Fatalf("Failed to write CMakeLists.txt: %v", err)
	}

	// Create realistic source directory structure
	srcDir := filepath.Join(tmpDir, "src")
	includeDir := filepath.Join(tmpDir, "include")
	testsDir := filepath.Join(tmpDir, "tests")
	perfDir := filepath.Join(tmpDir, "perf")
	
	for _, dir := range []string{srcDir, includeDir, testsDir, perfDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create realistic source files that match the CMakeLists.txt sources list
	sourceFiles := []string{
		"src/address.cpp", "src/client.cpp", "src/clock.cpp", "src/ctx.cpp",
		"src/curve_client.cpp", "src/curve_server.cpp", "src/dealer.cpp", "src/devpoll.cpp",
		"src/dgram.cpp", "src/dist.cpp", "src/epoll.cpp", "src/err.cpp",
		"src/fq.cpp", "src/io_object.cpp", "src/io_thread.cpp", "src/ip.cpp",
		"src/ipc_address.cpp", "src/ipc_connecter.cpp", "src/ipc_listener.cpp", "src/kqueue.cpp",
		"src/lb.cpp", "src/mailbox.cpp", "src/mailbox_safe.cpp", "src/mechanism.cpp",
		"src/metadata.cpp", "src/msg.cpp", "src/mtrie.cpp", "src/object.cpp",
		"src/options.cpp", "src/own.cpp", "src/pair.cpp", "src/pipe.cpp",
		"src/plain_client.cpp", "src/plain_server.cpp", "src/poll.cpp", "src/poller_base.cpp",
		"src/proxy.cpp", "src/pub.cpp", "src/pull.cpp", "src/push.cpp",
		"src/random.cpp", "src/raw_decoder.cpp", "src/raw_encoder.cpp", "src/reaper.cpp",
		"src/rep.cpp", "src/req.cpp", "src/router.cpp", "src/select.cpp",
		"src/server.cpp", "src/session_base.cpp", "src/signaler.cpp", "src/socket_base.cpp",
		"src/socks.cpp", "src/socks_connecter.cpp", "src/stream.cpp", "src/stream_engine.cpp",
		"src/sub.cpp", "src/tcp.cpp", "src/tcp_address.cpp", "src/tcp_connecter.cpp",
		"src/tcp_listener.cpp", "src/thread.cpp", "src/trie.cpp", "src/v1_decoder.cpp",
		"src/v1_encoder.cpp", "src/v2_decoder.cpp", "src/v2_encoder.cpp", "src/xpub.cpp",
		"src/xsub.cpp", "src/zmq.cpp", "src/zmq_utils.cpp",
	}
	
	headerFiles := []string{
		"include/zmq.h", "include/zmq_utils.h",
	}
	
	testFiles := []string{
		"tests/test_system.cpp", "tests/test_pair_inproc.cpp", "tests/test_pair_tcp.cpp",
		"tests/test_reqrep_inproc.cpp", "tests/test_reqrep_tcp.cpp", "tests/test_hwm.cpp",
		"tests/test_hwm_pubsub.cpp", "tests/test_reqrep_device.cpp", "tests/test_sub_forward.cpp",
		"tests/test_invalid_rep.cpp", "tests/test_msg_flags.cpp", "tests/test_msg_ffn.cpp",
	}
	
	perfFiles := []string{
		"perf/inproc_lat.cpp", "perf/inproc_thr.cpp", "perf/local_lat.cpp",
		"perf/local_thr.cpp", "perf/remote_lat.cpp", "perf/remote_thr.cpp",
	}

	// Create all the files
	allFiles := append(append(append(sourceFiles, headerFiles...), testFiles...), perfFiles...)
	for _, file := range allFiles {
		filePath := filepath.Join(tmpDir, file)
		content := fmt.Sprintf("// %s - Real libzmq source file\n", filepath.Base(file))
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", filePath, err)
		}
	}

	// Create a mock GenerateArgs pointing to our temp directory with real structure
	// Use a relative path for the mock args
	relDir := "real_libzmq_test"
	args := createMockGenerateArgs(t, relDir, allFiles, []string{"CMakeLists.txt"})
	// Override the Dir to point to our actual temp directory
	args.Dir = tmpDir

	// Generate rules using the GenerateRules function
	result := GenerateRules(args)

	// Verify we got rules - should be many from the real libzmq structure
	if len(result.Gen) == 0 {
		t.Error("Expected to generate rules from real libzmq structure, got none")
		return
	}

	// Collect generated rules by name and kind
	rulesByKind := make(map[string][]string)
	rulesByName := make(map[string]*rule.Rule)
	
	for _, r := range result.Gen {
		rulesByKind[r.Kind()] = append(rulesByKind[r.Kind()], r.Name())
		rulesByName[r.Name()] = r
	}

	// Verify we have cc_library rules (should include libzmq and libzmq-static)
	if libraries, exists := rulesByKind["cc_library"]; exists {
		t.Logf("Generated cc_library rules from real libzmq: %v", libraries)
		
		// Check for libzmq library
		if libzmqRule, exists := rulesByName["libzmq"]; exists {
			if srcs := libzmqRule.AttrStrings("srcs"); len(srcs) == 0 {
				t.Error("Expected libzmq library to have source files")
			} else {
				t.Logf("libzmq library has %d source files", len(srcs))
			}
			if hdrs := libzmqRule.AttrStrings("hdrs"); len(hdrs) == 0 {
				t.Error("Expected libzmq library to have header files")
			} else {
				t.Logf("libzmq library has %d header files", len(hdrs))
			}
		} else {
			t.Error("Expected to find libzmq library rule")
		}
		
		// Check for libzmq-static library
		if libzmqStaticRule, exists := rulesByName["libzmq-static"]; exists {
			if srcs := libzmqStaticRule.AttrStrings("srcs"); len(srcs) == 0 {
				t.Error("Expected libzmq-static library to have source files")
			} else {
				t.Logf("libzmq-static library has %d source files", len(srcs))
			}
		} else {
			t.Error("Expected to find libzmq-static library rule")
		}
	} else {
		t.Error("Expected to generate cc_library rules from real libzmq")
	}

	// Verify we have cc_binary rules (tests and perf tools)
	if binaries, exists := rulesByKind["cc_binary"]; exists {
		t.Logf("Generated cc_binary rules from real libzmq: %v", binaries)
		
		// Should have performance tools
		perfToolsFound := 0
		testsFound := 0
		for _, binaryName := range binaries {
			if strings.HasPrefix(binaryName, "inproc_") || strings.HasPrefix(binaryName, "local_") || strings.HasPrefix(binaryName, "remote_") {
				perfToolsFound++
			}
			if strings.HasPrefix(binaryName, "test_") {
				testsFound++
			}
			
			// Check that binary has source files
			if binaryRule := rulesByName[binaryName]; binaryRule != nil {
				if srcs := binaryRule.AttrStrings("srcs"); len(srcs) == 0 {
					t.Errorf("Expected binary %s to have source files", binaryName)
				}
			}
		}
		
		if perfToolsFound == 0 {
			t.Error("Expected to find performance tool binaries")
		} else {
			t.Logf("Found %d performance tool binaries", perfToolsFound)
		}
		
		if testsFound == 0 {
			t.Error("Expected to find test binaries")
		} else {
			t.Logf("Found %d test binaries", testsFound)
		}
	} else {
		t.Error("Expected to generate cc_binary rules from real libzmq")
	}

	t.Logf("Successfully generated %d rules from real libzmq project:", len(result.Gen))
	for kind, names := range rulesByKind {
		t.Logf("  %s (%d): %v", kind, len(names), names)
	}
}

// TODO: Add more test cases:
// - CMakeLists.txt with no targets
// - CMakeLists.txt with targets but no matching source files in args.RegularFiles
// - CMakeLists.txt with only add_library or only add_executable
// - File names with unusual characters (if your regexes are too simple)
// - Case variations in add_library/add_executable (already handled by (?i) in regex)
