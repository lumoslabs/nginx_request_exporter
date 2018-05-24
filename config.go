package main

import (
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
)

var cfg *Config

type Config struct {
	ListenAddress  string             `json:"listen_address,omitempty"`
	TelemetryPath  string             `json:"telemetry_path,omitempty"`
	SyslogAddress  string             `json:"syslog_address,omitempty"`
	Buckets        []float64          `json:"buckets,omitempty"`
	DeviceType     *LabelConfig       `json:"device_type"`
	Prefix         *LabelConfig       `json:"prefix"`
	HistogramRules *HistogramRuleList `json:"histogram_rules"`
}

type LabelConfig struct {
	Default string    `json:"default,omitempty"`
	Rules   *RuleList `json:"rules,omitempty"`
}

type Rule struct {
	Value string `json:"value"`
	Regex string `json:"regex,omitempty"`
}

type RuleList []Rule

type HistogramRule struct {
	Labels map[string]string `json:"labels"`
}

type HistogramRuleList []HistogramRule

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

func (c *Config) mergeWithOverwrite(src *Config) error {
	return mergo.Merge(src, c, mergo.WithOverride)
}
