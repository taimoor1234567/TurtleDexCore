#!/usr/bin/env bash
set -e

version="$1"

# Directory where the binaries produces by build-release.sh are stored.
binDir="$2"

# Directory of the TurtleDex-UI Repo.
uiDir="$3"
if [ -z "$version" ] || [ -z "$binDir" ] || [ -z "$uiDir" ]; then
  echo "Usage: $0 VERSION BIN_DIRECTORY UI_DIRECTORY"
  exit 1
fi

echo Version: "${version}"
echo Binaries Directory: "${binDir}"
echo UI Directory: "${uiDir}"
echo ""

if [ "$SIA_SILENT_RELEASE" != "true" ]; then
	read -p "Continue (y/n)?" CONT
	if [ "$CONT" != "y" ]; then
		exit 1
	fi
fi
echo "Building TurtleDex-UI...";

# Get the absolute paths to avoid funny business with relative paths.
uiDir=$(realpath "${uiDir}")
binDir=$(realpath "${binDir}")

# Remove previously built UI binaries.
rm -r "${uiDir}"/release/

cd "${uiDir}"

# Copy over all the ttdxc/ttdxd binaries.
mkdir -p bin/{linux,mac,win}
cp "${binDir}"/TurtleDex-"${version}"-darwin-amd64/ttdxc bin/mac/
cp "${binDir}"/TurtleDex-"${version}"-darwin-amd64/ttdxd bin/mac/

cp "${binDir}"/TurtleDex-"${version}"-linux-amd64/ttdxc bin/linux/
cp "${binDir}"/TurtleDex-"${version}"-linux-amd64/ttdxd bin/linux/

cp "${binDir}"/TurtleDex-"${version}"-windows-amd64/ttdxc.exe bin/win/
cp "${binDir}"/TurtleDex-"${version}"-windows-amd64/ttdxd.exe bin/win/

# Build yarn deps.
yarn

# Build each of the UI binaries.
yarn package-linux
yarn package-win
yarn package

# Copy the UI binaries into the binDir. Also change the name at the same time.
# The UI builder doesn't handle 4 digit versions correctly and if the UI hasn't
# been updated the version might be stale.
for ext in AppImage dmg exe; do 
  mv "${uiDir}"/release/*."${ext}" "${binDir}"/TurtleDex-UI-"${version}"."${ext}"
	(
		cd "${binDir}"
		sha256sum TurtleDex-UI-"${version}"."${ext}" >> TurtleDex-"${version}"-SHA256SUMS.txt
	)
done
