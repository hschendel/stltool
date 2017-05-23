#!/bin/bash
VERSION=stltool-1.0.1
IDENTIFIER=de.hschendel.stltool

rm -f ${VERSION}.pkg
go build
rm -rf bundle
mkdir bundle
cp stltool bundle
pkgbuild --root ./bundle --identifier $IDENTIFIER --install-location /usr/local/bin ${VERSION}.pkg
rm -rf bundle

