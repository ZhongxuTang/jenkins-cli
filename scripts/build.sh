#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
cd "$ROOT_DIR"

BIN_DIR=${BIN_DIR:-"$ROOT_DIR/bin"}
OUTPUT=${OUTPUT:-"$BIN_DIR/jcli"}
GOMODCACHE_PATH=${GOMODCACHE_PATH:-"$ROOT_DIR/.gocache"}
OUTPUT_DIR="$(dirname "$OUTPUT")"

mkdir -p "$OUTPUT_DIR" "$GOMODCACHE_PATH"

VERSION=${VERSION:-$(git describe --tags --always 2>/dev/null || echo "dev")}
COMMIT=${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "none")}
BUILD_DATE=${BUILD_DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}

LD_FLAGS=(
    "-X github.com/lemonsoul/jenkins-cli/pkg/version.Version=${VERSION}"
    "-X github.com/lemonsoul/jenkins-cli/pkg/version.Commit=${COMMIT}"
    "-X github.com/lemonsoul/jenkins-cli/pkg/version.BuildDate=${BUILD_DATE}"
)

echo "Building jcli"
echo "  version:    ${VERSION}"
echo "  commit:     ${COMMIT}"
echo "  build date: ${BUILD_DATE}"

env \
    GOMODCACHE="${GOMODCACHE_PATH}" \
    GOFLAGS="${GOFLAGS:-}" \
    go build -o "$OUTPUT" -ldflags "-s -w ${LD_FLAGS[*]}" ./

echo "Binary written to $OUTPUT"
