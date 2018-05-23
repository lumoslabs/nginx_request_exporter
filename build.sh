#!/usr/bin/env bash

export DEP_VERSION=v0.4.1
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0
export GODEBUG=netdns=cgo

export VERSION=`perl -ne 'print "v$1\n" if /Version = "(.+)"/' doc.go`
export OUTFILE="nginx_request_exporter-$VERSION.$GOOS-$GOARCH"

dep ensure -v -vendor-only

go build -v -a \
  -ldflags "-s -X main.REFERENCE=$REF -X main.REVISION=$REV" \
  -installsuffix cgo \
  -o $OUTFILE

chmod -v +x $OUTFILE
tar -czf $OUTFILE.tar.gz $OUTFILE
rm -f $OUTFILE
