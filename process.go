package main

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
)

// isGoFile checks if the file is a go file & not a directory
func isGoFile(f os.FileInfo) bool {
	name := f.Name()
	return !f.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
}

// walkDir walks through a path, processing all go files recursively in a directory
func walkDir(path string) error {
	return filepath.Walk(
		path,
		func(path string, f os.FileInfo, err error) error {
			if err == nil && isGoFile(f) {
				_, err = processFile(path, nil, os.Stdout)
			}
			return err
		},
	)
}

// processFile reads a file and processes the content, then checks if they're equal.
func processFile(filename string, in io.Reader, out io.Writer) ([]byte, error) {
	log.Printf("processing %v\n", filename)

	if in == nil {
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer closeFile(f)
		in = f
	}

	src, err := io.ReadAll(in)
	if err != nil {
		return nil, err
	}

	res, err := process(src, filename)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(src, res) {
		// formatting has changed
		if *list {
			_, _ = fmt.Fprintln(out, string(res))
		}
		if *write {
			mode := os.FileMode(0644)
			if filename != "" {
				if info, statErr := os.Stat(filename); statErr == nil {
					mode = info.Mode().Perm()
				}
			}
			if err := os.WriteFile(filename, res, mode); err != nil {
				return nil, err
			}
		}
		if !*list && !*write {
			return res, nil
		}
	} else {
		log.Println("file has not been changed")
	}

	return res, nil
}

// closeFile tries to close a File and prints an error when it can't
func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		log.Println("could not close file")
	}
}

// process processes the source of a file, categorising the imports
// filePath is used to detect the local module path for the file
func process(src []byte, filePath string) (output []byte, err error) {
	var (
		fileSet          = token.NewFileSet()
		convertedImports *impManager
		node             *dst.File
	)

	node, err = decorator.ParseFile(fileSet, "", src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Determine local prefix for this file
	fileLocalPrefix := *localPrefix
	if fileLocalPrefix == "" && filePath != "" {
		// Auto-detect module path from file location
		fileLocalPrefix = findModulePath(filePath)
	}

	convertedImports, err = convertImportsToSlice(node, fileLocalPrefix)
	if err != nil {
		return nil, err
	}
	if convertedImports.countImports() == 0 {
		return src, nil
	}

	convertedImports.sortImports()
	convertedToGo := convertedImports.convertImportsToGo()
	output, err = replaceImports(convertedToGo, node)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// replaceImports replaces existing imports and handles multiple import statements
func replaceImports(newImports []byte, node *dst.File) ([]byte, error) {
	var (
		output []byte
		err    error
		buf    bytes.Buffer
	)

	// remove + update
	dstutil.Apply(node, func(cr *dstutil.Cursor) bool {
		n := cr.Node()

		if decl, ok := n.(*dst.GenDecl); ok && decl.Tok == token.IMPORT {
			cr.Delete()
		}

		return true
	}, nil)

	if err = decorator.Fprint(&buf, node); err != nil {
		return nil, err
	}

	packageName := node.Name.Name
	output = bytes.Replace(buf.Bytes(), []byte("package "+packageName), append([]byte("package "+packageName+"\n\n"), newImports...), 1)

	return output, nil
}

// convertImportsToSlice parses the file with AST and gets all imports
// localPrefix is the module prefix to identify local packages
func convertImportsToSlice(node *dst.File, localPrefix string) (*impManager, error) {
	importCategories := newImpManager()

	for _, importSpec := range node.Imports {
		impName := importSpec.Path.Value
		impNameWithoutQuotes := strings.Trim(impName, "\"")
		locName := importSpec.Name

		var locImpModel impModel
		if locName != nil {
			locImpModel.localReference = locName.Name
		}
		locImpModel.path = impName

		if localPrefix != "" && isLocalPackageWithPrefix(impName, localPrefix) {
			var group = importCategories.Local()
			group.append(&locImpModel)
		} else if isStandardPackage(impNameWithoutQuotes) {
			var group = importCategories.Standard()
			group.append(&locImpModel)
		} else if isSecondPackage(impNameWithoutQuotes) {
			var group = importCategories.SecondPart()
			group.append(&locImpModel)
		} else {
			var group = importCategories.ThirdPart()
			group.append(&locImpModel)
		}
	}

	return importCategories, nil
}
