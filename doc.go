package main

import "strings"

const Version = "0.0.2"

var (
	REFERENCE string
	REVISION  string
)

func version() string {
	return strings.Join([]string{Version, REFERENCE, REVISION}, "/")
}
