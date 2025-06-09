package common

// CMakeTarget represents a target defined in CMakeLists.txt
type CMakeTarget struct {
	Name               string
	Type               string // "library", "executable"
	Sources            []string
	Headers            []string // If explicitly listed or inferred
	IncludeDirectories []string
	LinkedLibraries    []string
}

// CMakeConfigureFile represents a configure_file command in CMakeLists.txt
type CMakeConfigureFile struct {
	Name       string            // Generated rule name
	InputFile  string            // Input template file
	OutputFile string            // Output configured file
	Variables  map[string]string // CMake variables for substitution
}