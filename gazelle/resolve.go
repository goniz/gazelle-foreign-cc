package gazelle // Ensure this is the correct package name

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/language" // For language.Language, if needed for cross-lang
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

var includeRegex = regexp.MustCompile(`^\s*#\s*include\s*([<"])([^>"]+)([>"])`)

// ResolveDeps analyzes the dependencies for a given rule.
func ResolveDeps(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, r *rule.Rule, lang language.Language, from label.Label) []resolve.FindResult {
	results := []resolve.FindResult{}
	// cfg := GetCMakeConfig(c) // Get CMake specific config if needed

	// --- 1. Resolve based on target_link_libraries (from CMake File API) ---
	linkedLibsAttr := r.PrivateAttr("cmake_linked_libraries")
	if linkedLibs, ok := linkedLibsAttr.([]string); ok && len(linkedLibs) > 0 {
		log.Printf("Rule %s (%s): Found linked libraries: %v", r.Name(), from.String(), linkedLibs)
		for _, libName := range linkedLibs {
			// Try to find by import spec for local libraries
			importSpec := resolve.ImportSpec{Lang: "cc", Imp: libName}
			findResults := ix.FindRulesByImport(importSpec, "cc")
			if len(findResults) > 0 {
				for _, findResult := range findResults {
					if findResult.Label.Name != "" && findResult.Label.Name != from.Name {
						results = append(results, findResult)
						log.Printf("Rule %s (%s): Resolved linked library %s to local target %s", r.Name(), from.String(), libName, findResult.Label.String())
					}
				}
			} else {
				// Try to find by import spec for external libraries
				importSpec := resolve.ImportSpec{Lang: "cc", Imp: libName}
				findResults := ix.FindRulesByImport(importSpec, "cc")
				if len(findResults) > 0 {
					for _, findResult := range findResults {
						if findResult.Label.Name != "" && findResult.Label.Name != from.Name {
							results = append(results, findResult)
							log.Printf("Rule %s (%s): Resolved linked library %s to external target %s", r.Name(), from.String(), libName, findResult.Label.String())
						}
					}
				} else {
					log.Printf("Rule %s (%s): Could not resolve linked library %s to any target.", r.Name(), from.String(), libName)
				}
			}
		}
	}

	// --- 2. Resolve based on #include directives ---
	// includeDirsAttr := r.PrivateAttr("cmake_include_directories") // Not used yet, but retrieved for future
	// cmakeIncludeDirs, _ := includeDirsAttr.([]string) // Error check if needed

	allFiles := append(r.AttrStrings("srcs"), r.AttrStrings("hdrs")...)
	if len(allFiles) == 0 {
		// If there are linked libraries but no source/header files in this rule,
		// we might still want to keep those dependencies.
		// The current logic returns early, but consider if linkedLibs deps should be kept.
		if len(results) > 0 {
			log.Printf("Rule %s (%s): No source/header files to parse for includes, but retaining linked library dependencies: %v", r.Name(), from.String(), results)
			// Deduplication will be handled at the end.
			return results
		}
		log.Printf("Rule %s (%s): No source/header files and no linked libraries to process.", r.Name(), from.String())
		return results // No files to parse for includes
	}

	// Get the directory of the current rule for resolving relative includes
	pkgDir := filepath.Join(c.RepoRoot, from.Pkg)


	for _, fileRelPath := range allFiles {
		absFilePath := filepath.Join(pkgDir, fileRelPath)
		if _, err := os.Stat(absFilePath); os.IsNotExist(err) {
			log.Printf("Rule %s (%s): Source/header file %s (abs: %s) not found for include parsing.", r.Name(), from.String(), fileRelPath, absFilePath)
			continue
		}

		file, err := os.Open(absFilePath)
		if err != nil {
			log.Printf("Rule %s (%s): Error opening file %s for include parsing: %v", r.Name(), from.String(), absFilePath, err)
			continue
		}
		// Ensure file is closed. Using defer inside loop is okay if number of files is not extremely large.
		// For very large number of files, manage f.Close() more carefully or open/close per scan.
		func() {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				matches := includeRegex.FindStringSubmatch(line)
				if len(matches) == 4 {
					// matches[1] is '<' or '"'
					// matches[2] is the include path
					// matches[3] is '>' or '"'
					includePath := matches[2]
					// isAngled := matches[1] == "<" // Not used yet, but could inform logic

					log.Printf("Rule %s (%s): Found include: %s%s%s in file %s", r.Name(), from.String(), matches[1], includePath, matches[3], fileRelPath)
					
					// Attempt to resolve this include path
					// For C++, the providing language is "cc". The consuming language is our plugin's name.
					res := ix.FindRulesByImport(resolve.ImportSpec{Lang: "cc", Imp: includePath}, lang.Name())
					if len(res) > 0 {
						for _, findResult := range res {
							if findResult.Label.Name != "" {
								// Avoid adding self-dependencies if the include resolves to the current rule
								if findResult.Label.Repo == from.Repo && findResult.Label.Pkg == from.Pkg && findResult.Label.Name == from.Name {
									log.Printf("Rule %s (%s): Ignoring self-dependency from include '%s' resolving to %s", r.Name(), from.String(), includePath, findResult.Label.String())
									continue
								}
								results = append(results, findResult)
								log.Printf("Rule %s (%s): Resolved include '%s' to %s", r.Name(), from.String(), includePath, findResult.Label.String())
							}
						}
					} else {
						log.Printf("Rule %s (%s): Could not resolve include '%s' using FindRulesByImport for lang 'cc'.", r.Name(), from.String(), includePath)
					}
				}
			}
			if err := scanner.Err(); err != nil {
				log.Printf("Rule %s (%s): Error scanning file %s for includes: %v", r.Name(), from.String(), absFilePath, err)
			}
		}() // Anonymous function call to manage defer file.Close() correctly per file
	}
	
	// Deduplicate results
	finalResults := []resolve.FindResult{}
	seen := make(map[label.Label]bool)
	for _, res := range results {
		if !seen[res.Label] {
			finalResults = append(finalResults, res)
			seen[res.Label] = true
		}
	}

	if len(finalResults) > 0 {
		log.Printf("Rule %s (%s): Final resolved dependencies: %v", r.Name(), from.String(), finalResults)
	} else {
		log.Printf("Rule %s (%s): No dependencies resolved.", r.Name(), from.String())
	}
	return finalResults
}
