package gazelle

import (
	"log"

	"github.com/bazelbuild/bazel-gazelle/language"
)

// GenerateRules is deprecated - regex fallback parsing has been removed
// Always use CMake File API through the main language interface
func GenerateRules(args language.GenerateArgs) language.GenerateResult {
	log.Printf("GenerateRules in gazelle package is deprecated. Regex fallback parsing has been removed. Use CMake File API instead.")
	return language.GenerateResult{}
}
