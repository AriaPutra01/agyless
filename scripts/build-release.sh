#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-$(git describe --tags --abbrev=0 2>/dev/null || echo 1.0.0)}"
VERSION="${VERSION#v}"
OUT="dist/release"

rm -rf "$OUT"
mkdir -p "$OUT"

TARGETS="
agyless-linux-x64 linux amd64
agyless-linux-arm64 linux arm64
agyless-darwin-x64 darwin amd64
agyless-darwin-arm64 darwin arm64
agyless-windows-x64.exe windows amd64
"

echo "$TARGETS" | while read -r asset goos goarch; do
  [ -z "$asset" ] && continue
  printf 'compiling %s (%s/%s) … ' "$asset" "$goos" "$goarch"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
    go build -trimpath -ldflags "-s -w" \
    -o "${OUT}/${asset}" ./cmd/agyless
  echo "ok"
done

echo
echo "Built binaries for v${VERSION} in ${OUT}/:"
ls -lh "$OUT"
