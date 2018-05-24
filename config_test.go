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
		cfg, er := Configure(tt.path, &Config{
			DeviceType: &LabelConfig{
				Default: "",
			},
			Prefix: &LabelConfig{
				Default: "",
			},
		})
		require.NoError(t, er)
		require.NotEqual(t, "", cfg.DeviceType.Default)
		require.NotEqual(t, 0, len(*cfg.DeviceType.Rules))
	}
}
