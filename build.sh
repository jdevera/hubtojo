#!/usr/bin/env sh

# Get the current version from git
version=$(git describe --tags --always)
echo "Building HubToJo version $version"

mkdir -p build

# Build the application with the current version
cd hubtojo && go build \
  -ldflags "-X main.Version=$version" \
  -o ../build/hubtojo
  "$@"
