package language

import (
	"path/filepath"
	"strings"
)

// appendIfMissing adds a string to a slice if it's not already present
func appendIfMissing(slice []string, str string) []string {
	for _, s := range slice {
		if s == str {
			return slice
		}
	}
	return append(slice, str)
}

// isHeaderFile checks if a file has a header extension
func isHeaderFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".h" || ext == ".hh" || ext == ".hpp" || ext == ".hxx"
}

// isSourceFile checks if a file has a C/C++ source extension
func isSourceFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".c" || ext == ".cc" || ext == ".cpp" || ext == ".cxx"
}

// fileExists checks if a file exists in a list of files
func fileExists(file string, fileList []string) bool {
	for _, f := range fileList {
		if f == file {
			return true
		}
	}
	return false
}