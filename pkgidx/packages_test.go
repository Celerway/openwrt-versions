package pkgidx

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	const sampleData = `
Package: arptables-nft
Version: 1.8.8-2
Depends: libc, kmod-nft-arp, xtables-nft, kmod-arptables
Provides: arptables
Alternatives: 300:/usr/sbin/arptables:/usr/sbin/xtables-nft-multi, 300:/usr/sbin/arptables-restore:/usr/sbin/xtables-nft-multi, 300:/usr/sbin/arptables-save:/usr/sbin/xtables-nft-multi
License: GPL-2.0
Section: net
CPE-ID: cpe:/a:netfilter:iptables
Architecture: x86_64
Installed-Size: 2336
Filename: arptables-nft_1.8.8-2_x86_64.ipk
Size: 3138
SHA256sum: 1ccb8c52d7ee035981a6743dd6312911b16f10502c1d3d6569fd855d46a09467
Description:  ARP firewall administration tool nft

Package: base-files
Version: 1562-r24106-10cc5fcd00
Depends: libc, netifd, jsonfilter, usign, openwrt-keyring, fstools, fwtool
License: GPL-2.0
Section: base
Architecture: x86_64
Installed-Size: 47357
Filename: base-files_1562-r24106-10cc5fcd00_x86_64.ipk
Size: 48353
SHA256sum: e57f1bcadbf00584cadef4c9a0a24025e159f45a93944ec222858c1027a1adde
Description:  This package contains a base filesystem and system scripts for OpenWrt.
`

	packages, err := Parse(strings.NewReader(sampleData))
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

func TestLoadFromManifest(t *testing.T) {
	const sampleData = `
6in4 - 28
6rd - 12
admb - 2017-02-01-1.4
`

	packages, err := LoadFromManifest(strings.NewReader(sampleData))
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

func TestLoadFromManifest_InvalidFormat(t *testing.T) {
	const invalidData = `
invalid-line
another-invalid-line
`

	_, err := LoadFromManifest(strings.NewReader(invalidData))
	if err == nil {
		t.Error("Expected an error for invalid line format")
	}
}
