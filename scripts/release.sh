#!/usr/bin/env bash
# Builds the py2go toolchain for every supported target into dist/, with
# -trimpath and no build id so that the same source gives the same bytes, and
# prints the SHA-256 of each file. The binary embeds the standard library
# (stdlib/*.py) and needs no cgo.
#
#   scripts/release.sh                 all targets
#   scripts/release.sh linux/amd64     one target
set -euo pipefail
cd "$(dirname "$0")/.."
targets=("$@")
if [ ${#targets[@]} -eq 0 ]; then
	targets=(linux/amd64 linux/arm64 darwin/amd64 darwin/arm64)
fi
mkdir -p dist
for t in "${targets[@]}"; do
	os=${t%/*}
	arch=${t#*/}
	out="dist/py2go-$os-$arch"
	CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -trimpath -ldflags="-s -w -buildid=" -o "$out" ./cmd/py2go
	sha256sum "$out"
done
