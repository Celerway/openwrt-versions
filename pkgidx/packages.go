package pkgidx

import (
	"bufio"
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
	Name          string
	Version       string
	Depends       []string
	Provides      string
	Alternatives  []string
	License       string
	Section       string
	CPEID         string
	Architecture  string
	InstalledSize int
	Filename      string
	Size          int
	SHA256sum     string
	Description   string
}

type PackageList []Package

type Diff struct {
	Name              string
	UpstreamVersion   string
	DownstreamVersion string
}

// LoadFromURL loads package information from a given HTTP URL.
func LoadFromURL(url string) ([]Package, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}

	return Parse(resp.Body)
}

// Parse parses package information from an io.Reader.
func Parse(r io.Reader) ([]Package, error) {
	var packages []Package
	var currentPackage *Package
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			if currentPackage != nil {
				packages = append(packages, *currentPackage)
				currentPackage = nil
			}
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line format: %s", line)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "Package" {
			if currentPackage != nil {
				packages = append(packages, *currentPackage)
			}
			currentPackage = &Package{Name: value}
		} else if currentPackage != nil {
			switch key {
			case "Version":
				currentPackage.Version = value
			case "Depends":
				currentPackage.Depends = strings.Split(value, ", ")
			case "Provides":
				currentPackage.Provides = value
			case "Alternatives":
				currentPackage.Alternatives = strings.Split(value, ", ")
			case "License":
				currentPackage.License = value
			case "Section":
				currentPackage.Section = value
			case "CPE-ID":
				currentPackage.CPEID = value
			case "Architecture":
				currentPackage.Architecture = value
			case "Installed-Size":
				n, err := fmt.Sscanf(value, "%d", &currentPackage.InstalledSize)
				if err != nil || n != 1 {
					return nil, fmt.Errorf("failed to parse Installed-Size: %w", err)
				}
			case "Filename":
				currentPackage.Filename = value
			case "Size":
				fmt.Sscanf(value, "%d", &currentPackage.Size)
			case "SHA256sum":
				currentPackage.SHA256sum = value
			case "Description":
				currentPackage.Description = value
			}
		}
	}

	if currentPackage != nil {
		packages = append(packages, *currentPackage)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan input: %w", err)
	}

	return packages, nil
}

func LoadFromManifest(r io.Reader) ([]Package, error) {
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
func (a PackageList) Equals(b PackageList) []Diff {
	diff := make([]Diff, 0)
	for _, pkgA := range a {
		for _, pkgB := range b {
			// check if the package names are the same:
			if pkgA.Name == pkgB.Name {
				if pkgA.Version != pkgB.Version {
					diff = append(diff, Diff{
						Name:              pkgA.Name,
						UpstreamVersion:   pkgA.Version,
						DownstreamVersion: pkgB.Version,
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
