#!/bin/bash
cd `dirname $0`
BASEDIR=`pwd`
set -ex

DATE=20260526 # `date +%Y%m%d`
VERSION=0.1.1-$DATE
BUILD_DIR=$BASEDIR/build
PACKAGES_DIR=$BUILD_DIR/packages

codesign_executable () {
   codesign --entitlements entitlements.plist --verbose=4 -f -s "Developer ID Application: AMPL Optimization Inc. (ZNNBG5892S)" "$1" --options=runtime --timestamp
}

build () {
   platform=$1
   bdir=$BUILD_DIR/gokestrel.$platform
   name=gokestrel.$platform.$DATE
   mkdir -p $bdir
   rm -rf $bdir/*
   # cp commands/* $bdir/
   if [[ "$platform" == mswin* ]]; then
      ext=".exe"
   else
      ext=""
   fi
   output=kestrel$ext

   FLAGS="-X 'main.Version=v$VERSION'"
   FLAGS="$FLAGS -X 'main.SolValidationEnv=$SOLVALIDATION_ENV'"
   FLAGS="$FLAGS -X 'main.SolValidationHeader=$SOLVALIDATION_HEADER'"
   if [[ "$platform" == macos64 ]]; then
      GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.VERSION=v$VERSION" -o $bdir/gokestrel_amd64 ./cmd/gokestrel
      GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.VERSION=v$VERSION" -o $bdir/gokestrel_arm64 ./cmd/gokestrel
      lipo -create -output $bdir/gokestrel $bdir/gokestrel_amd64 $bdir/gokestrel_arm64
      rm $bdir/gokestrel_amd64 $bdir/gokestrel_arm64
      codesign_executable $bdir/gokestrel
   else 
      GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w $FLAGS" -o $bdir/$output ./cmd/gokestrel
   fi

   mkdir -p $bdir/docs
   cp CHANGELOG.md $bdir/docs/CHANGES.kestrel.md
   cd $bdir
   if [[ "$platform" == mswin* ]]; then
      zip -r $name.zip *
      cp $name.zip $PACKAGES_DIR
   else
      tar czvf $name.tgz *
      cp $name.tgz $PACKAGES_DIR
   fi
   cd -
}

mkdir -p $PACKAGES_DIR
rm -rf $PACKAGES_DIR/*

GOOS=darwin build macos64
#GOOS=windows GOARCH=386 build mswin32
GOOS=windows GOARCH=amd64 build mswin64
#GOOS=linux GOARCH=386 build linux-intel32
GOOS=linux GOARCH=amd64 build linux-intel64
#GOOS=linux GOARCH=arm build linux-arm32
GOOS=linux GOARCH=arm64 build linux-arm64
#GOOS=linux GOARCH=ppc64le build linux-ppcle64
find build
echo
echo "SOLVALIDATION_ENV: $SOLVALIDATION_ENV"
echo "SOLVALIDATION_HEADER: $SOLVALIDATION_HEADER"