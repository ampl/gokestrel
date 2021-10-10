#!/bin/bash
cd `dirname $0`
BASEDIR=`pwd`
set -ex

BUILD_DIR=$BASEDIR/build
PACKAGES_DIR=$BUILD_DIR/packages

build () {
   platform=$1
   bdir=$BUILD_DIR/gokestrel.$platform
   name=gokestrel.$platform #.`date +%Y%m%d`
   mkdir -p $bdir
   # rm -rf $bdir/*
   if [[ "$platform" == mswin* ]]; then
      ext=".exe"
   else
      ext=""
   fi
   output=gokestrel$ext
   go build -o $bdir/$output ./cmd/gokestrel
   cd $bdir
   if [[ "$platform" == mswin* ]]; then
      zip -r $name.zip $output
      cp $name.zip $PACKAGES_DIR
   else
      tar czvf $name.tgz $output
      cp $name.tgz $PACKAGES_DIR
   fi
   cd -
}

mkdir -p $PACKAGES_DIR
rm -rf $PACKAGES_DIR/*
GOOS=windows GOARCH=386 build mswin32
GOOS=windows GOARCH=amd64 build mswin64
GOOS=linux GOARCH=386 build linux-intel32
GOOS=linux GOARCH=amd64 build linux-intel64
GOOS=darwin GOARCH=amd64 build macos64
GOOS=linux GOARCH=ppc64le build linux-ppcle64
find build
