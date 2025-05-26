package gazelle

import (
	// cmake "example.com/gazelle-foreign-cc/language" // Adjust the import path if your go.mod module name is different
	// "github.com/bazelbuild/bazel-gazelle/language"
)

func init() {
	// TODO: Fix language registration - RegisterLanguage doesn't exist in current Gazelle
	// Modern Gazelle uses a different registration mechanism
	// if err := language.RegisterLanguage(cmake.NewLanguage()); err != nil {
	// 	log.Fatalf("Failed to register CMake language plugin: %v", err)
	// }
}
