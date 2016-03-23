#!/bin/bash

set -e

DAEMON_NAME=skycli

mkdir -p dist

# build skygear server without C bindings
for GOOS in darwin linux windows; do
  for GOARCH in 386 amd64; do
    VERSION=`git describe --tags`
    FILENAME=$DAEMON_NAME-$VERSION-$GOOS-$GOARCH
    echo -n "Building $FILENAME... "
    GOOS=$GOOS GOARCH=$GOARCH \
        go build \
        -ldflags "-X github.com/skygeario/skycli/commands.version=$VERSION" \
        -o dist/$FILENAME \
        github.com/skygeario/skycli
    echo "Done"
  done
done
