package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/celerway/openwrt-versions/openwrt"
	"github.com/celerway/openwrt-versions/pkgidx"
)

//go:embed .version
var embeddedVersion string

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Println("error: ", err)
	}
	os.Exit(0)
}

func run(ctx context.Context, args []string) error {
	fmt.Printf("OpenWRT package version comparison tool version %s\n", embeddedVersion)

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	version := fs.String("version", "", "Version for upstream")
	verbose := fs.Bool("verbose", false, "Verbose output")
	err := fs.Parse(args)
	if err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}
	// check that the version is not empty
	if *version == "" {
		return fmt.Errorf("version cannot be empty")
	}
	baseUrl := openwrt.GetBasePackageUrl(*version)
	addonUrl := openwrt.GetAddOnPackageUrl(*version)

	UpstreamPackages, err := pkgidx.LoadFromURL(ctx, baseUrl)
	if err != nil {
		return fmt.Errorf("base packages: %w", err)
	}
	addonPackages, err := pkgidx.LoadFromURL(ctx, addonUrl)
	if err != nil {
		return fmt.Errorf("addon packages: %w", err)
	}

	// UpstreamPackages = append(UpstreamPackages, addonPackages...)
	mergedPackages := make(map[string]string)
	for name, pkg := range UpstreamPackages {
		// fmt.Println(name)
		mergedPackages[name] = pkg
	}
	for name, pkg := range addonPackages {
		// fmt.Println(name)
		mergedPackages[name] = pkg
	}
	UpstreamPackages = mergedPackages

	// Load the downstream packages, see if there is an argument on the command line
	// to read from a file instead of stdin
	var input io.Reader
	if len(fs.Args()) > 0 {
		input, err = os.Open(fs.Arg(0))
		if err != nil {
			return fmt.Errorf("open file: %w", err)
		}
	} else {
		input = os.Stdin
	}
	downstreamPackages, err := pkgidx.LoadFromManifest(input)
	if err != nil {
		return fmt.Errorf("downstream packages: %w", err)
	}

	// Compare the two package lists
	differences, cwyOnly := downstreamPackages.Equals(UpstreamPackages)
	if *verbose {
		fmt.Println("OpenWRT version used (x86_64): ", *version)
		fmt.Println("base url: ", baseUrl)
		fmt.Println("addon url: ", addonUrl)
		fmt.Println("Upstream packages count: ", len(UpstreamPackages))
		fmt.Println("Downstream package count: ", len(downstreamPackages))
		fmt.Println("Differences count: ", len(differences))
	}

	nameWidth := 0
	upstreamWidth := 0
	downstreamWidth := 0
	for _, pkg := range differences {
		if len(pkg.Name) > nameWidth {
			nameWidth = len(pkg.Name)
		}
		if len(pkg.UpstreamVersion) > upstreamWidth {
			upstreamWidth = len(pkg.UpstreamVersion)
		}
		if len(pkg.DownstreamVersion) > downstreamWidth {
			downstreamWidth = len(pkg.DownstreamVersion)
		}
	}

	// Print the header
	fmt.Printf("%-*s | %-*s | %-*s\n", nameWidth, "Name", upstreamWidth, "Upstream", downstreamWidth, "Downstream")
	fmt.Printf("%s-+-%s-+-%s\n", strings.Repeat("-", nameWidth), strings.Repeat("-", upstreamWidth), strings.Repeat("-", downstreamWidth))

	// Print the package information
	for _, pkg := range differences {
		fmt.Printf("%-*s | %-*s | %-*s\n", nameWidth, pkg.Name, upstreamWidth, pkg.UpstreamVersion, downstreamWidth, pkg.DownstreamVersion)
	}

	fmt.Println()

	fmt.Println("Packages only in CelerwayOS:")
	fmt.Println("Press enter to see the packages")
	_, err = fmt.Scanln()
	if err != nil {
		return fmt.Errorf("scanln: %w", err)
	}

	nameWidth = 0
	upstreamWidth = 0
	downstreamWidth = 0
	for _, pkg := range cwyOnly {
		if len(pkg.Name) > nameWidth {
			nameWidth = len(pkg.Name)
		}
		if len(pkg.UpstreamVersion) > upstreamWidth {
			upstreamWidth = len(pkg.UpstreamVersion)
		}
		if len(pkg.DownstreamVersion) > downstreamWidth {
			downstreamWidth = len(pkg.DownstreamVersion)
		}
	}

	// Print the package information
	for _, pkg := range cwyOnly {
		fmt.Printf("%-*s | %-*s | %-*s\n", nameWidth, pkg.Name, upstreamWidth, pkg.UpstreamVersion, downstreamWidth, pkg.DownstreamVersion)
	}

	return nil
}
