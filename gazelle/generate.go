package gazelle

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

// CMakeTarget represents a target defined in CMakeLists.txt
type CMakeTarget struct {
	Name               string
	Type               string // "library", "executable"
	Sources            []string
	Headers            []string // If explicitly listed or inferred
	IncludeDirectories []string
	LinkedLibraries    []string
}

// generateRulesFromCMakeFile attempts to parse a CMakeLists.txt file and extract target information.
func generateRulesFromCMakeFile(args language.GenerateArgs, cmakeFilePath string, cfg *CMakeConfig) language.GenerateResult {
	res := language.GenerateResult{}
	targets := make(map[string]*CMakeTarget) // Map of target name to CMakeTarget

	file, err := os.Open(cmakeFilePath)
	if err != nil {
		log.Printf("Error opening CMakeLists.txt %s: %v", cmakeFilePath, err)
		return res
	}
	defer file.Close()

	log.Printf("Parsing CMakeLists.txt: %s (Rel: %s)", cmakeFilePath, args.Rel)

	// Simplified parsing logic focusing on key commands.
	// A real CMake parser is much more complex.
	// This version will still use regex for command extraction but be more stateful.

	// Regex to identify CMake commands (e.g., add_library(...))
	// Captures: 1=command, 2=arguments string
	commandRegex := regexp.MustCompile(`(?im)^\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\(([\s\S]*?)\)\s*$`)
	// Regex to split arguments (very basic, won't handle all CMake quoting/escaping)
	argSplitRegex := regexp.MustCompile(`\s+`) // Simple split by space

	scanner := bufio.NewScanner(file)
	var currentContent strings.Builder
	for scanner.Scan() {
		currentContent.WriteString(scanner.Text() + "\n")
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading CMakeLists.txt %s: %v", cmakeFilePath, err)
		return res
	}

	fileContent := currentContent.String()
	allCommands := commandRegex.FindAllStringSubmatch(fileContent, -1)

	for _, cmdMatch := range allCommands {
		commandName := strings.ToLower(cmdMatch[1])
		argsString := strings.TrimSpace(cmdMatch[2])

		// Split arguments. This is a major simplification.
		// CMake uses space, semicolon for lists, quotes, variable expansion etc.
		rawArgs := argSplitRegex.Split(argsString, -1)
		var cmdArgs []string
		for _, arg := range rawArgs {
			if arg != "" { // Remove empty strings resulting from multiple spaces
				cmdArgs = append(cmdArgs, strings.Trim(arg, `"`)) // Basic quote stripping
			}
		}

		if len(cmdArgs) == 0 {
			continue
		}
		targetName := cmdArgs[0] // First argument is usually the target name

		switch commandName {
		case "add_library":
			if len(cmdArgs) < 2 {
				continue
			}
			target, ok := targets[targetName]
			if !ok {
				target = &CMakeTarget{Name: targetName, Type: "library"}
				targets[targetName] = target
			}
			target.Type = "library"               // Ensure type is library
			for _, srcFile := range cmdArgs[1:] { // Simplification: assumes all following args are sources
				// Basic check for header/source, could be improved
				if isHeaderFile(srcFile) {
					target.Headers = appendIfMissing(target.Headers, srcFile)
				} else if isSourceFile(srcFile) {
					target.Sources = appendIfMissing(target.Sources, srcFile)
				}
			}
		case "add_executable":
			if len(cmdArgs) < 2 {
				continue
			}
			target, ok := targets[targetName]
			if !ok {
				target = &CMakeTarget{Name: targetName, Type: "executable"}
				targets[targetName] = target
			}
			target.Type = "executable" // Ensure type
			for _, srcFile := range cmdArgs[1:] {
				if isSourceFile(srcFile) { // Executables usually don't list headers here
					target.Sources = appendIfMissing(target.Sources, srcFile)
				} else if isHeaderFile(srcFile) { // Though sometimes they might
					target.Headers = appendIfMissing(target.Headers, srcFile)
				}
			}
		case "target_sources": // Assumes target_sources(target_name PRIVATE src1 src2 ...)
			if len(cmdArgs) < 3 {
				continue
			} // target_name, scope, src1
			targetNameFromArgs := cmdArgs[0] // target_sources's first arg is the target name
			target, ok := targets[targetNameFromArgs]
			if !ok {
				continue
			} // Target must exist
			// Skipping scope (PRIVATE/PUBLIC/INTERFACE) for simplicity for now
			for _, srcFile := range cmdArgs[2:] { // Source files start from the third argument
				if isHeaderFile(srcFile) {
					target.Headers = appendIfMissing(target.Headers, srcFile)
				} else if isSourceFile(srcFile) {
					target.Sources = appendIfMissing(target.Sources, srcFile)
				}
			}
		case "target_include_directories": // Assumes target_include_directories(target_name PRIVATE dir1 dir2 ...)
			if len(cmdArgs) < 3 {
				continue
			}
			targetNameFromArgs := cmdArgs[0]
			target, ok := targets[targetNameFromArgs]
			if !ok {
				continue
			}
			// Skipping scope for simplicity
			for _, inclDir := range cmdArgs[2:] {
				// Here, inclDir might be relative to CMakeLists.txt or absolute.
				// It could also be ${CMAKE_CURRENT_SOURCE_DIR} etc.
				// For now, store as is. Resolution to Bazel paths is complex.
				target.IncludeDirectories = appendIfMissing(target.IncludeDirectories, inclDir)
			}
		case "target_link_libraries": // Handle target_link_libraries(target_name [scope] lib1 lib2 ...)
			if len(cmdArgs) < 2 {
				continue
			}
			targetNameFromArgs := cmdArgs[0]
			target, ok := targets[targetNameFromArgs]
			if !ok {
				continue
			}
			
			// Handle both formats: 
			// target_link_libraries(target lib1 lib2) and 
			// target_link_libraries(target PRIVATE/PUBLIC/INTERFACE lib1 lib2)
			startIdx := 1
			if len(cmdArgs) >= 3 {
				scope := strings.ToUpper(cmdArgs[1])
				if scope == "PRIVATE" || scope == "PUBLIC" || scope == "INTERFACE" {
					startIdx = 2 // Skip the scope keyword
				}
			}
			
			for _, linkedLib := range cmdArgs[startIdx:] {
				target.LinkedLibraries = appendIfMissing(target.LinkedLibraries, linkedLib)
			}
		}
	}

	// Convert CMakeTargets to Gazelle rules
	for _, cmTarget := range targets {
		var r *rule.Rule
		if cmTarget.Type == "library" {
			r = rule.NewRule("cc_library", cmTarget.Name)
		} else if cmTarget.Type == "executable" {
			r = rule.NewRule("cc_binary", cmTarget.Name)
		} else {
			log.Printf("Unknown target type for %s: %s", cmTarget.Name, cmTarget.Type)
			continue
		}

		// Filter sources/headers against args.RegularFiles (files actually in this directory)
		// This is a major simplification. CMake sources can be in subdirs.
		var finalSrcs, finalHdrs []string
		for _, s := range cmTarget.Sources {
			if fileExists(s, args.RegularFiles) {
				finalSrcs = append(finalSrcs, s)
			} else {
				log.Printf("Source file %s for target %s not found in current directory's regular files, skipping.", s, cmTarget.Name)
			}
		}
		for _, h := range cmTarget.Headers {
			if fileExists(h, args.RegularFiles) {
				finalHdrs = append(finalHdrs, h)
			} else {
				log.Printf("Header file %s for target %s not found in current directory's regular files, skipping.", h, cmTarget.Name)
			}
		}

		if len(finalSrcs) > 0 {
			r.SetAttr("srcs", finalSrcs)
		}
		if len(finalHdrs) > 0 {
			r.SetAttr("hdrs", finalHdrs)
		}

		// 'copts' from include_directories (very simplified)
		// A proper version would convert paths and use "includes" or "-I" prefixes.
		if len(cmTarget.IncludeDirectories) > 0 {
			var copts []string
			for _, dir := range cmTarget.IncludeDirectories {
				// This is a placeholder. Actual mapping of CMake include dirs to Bazel copts/includes is complex.
				// e.g., if dir is relative "inc", it might become "-Iinc" or "includes = ["inc"]"
				// For now, just add a placeholder to show it was captured.
				copts = append(copts, "-I"+dir) // Highly simplified!
			}
			if len(copts) > 0 {
				// r.SetAttr("copts", copts) // Example, might need adjustment based on rules_cc behavior
			}
		}

		// Generate deps attribute for locally linked libraries
		var deps []string
		for _, linkedLib := range cmTarget.LinkedLibraries {
			// Check if the linked library matches another target in this directory
			if _, exists := targets[linkedLib]; exists {
				deps = append(deps, ":"+linkedLib) // Use Bazel label syntax for local targets
			}
		}
		if len(deps) > 0 {
			r.SetAttr("deps", deps)
		}

		// Store linked libraries for dependency resolution (external libraries, includes, etc.)
		if len(cmTarget.LinkedLibraries) > 0 {
			r.SetPrivateAttr("cmake_linked_libraries", cmTarget.LinkedLibraries)
		}
		if len(cmTarget.IncludeDirectories) > 0 {
			r.SetPrivateAttr("cmake_include_directories", cmTarget.IncludeDirectories)
		}

		if r.Attr("srcs") != nil || r.Attr("hdrs") != nil { // Only add rule if it has sources/headers
			res.Gen = append(res.Gen, r)
			res.Empty = append(res.Empty, rule.NewRule(r.Kind(), r.Name()))
			log.Printf("Generated %s %s in %s with srcs: %v, hdrs: %v, includes: %v, links: %v",
				r.Kind(), r.Name(), args.Rel, finalSrcs, finalHdrs, cmTarget.IncludeDirectories, cmTarget.LinkedLibraries)
		} else {
			log.Printf("Skipping rule for target %s as no valid sources/headers were found in the current directory.", cmTarget.Name)
		}
	}
	// Gazelle expects Imports to have the same length as Gen. Populate with nils for now.
	if len(res.Gen) > 0 && len(res.Imports) == 0 {
		res.Imports = make([]interface{}, len(res.Gen))
	}
	return res
}

// Helper: Check if a file exists in a list of files
func fileExists(file string, fileList []string) bool {
	for _, f := range fileList {
		if f == file {
			return true
		}
	}
	return false
}

// Helper: Append if string is not already in slice
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

// Main GenerateRules function (updated to use CMake File API)
func GenerateRules(args language.GenerateArgs) language.GenerateResult {
	cfg := GetCMakeConfig(args.Config)
	cmakeFilePath := filepath.Join(args.Dir, "CMakeLists.txt")

	if _, err := os.Stat(cmakeFilePath); os.IsNotExist(err) {
		log.Printf("No CMakeLists.txt found in %s (%s). Skipping directory.", args.Rel, cmakeFilePath)
		return language.GenerateResult{}
	}

	// Use regex-based parsing (fallback method)
	return generateRulesFromCMakeFile(args, cmakeFilePath, cfg)
}
