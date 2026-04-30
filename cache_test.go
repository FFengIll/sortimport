package main

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func Test_loadStandardPackages(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "load",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := loadStandardPackages(); (err != nil) != tt.wantErr {
				t.Errorf("loadStandardPackages() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCacheManager_New(t *testing.T) {
	cm, err := newCacheManager()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if cm == nil {
		t.Error("expected non-nil cache manager")
	}
	if cm.version == "" {
		t.Error("expected non-empty version")
	}
	if !strings.Contains(cm.cacheDir, ".cache/sortimport") {
		t.Errorf("expected cacheDir to contain .cache/sortimport, got: %s", cm.cacheDir)
	}
}

func TestCacheManager_GetCacheFile(t *testing.T) {
	cm, err := newCacheManager()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	cacheFile := cm.getCacheFile()
	if !strings.Contains(cacheFile, ".cache/sortimport") {
		t.Errorf("expected cacheFile to contain .cache/sortimport, got: %s", cacheFile)
	}
	if !strings.Contains(cacheFile, cm.version) {
		t.Errorf("expected cacheFile to contain version %s, got: %s", cm.version, cacheFile)
	}
	if !strings.HasSuffix(cacheFile, ".json") {
		t.Errorf("expected cacheFile to end with .json, got: %s", cacheFile)
	}
}

func TestCacheManager_WriteAndRead(t *testing.T) {
	// Create a temp directory for testing
	tmpDir := t.TempDir()
	cm := &CacheManager{
		cacheDir: tmpDir,
		version:  "go1.21.0",
	}

	// Test data
	testPackages := map[string]struct{}{
		"fmt":     {},
		"os":      {},
		"strings": {},
	}

	// Write cache
	err := cm.write(testPackages)
	if err != nil {
		t.Errorf("expected no error on write, got: %v", err)
	}

	// Verify file exists
	cacheFile := cm.getCacheFile()
	_, err = os.Stat(cacheFile)
	if err != nil {
		t.Errorf("expected cache file to exist, got error: %v", err)
	}

	// Read cache
	info, err := cm.read()
	if err != nil {
		t.Errorf("expected no error on read, got: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil cache info")
	}
	if info.Version != "go1.21.0" {
		t.Errorf("expected version go1.21.0, got: %s", info.Version)
	}
	if _, ok := info.Data["fmt"]; !ok {
		t.Error("expected fmt in cache data")
	}
	if _, ok := info.Data["os"]; !ok {
		t.Error("expected os in cache data")
	}
	if _, ok := info.Data["strings"]; !ok {
		t.Error("expected strings in cache data")
	}
}

func TestCacheManager_ReadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	cm := &CacheManager{
		cacheDir: tmpDir,
		version:  "go1.99.0", // Non-existent version
	}

	info, err := cm.read()
	if err == nil {
		t.Error("expected error for non-existent cache")
	}
	if info != nil {
		t.Errorf("expected nil info, got: %v", info)
	}
}

func TestCacheManager_VersionIndependent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create cache manager for go1.21.0
	cm1 := &CacheManager{
		cacheDir: tmpDir,
		version:  "go1.21.0",
	}
	packages1 := map[string]struct{}{"fmt": {}}
	err := cm1.write(packages1)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Create cache manager for go1.22.0
	cm2 := &CacheManager{
		cacheDir: tmpDir,
		version:  "go1.22.0",
	}
	packages2 := map[string]struct{}{"os": {}, "io": {}}
	err = cm2.write(packages2)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Both cache files should exist
	if _, err := os.Stat(cm1.getCacheFile()); os.IsNotExist(err) {
		t.Error("expected cm1 cache file to exist")
	}
	if _, err := os.Stat(cm2.getCacheFile()); os.IsNotExist(err) {
		t.Error("expected cm2 cache file to exist")
	}

	// Read both and verify they are independent
	info1, err := cm1.read()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if _, ok := info1.Data["fmt"]; !ok {
		t.Error("expected fmt in info1.Data")
	}
	if _, ok := info1.Data["os"]; ok {
		t.Error("did not expect os in info1.Data")
	}

	info2, err := cm2.read()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if _, ok := info2.Data["os"]; !ok {
		t.Error("expected os in info2.Data")
	}
	if _, ok := info2.Data["io"]; !ok {
		t.Error("expected io in info2.Data")
	}
	if _, ok := info2.Data["fmt"]; ok {
		t.Error("did not expect fmt in info2.Data")
	}
}

func TestCurrentGoVersionCache(t *testing.T) {
	cm, err := newCacheManager()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// The cache file should include the current Go version
	expectedVersion := runtime.Version()
	if cm.version != expectedVersion {
		t.Errorf("expected version %s, got: %s", expectedVersion, cm.version)
	}

	cacheFile := cm.getCacheFile()
	if !strings.Contains(cacheFile, expectedVersion) {
		t.Errorf("expected cacheFile to contain %s, got: %s", expectedVersion, cacheFile)
	}
}

func TestCacheManager_Update(t *testing.T) {
	tmp := t.TempDir()
	cm := &CacheManager{cacheDir: tmp, version: runtime.Version()}

	if err := cm.update(); err != nil {
		t.Fatalf("update: %v", err)
	}

	cacheFile := cm.getCacheFile()
	if _, err := os.Stat(cacheFile); err != nil {
		t.Fatalf("expected cache file to exist: %v", err)
	}

	info, err := cm.read()
	if err != nil {
		t.Fatalf("read after update: %v", err)
	}
	if info.Version != runtime.Version() {
		t.Errorf("version = %q, want %q", info.Version, runtime.Version())
	}
	if _, ok := info.Data["fmt"]; !ok {
		t.Error("expected fmt in updated cache data")
	}
}

func TestCacheManager_LoadOrFetch_Hit(t *testing.T) {
	tmp := t.TempDir()
	cm := &CacheManager{cacheDir: tmp, version: "go-test-fake"}

	seed := map[string]struct{}{"my/fake/pkg": {}}
	if err := cm.write(seed); err != nil {
		t.Fatalf("seed write: %v", err)
	}

	got, err := cm.loadOrFetch()
	if err != nil {
		t.Fatalf("loadOrFetch: %v", err)
	}
	if _, ok := got["my/fake/pkg"]; !ok {
		t.Error("expected seeded data to be returned (cache hit)")
	}
	// Real std package must NOT be in the seeded result.
	if _, ok := got["fmt"]; ok {
		t.Error("did not expect fmt in seeded cache hit result")
	}
}

func TestCacheManager_LoadOrFetch_Miss(t *testing.T) {
	tmp := t.TempDir()
	cm := &CacheManager{cacheDir: tmp, version: runtime.Version()}

	got, err := cm.loadOrFetch()
	if err != nil {
		t.Fatalf("loadOrFetch: %v", err)
	}
	if _, ok := got["fmt"]; !ok {
		t.Error("expected fetched std packages to include fmt")
	}
	// After miss, cache file should now exist for next call.
	if _, err := os.Stat(cm.getCacheFile()); err != nil {
		t.Errorf("expected cache file after miss: %v", err)
	}
}

func TestIsStandardPackage(t *testing.T) {
	if !isStandardPackage("fmt") {
		t.Error("fmt should be standard")
	}
	if !isStandardPackage("os") {
		t.Error("os should be standard")
	}
	if isStandardPackage("github.com/external/lib") {
		t.Error("third-party package should not be standard")
	}
	if isStandardPackage("") {
		t.Error("empty string should not be standard")
	}
}
