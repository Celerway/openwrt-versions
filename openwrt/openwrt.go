// Package openwrt provides functions to get package URLs for OpenWRT.
package openwrt

import "strings"

const (
	basePackages  = "https://downloads.openwrt.org/releases/$VERSION/packages/$ARCH/base/Packages"
	addOnPackages = "https://downloads.openwrt.org/releases/$VERSION/targets/$ARCH/packages/Packages"
)

func GetBasePackageUrl(version, arch string) string {
	url := basePackages
	url = strings.Replace(url, "$VERSION", version, 1)
	url = strings.Replace(url, "$ARCH", arch, 1)
	return url
}

func GetAddOnPackageUrl(version, arch string) string {
	url := addOnPackages
	url = strings.Replace(url, "$VERSION", version, 1)
	url = strings.Replace(url, "$ARCH", arch, 1)
	return url
}
