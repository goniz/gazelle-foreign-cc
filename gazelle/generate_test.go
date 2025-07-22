package gazelle

import (
	"testing"
)

// Note: Tests for regex fallback parsing have been removed since that functionality
// has been deprecated. All CMake parsing now uses the CMake File API through the
// main language interface in language/cmake.go.
//
// The regex fallback functionality has been removed as per the requirement to
// "Always use Cmake FILE API and fail if that fails."

func TestDeprecatedFunctionality(t *testing.T) {
	// This test documents that the regex fallback functionality has been removed
	t.Log("Regex fallback parsing has been removed. Use CMake File API through the main language interface.")
}
