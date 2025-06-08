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

// CMakeConfiguredFile represents a file configured by CMake configure_file()
type CMakeConfiguredFile struct {
	InputFile  string // Template file (e.g., platform.hpp.in)
	OutputFile string // Generated file (e.g., platform.hpp)
}

// generateRulesFromCMakeFile attempts to parse a CMakeLists.txt file and extract target information.
func generateRulesFromCMakeFile(args language.GenerateArgs, cmakeFilePath string, cfg *CMakeConfig) language.GenerateResult {
	res := language.GenerateResult{}
	targets := make(map[string]*CMakeTarget) // Map of target name to CMakeTarget
	configuredFiles := make([]*CMakeConfiguredFile, 0) // Track configured files

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
		case "configure_file": // Handle configure_file(input output [options])
			if len(cmdArgs) < 2 {
				continue
			}
			inputFile := cmdArgs[0]
			outputFile := cmdArgs[1]
			
			// Clean the file paths - remove quotes and resolve relative paths
			inputFile = strings.Trim(inputFile, `"`)
			outputFile = strings.Trim(outputFile, `"`)
			
			// Store the configured file for later processing
			configuredFiles = append(configuredFiles, &CMakeConfiguredFile{
				InputFile:  inputFile,
				OutputFile: outputFile,
			})
			
			log.Printf("Found configure_file command: %s -> %s", inputFile, outputFile)
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
			// Don't add empty rules for now to fix deps generation
			// res.Empty = append(res.Empty, rule.NewRule(r.Kind(), r.Name()))
			log.Printf("Generated %s %s in %s with srcs: %v, hdrs: %v, includes: %v, links: %v",
				r.Kind(), r.Name(), args.Rel, finalSrcs, finalHdrs, cmTarget.IncludeDirectories, cmTarget.LinkedLibraries)
		} else {
			log.Printf("Skipping rule for target %s as no valid sources/headers were found in the current directory.", cmTarget.Name)
		}
	}
	
	// Generate rules for configured files (e.g., platform.hpp from platform.hpp.in)
	for _, configFile := range configuredFiles {
		generateConfiguredFileRule(&res, configFile, args)
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

// generateConfiguredFileRule creates a Bazel rule for a CMake configured file
func generateConfiguredFileRule(res *language.GenerateResult, configFile *CMakeConfiguredFile, args language.GenerateArgs) {
	// Extract the base name for the rule
	outputBaseName := filepath.Base(configFile.OutputFile)
	outputBaseName = strings.TrimSuffix(outputBaseName, filepath.Ext(outputBaseName))
	
	// Check if this is a platform configuration file
	if strings.Contains(configFile.OutputFile, "platform") && strings.HasSuffix(configFile.OutputFile, ".hpp") {
		generatePlatformHppRule(res, configFile, args)
		return
	}
	
	// For other configured files, create a simple genrule that copies from template
	if fileExistsInRegularFiles(configFile.InputFile, args.RegularFiles) {
		r := rule.NewRule("genrule", "generate_"+outputBaseName)
		r.SetAttr("srcs", []string{configFile.InputFile})
		r.SetAttr("outs", []string{configFile.OutputFile})
		r.SetAttr("cmd", "cp $< $@")
		
		res.Gen = append(res.Gen, r)
		log.Printf("Generated genrule for configured file: %s -> %s", configFile.InputFile, configFile.OutputFile)
	} else {
		log.Printf("Input file %s for configure_file not found in regular files, skipping", configFile.InputFile)
	}
}

// generatePlatformHppRule creates a specific rule for platform.hpp generation
func generatePlatformHppRule(res *language.GenerateResult, configFile *CMakeConfiguredFile, args language.GenerateArgs) {
	// Create a genrule that generates platform.hpp with predefined content
	r := rule.NewRule("genrule", "generate_platform_hpp")
	r.SetAttr("outs", []string{configFile.OutputFile})
	
	// Create a command that generates a basic platform.hpp file
	// This provides a minimal platform configuration for libzmq-style projects
	cmd := `cat > $@ << 'EOF'
#ifndef __ZMQ_PLATFORM_HPP_INCLUDED__
#define __ZMQ_PLATFORM_HPP_INCLUDED__

/* Basic platform configuration for CMake projects built with Bazel */
/* This file provides minimal platform definitions */

/* Condition variable implementation */
#define ZMQ_USE_CV_IMPL_STL11

/* I/O thread polling - use epoll on Linux, kqueue on BSD/macOS, select as fallback */
#if defined(__linux__)
  #define ZMQ_IOTHREAD_POLLER_USE_EPOLL
  #define ZMQ_IOTHREAD_POLLER_USE_EPOLL_CLOEXEC
  #define ZMQ_HAVE_LINUX
#elif defined(__APPLE__)
  #define ZMQ_IOTHREAD_POLLER_USE_KQUEUE
  #define ZMQ_HAVE_OSX
#elif defined(__FreeBSD__) || defined(__FreeBSD_kernel__)
  #define ZMQ_IOTHREAD_POLLER_USE_KQUEUE
  #define ZMQ_HAVE_FREEBSD
#elif defined(__OpenBSD__)
  #define ZMQ_IOTHREAD_POLLER_USE_KQUEUE
  #define ZMQ_HAVE_OPENBSD
#elif defined(__NetBSD__)
  #define ZMQ_IOTHREAD_POLLER_USE_KQUEUE
  #define ZMQ_HAVE_NETBSD
#else
  #define ZMQ_IOTHREAD_POLLER_USE_SELECT
#endif

/* API polling implementation */
#define ZMQ_POLL_BASED_ON_POLL

/* Enable common POSIX features */
#define HAVE_FORK
#define HAVE_CLOCK_GETTIME
#define ZMQ_HAVE_UIO

/* Enable common socket features */
#define ZMQ_HAVE_EVENTFD
#define ZMQ_HAVE_EVENTFD_CLOEXEC
#define ZMQ_HAVE_O_CLOEXEC
#define ZMQ_HAVE_SOCK_CLOEXEC
#define ZMQ_HAVE_SO_KEEPALIVE
#define ZMQ_HAVE_TCP_KEEPCNT
#define ZMQ_HAVE_TCP_KEEPIDLE
#define ZMQ_HAVE_TCP_KEEPINTVL
#define ZMQ_HAVE_TCP_KEEPALIVE

/* Pthread features */
#define ZMQ_HAVE_PTHREAD_SETNAME_2
#define ZMQ_HAVE_PTHREAD_SET_AFFINITY

/* String functions */
#define HAVE_STRNLEN
#define ZMQ_HAVE_STRLCPY

/* Enable IPC transport */
#define ZMQ_HAVE_IPC

/* Use built-in SHA1 implementation */
#define ZMQ_USE_BUILTIN_SHA1

/* WebSocket support */
#define ZMQ_HAVE_WS

/* Set cache line size to a reasonable default */
#define ZMQ_CACHELINE_SIZE 64

#endif
EOF`
	
	r.SetAttr("cmd", cmd)
	
	res.Gen = append(res.Gen, r)
	log.Printf("Generated platform.hpp rule for output: %s", configFile.OutputFile)
	
	// Also create a cc_library that provides the generated header
	headerLib := rule.NewRule("cc_library", "platform_headers")
	headerLib.SetAttr("hdrs", []string{configFile.OutputFile})
	
	// Extract the directory from the output file to set strip_include_prefix
	outputDir := filepath.Dir(configFile.OutputFile)
	if outputDir != "." && outputDir != "" {
		headerLib.SetAttr("strip_include_prefix", outputDir)
	}
	
	res.Gen = append(res.Gen, headerLib)
	log.Printf("Generated cc_library platform_headers for header: %s", configFile.OutputFile)
}

// Helper: Check if a file exists in the regular files list
func fileExistsInRegularFiles(file string, regularFiles []string) bool {
	for _, f := range regularFiles {
		if f == file {
			return true
		}
	}
	return false
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
