package main

import (
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
)

var cfg *Config

type Config struct {
	ListenAddress string                  `json:"listen_address,omitempty"`
	TelemetryPath string                  `json:"telemetry_path,omitempty"`
	SyslogAddress string                  `json:"syslog_address,omitempty"`
	Buckets       []float64               `json:"buckets,omitempty"`
	DeviceType    *LabelConfig            `json:"device_type"`
	Prefix        *LabelConfig            `json:"prefix"`
	Histograms    []*HistogramLabelConfig `json:"histograms"`
}

type HistogramLabelConfig struct {
	Labels map[string]string `json:"label"`
}

type LabelConfig struct {
	Default string    `json:"default,omitempty"`
	Rules   *RuleList `json:"rules,omitempty"`
}

type RuleList []Rule

type Rule struct {
	Value string `json:"value"`
	Regex string `json:"regex,omitempty"`
}

func Configure(path string, src *Config) (*Config, error) {
	var (
		c  = new(Config)
		er error
	)
	if path != "" {
		if c, er = newConfigFromFile(path); er != nil {
			return nil, fmt.Errorf("Failed to read config: %v", er)
		}
	}
	return c, c.mergeWithOverwrite(src)
}

func newConfigFromFile(path string) (*Config, error) {
	var c Config
	data, er := ioutil.ReadFile(path)
	if er != nil {
		return nil, er
	}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) merge(src *Config) error {
	return mergo.Merge(c, src)
}

func (c *Config) mergeWithOverwrite(src *Config) error {
	return mergo.MergeWithOverwrite(c, src)
}
