package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/tools/go/packages"
)

type PackageInfo struct {
	Data    map[string]struct{} `json:"data"`
	Version string              `json:"version"`
}

// CacheManager handles version-aware cache operations
type CacheManager struct {
	cacheDir string
	version  string
}

// newCacheManager creates a new CacheManager for the current Go version
func newCacheManager() (*CacheManager, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	cacheDir := filepath.Join(homedir, ".cache", "sortimport")
	version := runtime.Version()

	return &CacheManager{
		cacheDir: cacheDir,
		version:  version,
	}, nil
}

// getCacheFile returns the version-specific cache file path
func (c *CacheManager) getCacheFile() string {
	// Sanitize version for filename (replace spaces and special chars)
	safeVersion := strings.ReplaceAll(c.version, " ", "_")
	return filepath.Join(c.cacheDir, safeVersion+".json")
}

// read loads the cache for the current Go version
func (c *CacheManager) read() (*PackageInfo, error) {
	cacheFile := c.getCacheFile()

	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return nil, err
	}

	bs, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}

	var info PackageInfo
	if err := json.Unmarshal(bs, &info); err != nil {
		return nil, err
	}

	fmt.Printf("load standard package cache from %s\n", cacheFile)
	return &info, nil
}

// write saves the cache for the current Go version
func (c *CacheManager) write(packages map[string]struct{}) error {
	// Ensure cache directory exists
	if err := os.MkdirAll(c.cacheDir, 0755); err != nil {
		return err
	}

	cacheFile := c.getCacheFile()
	info := PackageInfo{
		Data:    make(map[string]struct{}),
		Version: c.version,
	}
	for k, v := range packages {
		info.Data[k] = v
	}

	bs, err := json.Marshal(info)
	if err != nil {
		return err
	}

	if err := os.WriteFile(cacheFile, bs, 0644); err != nil {
		return err
	}

	fmt.Printf("write standard package cache to %s\n", cacheFile)
	return nil
}

// update forces a cache refresh for the current Go version
func (c *CacheManager) update() error {
	pkgs, err := packages.Load(nil, "std")
	if err != nil {
		return err
	}

	packages := make(map[string]struct{})
	for _, p := range pkgs {
		packages[p.PkgPath] = struct{}{}
	}

	return c.write(packages)
}

// loadOrFetch loads from cache if available, otherwise fetches and caches
func (c *CacheManager) loadOrFetch() (map[string]struct{}, error) {
	// Try to read from cache first
	info, err := c.read()
	if err == nil && info != nil {
		return info.Data, nil
	}

	// Cache miss or error - fetch fresh data
	pkgs, err := packages.Load(nil, "std")
	if err != nil {
		return nil, err
	}

	packages := make(map[string]struct{})
	for _, p := range pkgs {
		packages[p.PkgPath] = struct{}{}
	}

	// Write to cache
	if err := c.write(packages); err != nil {
		log.Printf("warning: failed to write cache: %v", err)
	}

	return packages, nil
}

// loadStandardPackages tries to fetch all golang std packages
func loadStandardPackages() error {
	// Initialize cacheManager if not already done
	if cacheManager == nil {
		var err error
		cacheManager, err = newCacheManager()
		if err != nil {
			log.Printf("warning: failed to initialize cache manager: %v\n", err)
		}
	}

	// Use CacheManager if available
	if cacheManager != nil {
		pkgs, err := cacheManager.loadOrFetch()
		if err != nil {
			return err
		}
		for k, v := range pkgs {
			standardPackages[k] = v
		}
		return nil
	}

	// Fallback: load directly without cache
	pkgs, err := packages.Load(nil, "std")
	if err != nil {
		return err
	}
	for _, p := range pkgs {
		standardPackages[p.PkgPath] = struct{}{}
	}
	return nil
}

// isStandardPackage checks if a package string is included in the standardPackages map
func isStandardPackage(pkg string) bool {
	_, ok := standardPackages[pkg]
	return ok
}
