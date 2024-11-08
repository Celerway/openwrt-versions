# OpenWRT Version diffing tool

## Motivation & executive summary

We should have a way to track differences in versions across versions of our firmware compared to upstream releases.

## Things to note:

OpenWRT produces a manifest file with the packaged versions of the packages in a release.
OpenWRT is a very minimal operating system that doesn’t have many packages; ~130 in total. Our distribution has about
550 packages.
Just comparing the manifest is not enough. We also need to figure out the OpenWRT versions of the packages that OpenWRT
doesn't ship, but are available as add-ons in OpenWRT.
So, on our side we just need the build manifest.
On the OpenWRT side we need the equivalent build manifest as well as the complete package
index (https://downloads.openwrt.org/releases/23.05.5/targets/x86/64/packages/Packages)

## How it works

Compare the versions of the packages in our base firmware image to the upstream packages. If the package doesn’t exist,
ignore. If the package has the same version, ignore it, if our package is outdated print out a warning.

## Usage

The resulting cli should work like this:

```shell
$ differ -release 23.05.5 celerway-2.12.0-86-64.manifest
```

The look will find the OpenWRT release manifest and package index itself, then go through the supplied manifest and
highlight any outdated packages.
