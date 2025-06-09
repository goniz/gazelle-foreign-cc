package gazelle

import (
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/goniz/gazelle-foreign-cc/common"
)

// GenerateRules delegates to the common package to avoid circular dependencies
func GenerateRules(args language.GenerateArgs) language.GenerateResult {
	return common.GenerateRules(args)
}
