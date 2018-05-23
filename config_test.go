package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	var (
		tests = []struct {
			path string
		}{
			{"fixtures/config/device_types.yml"},
			{"fixtures/config/full.yml"},
		}
	)

	for _, tt := range tests {
		_, er := Configure(tt.path, &Config{})
		require.NoError(t, er)
	}
}
