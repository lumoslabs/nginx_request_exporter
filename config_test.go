package main

import (
	"io/ioutil"
	"testing"

	"github.com/smallfish/simpleyaml"
	"github.com/stretchr/testify/require"
)

func parseYAML(p string) *simpleyaml.Yaml {
	f, _ := ioutil.ReadFile(p)
	y, _ := simpleyaml.NewYaml(f)
	return y
}

func TestConfig(t *testing.T) {
	var (
		tests = []struct {
			path   string
			checks map[string]string
		}{
			{"fixtures/config/device_types.yml", map[string]string{"device_types."}},
		}
	)

	for _, tt := range tests {
		y := parseYAML(tt.path)
		_, er := Configure(tt.path, &Config{})
		if require.NoError(t, er) {

		}
	}
}
