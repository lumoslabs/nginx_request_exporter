package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type configTestCase func(*testing.T, *Config)

func deviceTypeTests(t *testing.T, c *Config) {
	assert.Equal(t, "web", c.DeviceType.Default)
	assert.Equal(t, "iPhone", (*c.DeviceType.Rules)[0].Regex)
}

func prefixTests(t *testing.T, c *Config) {
	assert.Equal(t, "web", c.Prefix.Default)
	assert.Equal(t, "^/$", (*c.Prefix.Rules)[0].Regex)
	assert.Equal(t, "", (*c.Prefix.Rules)[1].Regex)
	assert.Equal(t, "/api", (*c.Prefix.Rules)[2].Value)
}

func prefixNoDefaultTests(t *testing.T, c *Config) {
	assert.Equal(t, "", c.Prefix.Default)
	assert.Equal(t, "^/$", (*c.Prefix.Rules)[0].Regex)
	assert.Equal(t, "", (*c.Prefix.Rules)[1].Regex)
	assert.Equal(t, "/api", (*c.Prefix.Rules)[2].Value)
}

func histogramTests(t *testing.T, c *Config) {
	assert.Equal(t, "https", (*c.HistogramRules)[0].Labels["scheme"])
	assert.Equal(t, "200", (*c.HistogramRules)[0].Labels["status"])
	assert.Equal(t, ".*", (*c.HistogramRules)[1].Labels["prefix"])
}

func defaultConfigTests(t *testing.T, c *Config) {
	assert.Equal(t, defaultListenAddr, c.ListenAddress)
	assert.Equal(t, defaultTelemetryPath, c.TelemetryPath)
	assert.Equal(t, defaultSyslogAddr, c.SyslogAddress)
}

func TestConfig(t *testing.T) {
	var (
		tests = []struct {
			path  string
			pass  bool
			cases configTestCase
		}{
			{"fixtures/config/bad.yaml", false, nil},
			{"fixtures/config/noexist.yml", false, nil},
			{"fixtures/config/device_types.yml", true, deviceTypeTests},
			{"fixtures/config/prefix.yml", true, prefixTests},
			{"fixtures/config/prefix-no-default.yml", true, prefixNoDefaultTests},
			{"fixtures/config/histograms.yml", true, histogramTests},
			{"fixtures/config/histograms.yml", true, defaultConfigTests},
		}
	)

	defaultConfig := &Config{
		ListenAddress: defaultListenAddr,
		TelemetryPath: defaultTelemetryPath,
		SyslogAddress: defaultSyslogAddr,
		Buckets:       []float64{0.1, 0.5},
		Prefix: &LabelConfig{
			Default: "",
			Rules:   nil,
		},
		DeviceType: &LabelConfig{
			Default: "",
			Rules:   nil,
		},
	}

	for _, tt := range tests {
		if c, er := Configure(tt.path, defaultConfig); tt.pass {
			assert.NoError(t, er)
			tt.cases(t, c)
		} else {
			assert.Error(t, er)
		}
	}
}
