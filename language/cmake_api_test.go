package language

import (
	"encoding/json"
	"testing"
)

func TestTargetJSONParsing(t *testing.T) {
	// Test parsing of actual CMake File API target JSON structure
	testCases := []struct {
		name     string
		jsonData string
		shouldFail bool
	}{
		{
			name: "simple_library_target",
			jsonData: `{
				"archive" : {},
				"artifacts" : [{"path" : "libmy_lib.a"}],
				"backtrace" : 1,
				"backtraceGraph" : {
					"commands" : ["add_library"],
					"files" : ["CMakeLists.txt"],
					"nodes" : [{"file" : 0}, {"command" : 0, "file" : 0, "line" : 5, "parent" : 0}]
				},
				"compileGroups" : [{
					"language" : "CXX",
					"sourceIndexes" : [0]
				}],
				"id" : "my_lib::@6890427a1f51a3e7e1df",
				"name" : "my_lib",
				"nameOnDisk" : "libmy_lib.a",
				"paths" : {"build" : ".", "source" : "."},
				"sourceGroups" : [{
					"name" : "Source Files",
					"sourceIndexes" : [0]
				}],
				"sources" : [{
					"backtrace" : 1,
					"compileGroupIndex" : 0,
					"path" : "lib.cc",
					"sourceGroupIndex" : 0
				}],
				"type" : "STATIC_LIBRARY"
			}`,
			shouldFail: false,
		},
		{
			name: "executable_with_dependencies",
			jsonData: `{
				"artifacts" : [{"path" : "app"}],
				"backtrace" : 1,
				"backtraceGraph" : {
					"commands" : ["add_executable", "target_link_libraries"],
					"files" : ["CMakeLists.txt"],
					"nodes" : [
						{"file" : 0},
						{"command" : 0, "file" : 0, "line" : 8, "parent" : 0},
						{"command" : 1, "file" : 0, "line" : 11, "parent" : 0}
					]
				},
				"compileGroups" : [{
					"language" : "CXX",
					"sourceIndexes" : [0]
				}],
				"dependencies" : [{
					"backtrace" : 2,
					"id" : "my_lib::@6890427a1f51a3e7e1df"
				}],
				"id" : "app::@6890427a1f51a3e7e1df",
				"link" : {
					"commandFragments" : [
						{"fragment" : "", "role" : "flags"},
						{"backtrace" : 2, "fragment" : "libmy_lib.a", "role" : "libraries"}
					],
					"language" : "CXX"
				},
				"name" : "app",
				"nameOnDisk" : "app",
				"paths" : {"build" : ".", "source" : "."},
				"sources" : [{
					"backtrace" : 1,
					"compileGroupIndex" : 0,
					"path" : "main.cc",
					"sourceGroupIndex" : 0
				}],
				"type" : "EXECUTABLE"
			}`,
			shouldFail: false,
		},
		{
			name: "target_without_optional_fields",
			jsonData: `{
				"id" : "minimal::@123",
				"name" : "minimal",
				"type" : "STATIC_LIBRARY"
			}`,
			shouldFail: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var target Target
			err := json.Unmarshal([]byte(tc.jsonData), &target)
			
			if tc.shouldFail && err == nil {
				t.Errorf("Expected parsing to fail for %s, but it succeeded", tc.name)
			}
			if !tc.shouldFail && err != nil {
				t.Errorf("Expected parsing to succeed for %s, but got error: %v", tc.name, err)
			}
			
			if err == nil {
				// Basic validation
				if target.Name == "" {
					t.Errorf("Target name should not be empty for %s", tc.name)
				}
				if target.ID == "" {
					t.Errorf("Target ID should not be empty for %s", tc.name)
				}
				if target.Type == "" {
					t.Errorf("Target type should not be empty for %s", tc.name)
				}
			}
		})
	}
}

func TestTargetJSONParsingWithComplexStructures(t *testing.T) {
	// Test more complex JSON structures that might appear in real projects like libzmq
	testCases := []struct {
		name     string
		jsonData string
		shouldFail bool
	}{
		{
			name: "target_with_complex_compile_groups",
			jsonData: `{
				"id": "libzmq::@abc123",
				"name": "libzmq",
				"type": "STATIC_LIBRARY",
				"compileGroups": [
					{
						"language": "CXX",
						"sourceIndexes": [0, 1, 2],
						"includes": [
							{"path": "/usr/include", "isSystem": true},
							{"path": "include", "isSystem": false},
							{"path": "src", "isSystem": false, "backtrace": 5}
						],
						"preprocessorDefinitions": [
							{"define": "ZMQ_STATIC", "backtrace": 3},
							{"define": "NDEBUG"}
						],
						"compileCommandFragments": [
							{"fragment": "-Wall", "backtrace": 1},
							{"fragment": "-O2"}
						]
					}
				],
				"sources": [
					{"path": "src/zmq.cpp", "compileGroupIndex": 0},
					{"path": "src/socket.cpp", "compileGroupIndex": 0},
					{"path": "include/zmq.h"}
				]
			}`,
			shouldFail: false,
		},
		{
			name: "target_with_additional_unknown_fields",
			jsonData: `{
				"id": "test::@xyz789",
				"name": "test",
				"type": "EXECUTABLE",
				"unknownField": "someValue",
				"anotherUnknownField": {
					"nested": "data",
					"array": [1, 2, 3]
				},
				"yetAnotherField": ["item1", "item2"]
			}`,
			shouldFail: false,
		},
		{
			name: "target_with_empty_compile_groups",
			jsonData: `{
				"id": "empty::@def456",
				"name": "empty",
				"type": "UTILITY",
				"compileGroups": []
			}`,
			shouldFail: false,
		},
		{
			name: "target_with_null_optional_fields",
			jsonData: `{
				"id": "minimal::@ghi789",
				"name": "minimal",
				"type": "INTERFACE_LIBRARY",
				"compileGroups": null,
				"archive": null,
				"backtraceGraph": null
			}`,
			shouldFail: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var target Target
			err := json.Unmarshal([]byte(tc.jsonData), &target)
			
			if tc.shouldFail && err == nil {
				t.Errorf("Expected parsing to fail for %s, but it succeeded", tc.name)
			}
			if !tc.shouldFail && err != nil {
				t.Errorf("Expected parsing to succeed for %s, but got error: %v", tc.name, err)
			}
			
			if err == nil {
				// Basic validation
				if target.Name == "" {
					t.Errorf("Target name should not be empty for %s", tc.name)
				}
				if target.ID == "" {
					t.Errorf("Target ID should not be empty for %s", tc.name)
				}
				if target.Type == "" {
					t.Errorf("Target type should not be empty for %s", tc.name)
				}
				
				// Test that we can extract include directories without errors
				includeDirectories := extractIncludeDirectories(&target, "/test")
				// Should not panic or fail, even if empty
				_ = includeDirectories
			}
		})
	}
}

func TestTargetJSONParsingWithMissingFields(t *testing.T) {
	// Test that parsing fails gracefully with completely invalid JSON
	invalidJSONs := []string{
		`{"name": "test", "id": "test::123"}`, // missing type
		`{"type": "STATIC_LIBRARY"}`,          // missing name and id
		`{]`,                                  // malformed JSON
	}
	
	for i, jsonData := range invalidJSONs {
		var target Target
		err := json.Unmarshal([]byte(jsonData), &target)
		
		// For missing required fields, Go will still parse successfully but fields will be empty
		// Only malformed JSON should actually fail
		if i == 2 && err == nil {
			t.Errorf("Expected malformed JSON to fail parsing, but it succeeded")
		}
	}
}