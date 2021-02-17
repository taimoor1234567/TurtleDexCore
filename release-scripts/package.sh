#!/usr/bin/env bash
set -e

version="$1"

# Directory where the binaries were placed. 
binDir="$2"

function package {
  os=$1
  arch=$2
 	
	echo Packaging ${os}...
 	folder=$binDir/TurtleDex-$version-$os-$arch
 	(
		cd $binDir
		zip -rq TurtleDex-$version-$os-$arch.zip TurtleDex-$version-$os-$arch
		sha256sum  TurtleDex-$version-$os-$arch.zip >> TurtleDex-$version-SHA256SUMS.txt
 	)
}

# Package amd64 binaries.
for os in darwin linux windows; do
  package "$os" "amd64"
done

# Package Raspberry Pi binaries.
package "linux" "arm64"
