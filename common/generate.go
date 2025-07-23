package common

import (
	"log"

	"github.com/bazelbuild/bazel-gazelle/language"
)

// GenerateRules is deprecated - regex fallback parsing has been removed
// Always use CMake File API through the main language interface
func GenerateRules(args language.GenerateArgs) language.GenerateResult {
	log.Printf("GenerateRules in common package is deprecated. Regex fallback parsing has been removed. Use CMake File API instead.")
	return language.GenerateResult{}
}

// GenerateRulesWithDefines is deprecated - regex fallback parsing has been removed
// Always use CMake File API through the main language interface
func GenerateRulesWithDefines(args language.GenerateArgs, packageDefines map[string]string) language.GenerateResult {
	log.Printf("GenerateRulesWithDefines in common package is deprecated. Regex fallback parsing has been removed. Use CMake File API instead.")
	return language.GenerateResult{}
}
