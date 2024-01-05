#!/usr/bin/env bash

OS="linux darwin windows"
ARCHITECTURES="amd64 arm64"

VERSION=$1
NAME=$2
MODULE_NAME=$3

mkdir -p "build"
for arch in $ARCHITECTURES; do
  for os in $OS; do
    echo "Building $os-$arch"
    if [ "$os" == "windows" ]; then
      GOOS=$os GOARCH=$arch go build -ldflags "-s -w -X $MODULE_NAME/pkg/version.Version=$VERSION" -o build/${NAME}.exe ./cmd/syncer
      cd build
      tar zcvf $NAME-$os-$arch-${VERSION}.tar.gz ${NAME}.exe
      rm ${NAME}.exe
      cd ..
    else
      GOOS=$os GOARCH=$arch go build -ldflags "-s -w -X $MODULE_NAME/pkg/version.Version=$VERSION" -o build/${NAME} ./cmd/syncer
      cd build
      chmod +x ${NAME}
      tar zcvf $NAME-$os-$arch-${VERSION}.tar.gz ${NAME}
      rm -f ${NAME}
      cd ..
    fi
  done
done