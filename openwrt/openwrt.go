// Package openwrt provides functions to get package URLs for OpenWRT.
package openwrt

import "strings"

const (
	basePackages      = "https://downloads.openwrt.org/releases/$VERSION/packages/x86_64/base/Packages"
	addOnPackages     = "https://downloads.openwrt.org/releases/$VERSION/packages/x86_64/packages/Packages"
	luciPackages      = "https://downloads.openwrt.org/releases/$VERSION/packages/x86_64/luci/Packages"
	routingPackages   = "https://downloads.openwrt.org/releases/$VERSION/packages/x86_64/routing/Packages"
	telephonyPackages = "https://downloads.openwrt.org/releases/$VERSION/packages/x86_64/telephony/Packages"
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

func GetLuciPackageUrl(version string) string {
	url := luciPackages
	url = strings.Replace(url, "$VERSION", version, 1)
	return url
}

func GetRoutingPackageUrl(version string) string {
	url := routingPackages
	url = strings.Replace(url, "$VERSION", version, 1)
	return url
}

func GetTelephonyPackageUrl(version string) string {
	url := telephonyPackages
	url = strings.Replace(url, "$VERSION", version, 1)
	return url
}
