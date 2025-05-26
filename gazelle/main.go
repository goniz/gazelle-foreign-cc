package main

import (
	"fmt"
	"os"
	
	_ "example.com/gazelle-foreign-cc/language" // Import to trigger language registration
)

func main() {
	fmt.Println("CMake Gazelle plugin loaded")
	os.Exit(0)
}