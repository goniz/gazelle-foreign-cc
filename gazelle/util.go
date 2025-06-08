package gazelle

import (
	"path/filepath"
	"strings"
)

// AppendIfMissing adds a string to a slice if it's not already present
func AppendIfMissing(slice []string, str string) []string {
	for _, s := range slice {
		if s == str {
			return slice
		}
	}
	return append(slice, str)
}

// IsHeaderFile checks if a file has a header extension
func IsHeaderFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".h" || ext == ".hh" || ext == ".hpp" || ext == ".hxx"
}

// IsSourceFile checks if a file has a C/C++ source extension
func IsSourceFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".c" || ext == ".cc" || ext == ".cpp" || ext == ".cxx"
}

// FileExists checks if a file exists in a list of files
func FileExists(file string, fileList []string) bool {
	for _, f := range fileList {
		if f == file {
			return true
		}
	}
	return false
}