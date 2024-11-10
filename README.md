**# OpenWRT Version diffing tool

## Motivation & executive summary

We should have a way to track differences in versions across versions of custom firmware release compared to upstream
releases.

## Things to note:

* OpenWRT produces a manifest file with the packaged versions of the packages in a release.
* OpenWRT is a very minimal operating system that doesn’t have many packages; ~130 in total. Downstream distributions
  might have more packages.
* Just comparing the manifest is not enough. We also need to figure out the OpenWRT versions of the packages that
  OpenWRT doesn't ship, but are available as add-ons in OpenWRT.
* We need the build manifest for the custom firmware.
* On the OpenWRT side we need the equivalent build manifest as well as the complete package
  index (https://downloads.openwrt.org/releases/23.05.5/targets/x86/64/packages/Packages)

## How it works

Compare the versions of the packages in our base firmware image to the upstream packages. If the package doesn’t exist,
ignore. If the package has the same version, ignore it, if our package is outdated print out a warning.

## Usage

The resulting cli should work like this:

```shell
$ openwrt-versions -release 23.05.5 celerway-2.12.0-86-64.manifest
Name           | Upstream               | Downstream            
---------------+------------------------+-----------------------
base-files     | 1562-r24106-10cc5fcd00 | 1586-r24106-10cc5fcd00
ipset          | 7.17-1                 | 7.6-1                 
libipset13     | 7.17-1                 | 7.6-1                 
ugps           | 2021-06-08-5e88403f-2  | 2022-02-19-fb87d0fd-2 
wireless-regdb | 2024.10.07-1           | 2024.07.04-1          
```

The look will find the OpenWRT release manifest and package index itself, then go through the supplied manifest and
highlight any outdated packages.


## Building the tool

```shell
CGO_ENABLED=0 go build -o bin/openwrt-versions .
```

## Development hygiene

```shell
golangci-lint run
```

## Tag a release:

Use `bump` to tag a release:

```shell
bump (-patch|-minor|-major)
```
See [bump](https://github.com/celerway/bump) for more information and source.