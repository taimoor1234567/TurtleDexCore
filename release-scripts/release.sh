#!/usr/bin/env bash
set -e

# version and keys are supplied as arguments
version="$1"
rc=`echo $version | awk -F - '{print $2}'`
if [[ -z $version ]]; then
	echo "Usage: $0 VERSION"
	exit 1
fi

# setup build-time vars
ldflags="-s -w -X 'github.com/turtledex/TurtleDexCore/build.GitRevision=`git rev-parse --short HEAD`' -X 'github.com/turtledex/TurtleDexCore/build.BuildTime=`git show -s --format=%ci HEAD`' -X 'github.com/turtledex/TurtleDexCore/build.ReleaseTag=${rc}'"

function build {
  os=$1
  arch=$2

	echo Building ${os}...
	# create workspace
	folder=release/TurtleDex-$version-$os-$arch
	rm -rf $folder
	mkdir -p $folder
	# compile and hash binaries
	for pkg in ttdxc ttdxd; do
		bin=$pkg
		if [ "$os" == "windows" ]; then
			bin=${pkg}.exe
		fi
		GOOS=${os} GOARCH=${arch} go build -a -tags 'netgo' -trimpath -ldflags="$ldflags" -o $folder/$bin ./cmd/$pkg
		(
			cd release/
			sha256sum TurtleDex-$version-$os-$arch/$bin >> TurtleDex-$version-SHA256SUMS.txt
		)
  done

	cp -r doc LICENSE README.md $folder
}

# Build amd64 binaries.
for os in darwin linux windows; do
  build "$os" "amd64"
done

# Build Raspberry Pi binaries.
build "linux" "arm64"
