#!/bin/sh

set -e

: ${SKYGEAR_VERSION:=latest}
IMAGE_NAME=skygeario/skycli:v$SKYGEAR_VERSION

if [ -d dist ]; then
    echo "Error: Directory 'dist' exists."
    exit 1
fi

if [ -f skycli ]; then
    echo "Error: File 'skycli' exists."
    exit 1
fi

docker build -t skycli-build -f Dockerfile .
docker run -it --rm -v `pwd`:/go/src/app -w /go/src/app skycli-build /go/src/app/scripts/build-binary.sh
cp dist/skycli-linux-amd64 skycli
docker build --pull -t $IMAGE_NAME -f Dockerfile-release .

echo "Done. Run \`docker push $IMAGE_NAME\` to push image."
