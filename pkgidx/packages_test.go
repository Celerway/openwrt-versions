package pkgidx

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

//go:embed testdata/valid.Packages
var validPackages []byte

func TestParse(t *testing.T) {
	packages, err := Parse(bytes.NewReader(validPackages))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(packages) != 3 {
		t.Errorf("Expected 2 packages, got %d", len(packages))
	}

	// Add assertions to check specific fields of the parsed packages
	// For example:

	// Check the fields of the first package in the map
	if _, ok := packages["arptables-nft"]; !ok {
		t.Errorf("Expected package name 'arptables-nft' to be in packages")
	}

	if v, ok := packages["base-files"]; ok && v != "1562-r24106-10cc5fcd00" {
		t.Errorf("Expected package version '1562-r24106-10cc5fcd00', got '%s'", v)
	}

	// ... add more assertions as needed
}

func TestLoadFromURL(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `Package: test-package
Version: 1.0.0
Description: A test package`)
	}))
	defer ts.Close()

	// Test LoadFromURL with the test server's URL
	packages, err := LoadFromURL(context.TODO(), ts.URL)
	if err != nil {
		t.Fatalf("LoadFromURL failed: %v", err)
	}

	if len(packages) != 1 {
		t.Fatalf("Expected 1 package, got %d", len(packages))
	}

	if _, ok := packages["test-package"]; !ok {
		t.Errorf("Expected package name 'test-package' to be in packages")
	}
}

func TestLoadFromURL_Error(t *testing.T) {
	// Test with an invalid URL
	_, err := LoadFromURL(context.TODO(), "invalid-url")
	if err == nil {
		t.Error("Expected an error for an invalid URL")
	}

	// Create a test server that returns an error status code
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	// Test with the error-returning server
	_, err = LoadFromURL(context.TODO(), ts.URL)
	if err == nil {
		t.Error("Expected an error for a non-200 status code")
	}
}

//go:embed testdata/valid.manifest
var validManifest []byte

func TestLoadFromManifest(t *testing.T) {

	packages, err := LoadFromManifest(bytes.NewReader(validManifest))
	if err != nil {
		t.Fatalf("LoadFromManifest failed: %v", err)
	}

	if len(packages) != 127 {
		t.Errorf("Expected 127 packages, got %d", len(packages))
	}

	// Check the fields of the first package
	if _, ok := packages["base-files"]; !ok {
		t.Errorf("Expected package name 'base-files' to be in packages")
	}

	if v, ok := packages["base-files"]; ok && v != "1562-r24106-10cc5fcd00" {
		t.Errorf("Expected package version '1562-r24106-10cc5fcd00', got '%s'", v)
	}

	// Add similar assertions for other packages as needed
}

//go:embed testdata/invalid.0.manifest
var invalidManifest0 []byte

//go:embed testdata/invalid.1.manifest
var invalidManifest1 []byte

//go:embed testdata/invalid.2.manifest
var invalidManifest2 []byte

func TestLoadFromManifest_InvalidFormat(t *testing.T) {
	tests := []struct {
		name     string
		manifest []byte
	}{
		{"invalid.0", invalidManifest0},
		{"invalid.1", invalidManifest1},
		{"invalid.2", invalidManifest2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := LoadFromManifest(bytes.NewReader(tt.manifest))
			if err == nil {
				t.Error("Expected an error for invalid line format")
			}
			if res != nil {
				t.Error("Expected nil result for invalid manifest")
			}
		})
	}
}

func TestCompareOpenWrtVersions(t *testing.T) {
	var testCases = []struct {
		v1       string
		v2       string
		expected int
	}{
		// Basic comparisons
		{"1.0.0-r1", "1.0.0-r2", -1}, // v2 is newer (higher revision)
		{"1.0.0-r1", "1.0.1-r0", -1}, // v2 is newer (higher version)
		{"1.0.1-r0", "1.0.0-r2", 1},  // v1 is newer (higher version)
		{"1.0.0-r1", "1.0.0-r1", 0},  // Equal

		// Epoch comparisons
		{"1:1.0.0-r1", "2:1.0.0-r0", -1}, // v2 is newer (higher epoch)
		{"2:1.0.0-r0", "1:1.0.0-r1", 1},  // v1 is newer (higher epoch)

		// Edge cases
		{"1.0.0", "1.0.0-r1", -1},     // v2 is newer (has revision)
		{"1.0.0-r1", "1.0.0", 1},      // v1 is newer (has revision)
		{"1.0.0-r1", "1.0.0-r10", -1}, // v2 is newer (higher revision)

		// weird stuff:
		{"a", "b", -1}, // v2 is newer (higher letter)
		{"1.2.3.4.5.6.7.8.9.10", "1.2.3.4.5.6.7.8.9.20", -1},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s vs %s", tc.v1, tc.v2), func(t *testing.T) {
			actual := compareOpenWrtVersions(tc.v1, tc.v2)
			if actual != tc.expected {
				t.Errorf("Expected %d, got %d", tc.expected, actual)
			}
		})
	}
}

//go:embed testdata/23.05.5.base.Packages
var basePackages []byte

//go:embed testdata/23.05.5.addons.Packages
var addonsPackages []byte

//go:embed testdata/release_1.0.0.manifest
var releaseManifest []byte

func TestDiffing(t *testing.T) {
	manifest := bytes.NewReader(releaseManifest)
	upstream, err := Parse(bytes.NewReader(basePackages))
	if err != nil {
		t.Fatalf("Parse failed(base): %v", err)
	}
	fmt.Printf("Loaded %d upstream base packages\n", len(upstream))
	addons, err := Parse(bytes.NewReader(addonsPackages))
	if err != nil {
		t.Fatalf("Parse failed(addons): %v", err)
	}
	fmt.Printf("Loaded %d upstream addon packages\n", len(addons))

	// add the addons to the upstream packages:
	mergedPackages := make(map[string]string)
	for name, pkg := range upstream {
		mergedPackages[name] = pkg
	}
	for name, pkg := range addons {
		mergedPackages[name] = pkg
	}
	upstream = mergedPackages

	// Load the downstream manifest
	downstream, err := LoadFromManifest(manifest)
	if err != nil {
		t.Fatalf("LoadFromManifest failed: %v", err)
	}
	fmt.Printf("Loaded %d downstream packages\n", len(downstream))

	// Perform the diff
	delta := downstream.Equals(upstream)
	fmt.Printf("Found %d differences\n", len(delta))

	// Check the diff results
	for _, diff := range delta {
		t.Logf("Package %s: upstream=%s, downstream=%s", diff.Name, diff.UpstreamVersion, diff.DownstreamVersion)
	}

}
