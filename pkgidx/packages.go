package pkgidx

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strings"
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
