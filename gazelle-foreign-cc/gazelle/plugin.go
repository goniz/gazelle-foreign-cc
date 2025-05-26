package gazelle

import (
	"log"

	"example.com/gazelle-foreign-cc/language" // Adjust the import path if your go.mod module name is different
	"github.com/bazelbuild/bazel-gazelle/language"
)

func init() {
	if err := language.RegisterLanguage(language.NewLanguage()); err != nil {
		log.Fatalf("Failed to register CMake language plugin: %v", err)
	}
}
