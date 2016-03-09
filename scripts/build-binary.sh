#!/bin/sh

set -e

DAEMON_NAME=skycli
DIST=dist

mkdir -p $DIST

# build skygear server without C bindings
for GOOS in darwin linux windows; do
  for GOARCH in 386 amd64; do
    FILENAME=$DAEMON_NAME-$GOOS-$GOARCH
    GOOS=$GOOS GOARCH=$GOARCH go build -o $DIST/$FILENAME github.com/oursky/skycli
  done
done
