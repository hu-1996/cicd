#!/bin/bash

if [ -z "$1" ] || [ -z "$2" ] || [ -z "$3" ]; then
    echo "用法: ./build-image.sh <os> <arch> <module>"
    exit 1
fi

OS=$1
ARCH=$2
MODULE=$3

version=$(date "+%Y%m%d%H%M")
image=${MODULE}:${version}

echo ${image}
docker build \
    --build-arg TARGETOS=$OS \
    --build-arg TARGETARCH=$ARCH \
    --build-arg MODULE=$MODULE \
    -t ${image} .

# docker push ${image}



# ./build-image.sh linux amd64 cicd-server
