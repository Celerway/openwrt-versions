package pkgidx

import (
	"bytes"
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

	if len(packages) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(packages))
	}

	// Add assertions to check specific fields of the parsed packages
	// For example:
	if packages[0].Name != "arptables-nft" {
		t.Errorf("Expected package name 'arptables-nft', got '%s'", packages[0].Name)
	}
	if packages[1].Version != "1562-r24106-10cc5fcd00" {
		t.Errorf("Expected package version '1562-r24106-10cc5fcd00', got '%s'", packages[1].Version)
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
	packages, err := LoadFromURL(ts.URL)
	if err != nil {
		t.Fatalf("LoadFromURL failed: %v", err)
	}

	if len(packages) != 1 {
		t.Errorf("Expected 1 package, got %d", len(packages))
	}

	if packages[0].Name != "test-package" {
		t.Errorf("Expected package name 'test-package', got '%s'", packages[0].Name)
	}
}

func TestLoadFromURL_Error(t *testing.T) {
	// Test with an invalid URL
	_, err := LoadFromURL("invalid-url")
	if err == nil {
		t.Error("Expected an error for an invalid URL")
	}

	// Create a test server that returns an error status code
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	// Test with the error-returning server
	_, err = LoadFromURL(ts.URL)
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

	if len(packages) != 3 {
		t.Errorf("Expected 3 packages, got %d", len(packages))
	}

	// Check the fields of the first package
	if packages[0].Name != "6in4" {
		t.Errorf("Expected package name '6in4', got '%s'", packages[0].Name)
	}
	if packages[0].Version != "28" {
		t.Errorf("Expected package version '28', got '%s'", packages[0].Version)
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
