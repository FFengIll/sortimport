package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// getModuleName parses the GOMOD name
func getModuleName() string {
	root, err := os.Getwd()
	if err != nil {
		log.Println("error when getting root path: ", err)
		return ""
	}

	goModBytes, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		log.Println("error when reading mod file: ", err)
		return ""
	}

	modName := modfile.ModulePath(goModBytes)

	return modName
}

// findModulePath searches for go.mod starting from the given path,
// traversing up the directory tree until found or reaching the root.
// Returns the module path from go.mod, or empty string if not found.
func findModulePath(startPath string) string {
	// Get absolute path
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		log.Println("error when getting absolute path: ", err)
		return ""
	}

	// If it's a file, start from its directory
	info, err := os.Stat(absPath)
	if err == nil && !info.IsDir() {
		absPath = filepath.Dir(absPath)
	}

	// Traverse up the directory tree
	currentPath := absPath
	for {
		goModPath := filepath.Join(currentPath, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Found go.mod, parse it
			goModBytes, err := os.ReadFile(goModPath)
			if err != nil {
				log.Println("error when reading mod file: ", err)
				return ""
			}
			modName := modfile.ModulePath(goModBytes)
			log.Printf("found module %s from %s\n", modName, goModPath)
			return modName
		}

		// Move up one directory
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached root, no go.mod found
			log.Println("no go.mod found in directory tree")
			return ""
		}
		currentPath = parentPath
	}
}

// isLocalPackageWithPrefix checks if the import is a local package using the given prefix
func isLocalPackageWithPrefix(impName string, prefix string) bool {
	if prefix == "" {
		return false
	}
	// name with " or not
	if strings.HasPrefix(impName, prefix) || strings.HasPrefix(impName, "\""+prefix) {
		return true
	}
	return false
}

func isSecondPackage(impName string) bool {
	if *secondPrefix != "" {
		// name with " or not
		if strings.HasPrefix(impName, *secondPrefix) || strings.HasPrefix(impName, "\""+*secondPrefix) {
			return true
		}
	}
	return false
}
