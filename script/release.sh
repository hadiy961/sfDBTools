#!/bin/bash

VERSION=${1:-"latest"}
PLATFORMS="linux/amd64 windows/amd64 darwin/amd64"

echo "Building sfDBTools version: $VERSION"

mkdir -p releases

for platform in $PLATFORMS; do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    
    output_name="sfDBTools-$VERSION-$GOOS-$GOARCH"
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi
    
    echo "Building $output_name..."
    env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-X main.version=$VERSION" -o releases/$output_name main.go
    
    if [ $? -ne 0 ]; then
        echo "Build failed for $platform"
        exit 1
    fi
done

echo "Build completed. Files in releases/ directory:"
ls -la releases/