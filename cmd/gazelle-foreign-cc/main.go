package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/goniz/gazelle-foreign-cc/language"
)

func main() {
	var sourceDir = flag.String("source_dir", "", "path to CMake source directory")
	flag.Parse()

	if *sourceDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		*sourceDir = wd
	}

	log.Printf("Testing CMake File API integration in directory: %s", *sourceDir)

	// Check if CMakeLists.txt exists
	cmakeFile := filepath.Join(*sourceDir, "CMakeLists.txt")
	if _, err := os.Stat(cmakeFile); os.IsNotExist(err) {
		log.Printf("No CMakeLists.txt found in %s", *sourceDir)
		return
	}

	// Test the CMake File API
	buildDir := filepath.Join(*sourceDir, ".cmake-build")
	api := language.NewCMakeFileAPI(*sourceDir, buildDir, "cmake", make(map[string]string))

	targets, err := api.GenerateFromAPI("")
	if err != nil {
		log.Printf("CMake File API failed: %v", err)
		log.Println("The system would fall back to regex parsing in the actual Gazelle plugin.")
		return
	}

	log.Printf("Generated %d targets from CMake File API for directory", len(targets))
	log.Printf("Successfully parsed %d targets using CMake File API:", len(targets))
	for _, target := range targets {
		log.Printf("  Target: %s (type: %s)", target.Name, target.Type)
		log.Printf("    Sources: %v", target.Sources)
		log.Printf("    Headers: %v", target.Headers)
		log.Printf("    Include dirs: %v", target.IncludeDirectories)
		log.Printf("    Linked libs: %v", target.LinkedLibraries)
	}

	log.Println("CMake File API test completed successfully!")
}