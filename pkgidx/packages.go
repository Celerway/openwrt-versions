package pkgidx

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Package represents a single package entry.
type Package struct {
	Name    string
	Version string
}

type PackageList []Package

type Diff struct {
	Name              string
	UpstreamVersion   string
	DownstreamVersion string
}

// LoadFromURL loads package information from a given HTTP URL.
func LoadFromURL(ctx context.Context, url string) ([]Package, error) {
	// Create a new HTTP request, with a context that can be cancelled
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	// Send the HTTP request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http.DefaultClient.Do: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	parsed, err := Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Parse: %w", err)
	}
	return parsed, nil
}

// Parse parses package information from an io.Reader.
func Parse(r io.Reader) (PackageList, error) {
	var packages []Package
	var currentPackage *Package
	scanner := bufio.NewScanner(r)
	var entry []byte
	for scanner.Scan() {
		// look for the start of a new package entry:
		if strings.HasPrefix(scanner.Text(), "Package: ") {
			if len(entry) == 0 { // first package
				entry = append(entry, scanner.Bytes()...)
				entry = append(entry, '\n')
				continue
			}
			// this wasn't the first package, parse the current buffer and add it to the list
			currentPackage = parsePackage(entry)
			packages = append(packages, *currentPackage)
			entry = entry[:0] // reset entry buffer
		}
		// append the line to the current entry
		entry = append(entry, scanner.Bytes()...)
		entry = append(entry, '\n')
	}
	// parse the last package:
	if len(entry) > 0 {
		currentPackage = parsePackage(entry)
		packages = append(packages, *currentPackage)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan input: %w", err)
	}
	return packages, nil
}

func parsePackage(entry []byte) *Package {
	p := &Package{}
	scanner := bufio.NewScanner(strings.NewReader(string(entry)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Package: ") {
			p.Name = strings.TrimPrefix(line, "Package: ")
		} else if strings.HasPrefix(line, "Version: ") {
			p.Version = strings.TrimPrefix(line, "Version: ")
		}
	}
	return p
}

func LoadFromManifest(r io.Reader) (PackageList, error) {
	var packages []Package
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Split on " - "
		parts := strings.Split(line, " - ")
		// Split by whitespace
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line format: %s", line)
		}

		packages = append(packages, Package{
			Name:    parts[0],
			Version: parts[1],
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan input: %w", err)
	}

	return packages, nil
}

// Equals - compares two Package slices for equality wrt versions.
// you run this on the downstream with the upstream as the argument
func (a PackageList) Equals(b PackageList) []Diff {
	diff := make([]Diff, 0)
	for _, pkgA := range a {
		for _, pkgB := range b {
			// check if the package names are the same:
			if pkgA.Name == pkgB.Name {
				if pkgA.Version != pkgB.Version {
					diff = append(diff, Diff{
						Name:              pkgA.Name,
						DownstreamVersion: pkgA.Version,
						UpstreamVersion:   pkgB.Version,
					})
				}
			}
		}
	}
	return diff
}

// CompareOpenWrtVersions compares two OpenWrt package versions.
// Returns 1 if v1 > v2, -1 if v1 < v2, and 0 if v1 == v2.
func compareOpenWrtVersions(v1, v2 string) int {
	// simple case, string are equal:
	if v1 == v2 {
		return 0
	}

	// 1. Try to compare using the 1.2.3-r4 pattern
	re1 := regexp.MustCompile(`(\d+):?(\d+)\.(\d+)\.(\d+)-r?(\d*)`)
	match1 := re1.FindStringSubmatch(v1)
	match2 := re1.FindStringSubmatch(v2)

	if match1 != nil && match2 != nil {
		// Compare epochs if present
		if len(match1[1]) > 0 && len(match2[1]) > 0 {
			epoch1, _ := strconv.Atoi(match1[1])
			epoch2, _ := strconv.Atoi(match2[1])
			if epoch1 > epoch2 {
				return 1
			} else if epoch1 < epoch2 {
				return -1
			}
		}
		// Compare major, minor, patch
		for i := 2; i <= 4; i++ {
			num1, _ := strconv.Atoi(match1[i])
			num2, _ := strconv.Atoi(match2[i])
			if num1 > num2 {
				return 1
			} else if num1 < num2 {
				return -1
			}
		}
		// Compare revisions if present
		if len(match1[5]) > 0 && len(match2[5]) > 0 {
			rev1, _ := strconv.Atoi(match1[5])
			rev2, _ := strconv.Atoi(match2[5])
			if rev1 > rev2 {
				return 1
			} else if rev1 < rev2 {
				return -1
			}
		}
		return 0 // Versions are equal
	}

	// 2. Try to compare using the 2023-04-13-... pattern
	re2 := regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`)
	match1 = re2.FindStringSubmatch(v1)
	match2 = re2.FindStringSubmatch(v2)

	if match1 != nil && match2 != nil {
		date1, _ := time.Parse("2006-01-02", fmt.Sprintf("%s-%s-%s", match1[1], match1[2], match1[3]))
		date2, _ := time.Parse("2006-01-02", fmt.Sprintf("%s-%s-%s", match2[1], match2[2], match2[3]))
		if date1.After(date2) {
			return 1
		} else if date1.Before(date2) {
			return -1
		}
		return 0 // Dates are equal
	}

	// 3. Fallback to simple string comparison
	if v1 > v2 {
		return 1
	} else if v1 < v2 {
		return -1
	}

	return 0 // Versions are equal
}
