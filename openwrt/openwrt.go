// Package openwrt provides functions to get package URLs for OpenWRT.
package openwrt

import "strings"

const (
	basePackages  = "https://downloads.openwrt.org/releases/$VERSION/packages/x86_64/base/Packages"
	addOnPackages = "https://downloads.openwrt.org/releases/$VERSION/packages/x86_64/packages/Packages"
)

func GetBasePackageUrl(version string) string {
	url := basePackages
	url = strings.Replace(url, "$VERSION", version, 1)
	return url
}

func GetAddOnPackageUrl(version string) string {
	url := addOnPackages
	url = strings.Replace(url, "$VERSION", version, 1)
	return url
}
