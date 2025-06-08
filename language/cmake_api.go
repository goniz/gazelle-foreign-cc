package language

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"log"

	"github.com/goniz/gazelle-foreign-cc/gazelle"
)

// CMake File API structures
// These structures match the JSON schema from the CMake File API documentation

// APIIndex represents the main index file from CMake File API
type APIIndex struct {
	CMake struct {
		Version struct {
			Major int `json:"major"`
			Minor int `json:"minor"`
			Patch int `json:"patch"`
		} `json:"version"`
	} `json:"cmake"`
	Objects []struct {
		Kind       string `json:"kind"`
		Version    struct {
			Major int `json:"major"`
			Minor int `json:"minor"`
		} `json:"version"`
		JSONFile string `json:"jsonFile"`
	} `json:"objects"`
	Reply struct {
		ClientStateless struct {
			Query struct {
				Requests []struct {
					Kind string `json:"kind"`
				} `json:"requests"`
			} `json:"query"`
		} `json:"client-stateless"`
	} `json:"reply"`
}

// Codemodel represents the codemodel object from CMake File API
type Codemodel struct {
	Kind    string `json:"kind"`
	Version struct {
		Major int `json:"major"`
		Minor int `json:"minor"`
	} `json:"version"`
	Paths struct {
		Source string `json:"source"`
		Build  string `json:"build"`
	} `json:"paths"`
	Configurations []struct {
		Name    string `json:"name"`
		Targets []struct {
			Name     string `json:"name"`
			ID       string `json:"id"`
			JSONFile string `json:"jsonFile"`
		} `json:"targets"`
		Directories []struct {
			Source   string `json:"source"`
			Build    string `json:"build"`
			JSONFile string `json:"jsonFile"`
		} `json:"directories"`
		Projects []struct {
			Name             string `json:"name"`
			DirectoryIndexes []int  `json:"directoryIndexes"`
			TargetIndexes    []int  `json:"targetIndexes"`
		} `json:"projects"`
	} `json:"configurations"`
}

// Target represents a target object from CMake File API
type Target struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Type string `json:"type"`
	Backtrace int `json:"backtrace,omitempty"`
	Paths struct {
		Source string `json:"source"`
		Build  string `json:"build"`
	} `json:"paths,omitempty"`
	NameOnDisk   string `json:"nameOnDisk,omitempty"`
	Artifacts    []struct {
		Path string `json:"path"`
	} `json:"artifacts,omitempty"`
	Build        string `json:"build,omitempty"`
	Source       string `json:"source,omitempty"`
	Install      json.RawMessage `json:"install,omitempty"`
	Link *struct {
		Language      string `json:"language"`
		CommandFragments []struct {
			Fragment  string `json:"fragment"`
			Role      string `json:"role"`
			Backtrace int    `json:"backtrace,omitempty"`
		} `json:"commandFragments,omitempty"`
		Flags []struct {
			Fragment  string `json:"fragment"`
			Backtrace int    `json:"backtrace,omitempty"`
		} `json:"flags,omitempty"`
		Libraries []struct {
			Fragment  string `json:"fragment"`
			Role      string `json:"role"`
			Backtrace int    `json:"backtrace,omitempty"`
		} `json:"libraries,omitempty"`
		Path      string `json:"path,omitempty"`
		SysRoot   string `json:"sysroot,omitempty"`
	} `json:"link,omitempty"`
	Archive json.RawMessage `json:"archive,omitempty"`
	Dependencies []struct {
		ID        string `json:"id"`
		Backtrace int    `json:"backtrace,omitempty"`
	} `json:"dependencies,omitempty"`
	Sources []struct {
		Path              string `json:"path"`
		CompileGroupIndex int    `json:"compileGroupIndex,omitempty"`
		SourceGroupIndex  int    `json:"sourceGroupIndex,omitempty"`
		IsGenerated       bool   `json:"isGenerated,omitempty"`
		Backtrace         int    `json:"backtrace,omitempty"`
	} `json:"sources,omitempty"`
	SourceGroups []struct {
		Name    string `json:"name"`
		Sources []int  `json:"sources"`
	} `json:"sourceGroups,omitempty"`
	CompileGroups json.RawMessage `json:"compileGroups,omitempty"`
	BacktraceGraph json.RawMessage `json:"backtraceGraph,omitempty"`
	Folder string `json:"folder,omitempty"`
}

// CMakeFileAPI handles interaction with CMake File API
type CMakeFileAPI struct {
	sourceDir    string
	buildDir     string
	cmakeExe     string
	cmakeDefines map[string]string
}

// NewCMakeFileAPI creates a new CMake File API handler
func NewCMakeFileAPI(sourceDir, buildDir, cmakeExe string, cmakeDefines map[string]string) *CMakeFileAPI {
	return &CMakeFileAPI{
		sourceDir:    sourceDir,
		buildDir:     buildDir,
		cmakeExe:     cmakeExe,
		cmakeDefines: cmakeDefines,
	}
}

// CreateQuery creates the query files for CMake File API
func (api *CMakeFileAPI) CreateQuery() error {
	queryDir := filepath.Join(api.buildDir, ".cmake", "api", "v1", "query")
	if err := os.MkdirAll(queryDir, 0755); err != nil {
		return fmt.Errorf("failed to create query directory: %w", err)
	}

	// Create query files for the objects we need
	queries := []string{
		"codemodel-v2",
		"cache-v2",
		"toolchains-v1",
	}

	for _, query := range queries {
		queryFile := filepath.Join(queryDir, query)
		if err := ioutil.WriteFile(queryFile, []byte{}, 0644); err != nil {
			return fmt.Errorf("failed to create query file %s: %w", query, err)
		}
	}

	return nil
}

// DetectConfigureFileCommands detects configure_file commands using a simpler approach
func (api *CMakeFileAPI) DetectConfigureFileCommands() ([]*gazelle.CMakeConfigureFile, error) {
	// Create a temporary CMake script that parses CMakeLists.txt 
	scriptPath := filepath.Join(api.buildDir, "extract_configure_files.cmake")
	
	script := fmt.Sprintf(`
# CMake script to extract configure_file commands and variables
cmake_minimum_required(VERSION 3.15)

# Read CMakeLists.txt
set(CMAKELISTS_PATH "%s/CMakeLists.txt")
if(NOT EXISTS "${CMAKELISTS_PATH}")
    message("CMakeLists.txt not found")
    return()
endif()

file(READ "${CMAKELISTS_PATH}" CONTENT)

# Simple regex-based extraction of configure_file commands
string(REGEX MATCHALL "configure_file[ \t]*\\(([^)]*)\\)" CONFIGURE_FILE_MATCHES "${CONTENT}")

# Output file path
set(OUTPUT_FILE "%s/configure_files.txt")
file(WRITE "${OUTPUT_FILE}" "")

foreach(MATCH ${CONFIGURE_FILE_MATCHES})
    # Extract arguments from configure_file(input output [options])
    string(REGEX REPLACE "configure_file[ \t]*\\([ \t]*([^ \t]+)[ \t]+([^ \t)]+).*\\)" "\\1;\\2" ARGS "${MATCH}")
    list(GET ARGS 0 INPUT_FILE)
    list(GET ARGS 1 OUTPUT_FILE_NAME)
    
    # Clean up quotes
    string(REGEX REPLACE "\"" "" INPUT_FILE "${INPUT_FILE}")
    string(REGEX REPLACE "\"" "" OUTPUT_FILE_NAME "${OUTPUT_FILE_NAME}")
    
    file(APPEND "${OUTPUT_FILE}" "CONFIGURE_FILE|${INPUT_FILE}|${OUTPUT_FILE_NAME}\n")
endforeach()

# Extract set() commands for variables
string(REGEX MATCHALL "set[ \t]*\\(([^)]*)\\)" SET_MATCHES "${CONTENT}")

foreach(MATCH ${SET_MATCHES})
    # Extract arguments from set(var value)
    string(REGEX REPLACE "set[ \t]*\\([ \t]*([^ \t]+)[ \t]+([^ \t)]+).*\\)" "\\1;\\2" ARGS "${MATCH}")
    list(LENGTH ARGS ARG_COUNT)
    if(ARG_COUNT EQUAL 2)
        list(GET ARGS 0 VAR_NAME)
        list(GET ARGS 1 VAR_VALUE)
        
        # Clean up quotes
        string(REGEX REPLACE "\"" "" VAR_NAME "${VAR_NAME}")
        string(REGEX REPLACE "\"" "" VAR_VALUE "${VAR_VALUE}")
        
        file(APPEND "${OUTPUT_FILE}" "SET|${VAR_NAME}|${VAR_VALUE}\n")
    endif()
endforeach()
`, api.sourceDir, api.buildDir)

	if err := ioutil.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		return nil, fmt.Errorf("failed to create configure file extraction script: %w", err)
	}

	// Run cmake with our custom script
	args := []string{"-P", scriptPath}
	cmd := exec.Command(api.cmakeExe, args...)
	cmd.Dir = api.buildDir

	if err := cmd.Run(); err != nil {
		log.Printf("Warning: failed to extract configure_file commands: %v", err)
		return nil, err
	}

	// Read the configure_files.txt output
	outputFile := filepath.Join(api.buildDir, "configure_files.txt")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		return []*gazelle.CMakeConfigureFile{}, nil
	}

	content, err := ioutil.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read configure files output: %w", err)
	}

	// Parse the output
	var configureFiles []*gazelle.CMakeConfigureFile
	variables := make(map[string]string)
	
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}
		
		switch parts[0] {
		case "SET":
			if len(parts) >= 3 {
				variables[parts[1]] = parts[2]
			}
		case "CONFIGURE_FILE":
			if len(parts) >= 3 {
				inputFile := parts[1]
				outputFile := parts[2]
				
				// Generate rule name
				ruleName := strings.ReplaceAll(strings.ReplaceAll(outputFile, ".", "_"), "/", "_")
				if ruleName == "" {
					ruleName = "config_file"
				}
				
				// Combine detected variables with defines
				configVars := make(map[string]string)
				for k, v := range variables {
					configVars[k] = v
				}
				for k, v := range api.cmakeDefines {
					configVars[k] = v
				}
				
				// Add standard CMake variables
				if _, exists := configVars["CMAKE_CURRENT_SOURCE_DIR"]; !exists {
					configVars["CMAKE_CURRENT_SOURCE_DIR"] = "."
				}
				if _, exists := configVars["PROJECT_NAME"]; !exists {
					configVars["PROJECT_NAME"] = "project"
				}
				
				configureFiles = append(configureFiles, &gazelle.CMakeConfigureFile{
					Name:       ruleName,
					InputFile:  inputFile,
					OutputFile: outputFile,
					Variables:  configVars,
				})
			}
		}
	}
	
	return configureFiles, nil
}

// Configure runs CMake configure to generate the File API response
func (api *CMakeFileAPI) Configure() error {
	// Ensure build directory exists
	if err := os.MkdirAll(api.buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}

	// Build cmake command with -D flags for defines
	args := []string{}
	for key, value := range api.cmakeDefines {
		args = append(args, fmt.Sprintf("-D%s=%s", key, value))
	}
	args = append(args, api.sourceDir)

	// Run cmake configure
	cmd := exec.Command(api.cmakeExe, args...)
	cmd.Dir = api.buildDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("Running CMake configure: %s %v (in %s)", api.cmakeExe, args, api.buildDir)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cmake configure failed: %w", err)
	}

	return nil
}

// ReadAPIResponse reads and parses the CMake File API response
func (api *CMakeFileAPI) ReadAPIResponse() (*APIIndex, *Codemodel, map[string]*Target, error) {
	replyDir := filepath.Join(api.buildDir, ".cmake", "api", "v1", "reply")
	
	// Find the index file
	indexPattern := filepath.Join(replyDir, "index-*.json")
	indexFiles, err := filepath.Glob(indexPattern)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to find index files: %w", err)
	}
	if len(indexFiles) == 0 {
		return nil, nil, nil, fmt.Errorf("no index files found in %s", replyDir)
	}

	// Read the most recent index file (they're timestamped)
	indexFile := indexFiles[len(indexFiles)-1]
	indexData, err := ioutil.ReadFile(indexFile)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read index file: %w", err)
	}

	var index APIIndex
	if err := json.Unmarshal(indexData, &index); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse index file: %w", err)
	}

	// Find codemodel in objects
	var codemodelJSONFile string
	for _, obj := range index.Objects {
		if obj.Kind == "codemodel" {
			codemodelJSONFile = obj.JSONFile
			break
		}
	}
	if codemodelJSONFile == "" {
		return nil, nil, nil, fmt.Errorf("no codemodel found in index")
	}

	codemodelFile := filepath.Join(replyDir, codemodelJSONFile)
	codemodelData, err := ioutil.ReadFile(codemodelFile)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read codemodel file: %w", err)
	}

	var codemodel Codemodel
	if err := json.Unmarshal(codemodelData, &codemodel); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse codemodel file: %w", err)
	}

	// Read all targets
	targets := make(map[string]*Target)
	if len(codemodel.Configurations) > 0 {
		config := codemodel.Configurations[0] // Use first configuration
		for _, targetRef := range config.Targets {
			targetFile := filepath.Join(replyDir, targetRef.JSONFile)
			targetData, err := ioutil.ReadFile(targetFile)
			if err != nil {
				log.Printf("Warning: failed to read target file %s: %v", targetFile, err)
				continue
			}

			var target Target
			if err := json.Unmarshal(targetData, &target); err != nil {
				log.Printf("Warning: failed to parse target file %s: %v", targetFile, err)
				// Log the first part of the JSON for debugging
				debugData := string(targetData)
				if len(debugData) > 200 {
					debugData = debugData[:200] + "..."
				}
				log.Printf("Debug: First 200 chars of problematic JSON: %s", debugData)
				continue
			}

			targets[target.ID] = &target
		}
	}

	return &index, &codemodel, targets, nil
}

// GenerateFromAPI generates Bazel rules using CMake File API
func (api *CMakeFileAPI) GenerateFromAPI(relativeDir string) ([]*gazelle.CMakeTarget, error) {
	// Create query files
	if err := api.CreateQuery(); err != nil {
		return nil, fmt.Errorf("failed to create API query: %w", err)
	}

	// Run CMake configure
	if err := api.Configure(); err != nil {
		return nil, fmt.Errorf("failed to configure CMake: %w", err)
	}

	// Read API response
	_, _, targets, err := api.ReadAPIResponse()
	if err != nil {
		return nil, fmt.Errorf("failed to read API response: %w", err)
	}

	// Convert targets to CMakeTarget format
	var cmakeTargets []*gazelle.CMakeTarget

	for _, target := range targets {
		// Skip utility targets and imported targets
		if target.Type == "UTILITY" || strings.HasPrefix(target.Type, "INTERFACE") {
			continue
		}

		cmakeTarget := &gazelle.CMakeTarget{
			Name: target.Name,
		}

		// Map CMake target type to our type
		switch target.Type {
		case "STATIC_LIBRARY", "SHARED_LIBRARY", "MODULE_LIBRARY", "OBJECT_LIBRARY":
			cmakeTarget.Type = "library"
		case "EXECUTABLE":
			cmakeTarget.Type = "executable"
		default:
			log.Printf("Unknown target type %s for target %s, skipping", target.Type, target.Name)
			continue
		}

		// Extract sources and headers
		for _, source := range target.Sources {
			// Make path relative to the source directory if it's absolute
			sourcePath := source.Path
			if filepath.IsAbs(sourcePath) {
				if relPath, err := filepath.Rel(api.sourceDir, sourcePath); err == nil {
					sourcePath = relPath
				}
			}

			// Only include files that are in the current directory or subdirectories
			if !strings.HasPrefix(sourcePath, "..") {
				if isHeaderFile(sourcePath) {
					cmakeTarget.Headers = appendIfMissing(cmakeTarget.Headers, sourcePath)
				} else if isSourceFile(sourcePath) {
					cmakeTarget.Sources = appendIfMissing(cmakeTarget.Sources, sourcePath)
				}
			}
		}

		// Extract include directories
		includeDirectories := extractIncludeDirectories(target, api.sourceDir)
		cmakeTarget.IncludeDirectories = append(cmakeTarget.IncludeDirectories, includeDirectories...)

		// Extract linked libraries from dependencies
		for _, dep := range target.Dependencies {
			if depTarget, exists := targets[dep.ID]; exists {
				cmakeTarget.LinkedLibraries = appendIfMissing(cmakeTarget.LinkedLibraries, depTarget.Name)
			}
		}

		// Extract linked libraries from link information
		if target.Link != nil {
			// Check both Libraries and CommandFragments with role "libraries"
			for _, lib := range target.Link.Libraries {
				// Parse library names from fragments
				libName := strings.TrimSpace(lib.Fragment)
				if libName != "" && !strings.HasPrefix(libName, "-") {
					cmakeTarget.LinkedLibraries = appendIfMissing(cmakeTarget.LinkedLibraries, libName)
				}
			}
			
			// Also check CommandFragments for libraries role
			for _, cmdFrag := range target.Link.CommandFragments {
				if cmdFrag.Role == "libraries" {
					libName := strings.TrimSpace(cmdFrag.Fragment)
					if libName != "" && !strings.HasPrefix(libName, "-") {
						// Remove lib prefix and .a/.so suffixes if present
						if strings.HasPrefix(libName, "lib") && (strings.HasSuffix(libName, ".a") || strings.HasSuffix(libName, ".so")) {
							libName = strings.TrimPrefix(libName, "lib")
							if strings.HasSuffix(libName, ".a") {
								libName = strings.TrimSuffix(libName, ".a")
							} else if strings.HasSuffix(libName, ".so") {
								libName = strings.TrimSuffix(libName, ".so")
							}
						}
						cmakeTarget.LinkedLibraries = appendIfMissing(cmakeTarget.LinkedLibraries, libName)
					}
				}
			}
		}

		cmakeTargets = append(cmakeTargets, cmakeTarget)
	}

	log.Printf("Generated %d targets from CMake File API for directory %s", len(cmakeTargets), relativeDir)
	return cmakeTargets, nil
}

// Helper functions

// extractIncludeDirectories safely extracts include directories from CompileGroups
func extractIncludeDirectories(target *Target, sourceDir string) []string {
	var includeDirectories []string
	
	if len(target.CompileGroups) == 0 {
		return includeDirectories
	}
	
	// Parse the CompileGroups JSON
	var compileGroups []struct {
		SourceIndexes []int `json:"sourceIndexes"`
		Language      string `json:"language"`
		Includes []struct {
			Path      string `json:"path"`
			IsSystem  bool   `json:"isSystem,omitempty"`
			Backtrace int    `json:"backtrace,omitempty"`
		} `json:"includes,omitempty"`
	}
	
	if err := json.Unmarshal(target.CompileGroups, &compileGroups); err != nil {
		log.Printf("Warning: failed to parse CompileGroups for target %s: %v", target.Name, err)
		return includeDirectories
	}
	
	// Extract includes from the first compile group
	if len(compileGroups) > 0 {
		for _, include := range compileGroups[0].Includes {
			includePath := include.Path
			if filepath.IsAbs(includePath) {
				if relPath, err := filepath.Rel(sourceDir, includePath); err == nil {
					includePath = relPath
				}
			}
			if !strings.HasPrefix(includePath, "..") && !include.IsSystem {
				includeDirectories = appendIfMissing(includeDirectories, includePath)
			}
		}
	}
	
	return includeDirectories
}

func appendIfMissing(slice []string, str string) []string {
	for _, s := range slice {
		if s == str {
			return slice
		}
	}
	return append(slice, str)
}

// Helper: Basic check for header file extensions
func isHeaderFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".h" || ext == ".hh" || ext == ".hpp" || ext == ".hxx"
}

// Helper: Basic check for C++ source file extensions
func isSourceFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".c" || ext == ".cc" || ext == ".cpp" || ext == ".cxx"
}