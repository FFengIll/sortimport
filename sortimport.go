package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"path/filepath"
	"strings"
)

var (
	list             = flag.Bool("l", false, "write results to stdout")
	write            = flag.Bool("w", false, "write result to (source) file instead of stdout")
	localPrefix      = flag.String("local", "", "put imports beginning with this string after 3rd-party packages; comma-separated list")
	secondPrefix     = flag.String("second", "", "put imports beginning with this string after 3rd-party packages; comma-separated list")
	updateCache      = flag.Bool("u", false, "update the standard package cache for current Go version")
	verbose          bool // verbose logging
	standardPackages = make(map[string]struct{})
	cacheManager     *CacheManager
)

// main is the entry point of the program
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	err := goImportsSortMain()
	if err != nil {
		log.Fatalln(err)
	}
}

// goImportsSortMain checks passed flags and starts processing files
func goImportsSortMain() error {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "usage: goimportssort [flags] [path ...]\n")
		flag.PrintDefaults()
		os.Exit(2)
	}
	paths := parseFlags()

	if verbose {
		log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	} else {
		log.SetOutput(io.Discard)
	}

	// Initialize cache manager
	var err error
	cacheManager, err = newCacheManager()
	if err != nil {
		log.Printf("warning: failed to initialize cache manager: %v\n", err)
	}

	// Handle cache update flag
	if *updateCache {
		if cacheManager == nil {
			return errors.New("cache manager not available")
		}
		if err := cacheManager.update(); err != nil {
			return fmt.Errorf("failed to update cache: %w", err)
		}
		fmt.Printf("Cache updated for %s\n", cacheManager.version)
		return nil
	}

	if *localPrefix == "" {
		log.Println("no prefix found, using module name")

		moduleName := getModuleName()
		if moduleName != "" {
			localPrefix = &moduleName
		} else {
			log.Println("module name not found. skipping localprefix")
		}
	}

	if len(paths) == 0 {
		return errors.New("please enter a path to fix")
	}

	// load it in global
	if err := loadStandardPackages(); err != nil {
		return fmt.Errorf("failed to load standard packages: %w", err)
	}

	return processPaths(paths, os.Stdout)
}

// processPaths processes each path (file or directory) sequentially.
// It continues on error so a single bad file does not abort the batch,
// returning the first error encountered (if any).
// Go-style "..." patterns are accepted: "./...", "pkg/...", "..." are
// expanded to their containing directory and walked recursively.
func processPaths(paths []string, out io.Writer) error {
	var firstErr error
	for _, path := range paths {
		path = stripGoEllipsis(path)
		dir, statErr := os.Stat(path)
		if statErr != nil {
			if firstErr == nil {
				firstErr = statErr
			}
			log.Printf("error stating %s: %v\n", path, statErr)
			continue
		}
		if dir.IsDir() {
			if err := walkDir(path); err != nil && firstErr == nil {
				firstErr = err
			}
			continue
		}
		if _, err := processFile(path, nil, out); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// parseFlags parses command line flags and returns the paths to process.
// It's a var so that custom implementations can replace it in other files.
var parseFlags = func() []string {
	flag.BoolVar(&verbose, "v", false, "verbose logging")
	flag.Parse()

	return flag.Args()
}

// stripGoEllipsis converts Go-style "..." path patterns to their containing
// directory so that the existing recursive walker handles them. Mirrors the
// `cmd/go` convention familiar to Go users:
//
//	"./..."        -> "./"
//	"./pkg/..."    -> "./pkg"
//	"pkg/..."      -> "pkg"
//	"..."          -> "."
//	"/abs/p/..."   -> "/abs/p"
//
// Paths without a trailing "..." segment are returned unchanged.
func stripGoEllipsis(path string) string {
	if path == "..." {
		return "."
	}
	const slashEllipsis = "/..."
	if strings.HasSuffix(path, slashEllipsis) {
		stripped := strings.TrimSuffix(path, slashEllipsis)
		if stripped == "" {
			return "/"
		}
		return stripped
	}
	if filepath.Separator != '/' {
		sepEllipsis := string(filepath.Separator) + "..."
		if strings.HasSuffix(path, sepEllipsis) {
			stripped := strings.TrimSuffix(path, sepEllipsis)
			if stripped == "" {
				return string(filepath.Separator)
			}
			return stripped
		}
	}
	return path
}
