#!/usr/bin/env bash

export DEP_VERSION=v0.4.1
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0
export GODEBUG=netdns=cgo

export VERSION=`perl -ne 'print "v$1\n" if /Version = "(.+)"/' doc.go`
export OUTFILE="nginx_request_exporter-$VERSION.$GOOS-$GOARCH"
export REFERENCE=${REFERENCE:-$(git rev-parse --abbrev-ref HEAD)}
export REVISION=${REVISION:-$(git rev-parse --short=8 HEAD)}

dep ensure -v -vendor-only

go build -v -a \
  -ldflags "-s -X main.REFERENCE=$REFERENCE -X main.REVISION=$REVISION" \
  -installsuffix cgo \
  -o $OUTFILE

chmod -v +x $OUTFILE
tar -czf $OUTFILE.tar.gz $OUTFILE
rm -f $OUTFILE
