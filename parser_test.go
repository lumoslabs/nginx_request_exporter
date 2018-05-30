package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRule(t *testing.T) {
	var (
		defaultValue = "web"
		rules        = &RuleList{
			Rule{
				Value: "ios",
				Regex: "iPhone",
			},
			Rule{
				Value: "/admin",
			},
		}
		tests = []struct {
			src      string
			expected string
		}{
			{"some iPhone device", "ios"},
			{"some stupid browser", "web"},
			{"/admin", "/admin"},
		}
	)

	for _, tt := range tests {
		assert.Equal(t, tt.expected, parseRule(tt.src, defaultValue, rules))
	}
}

func TestMatchHistogramRules(t *testing.T) {
	var (
		metric = "time"

		labels_no_matches = &labelset{
			Names:  []string{"host", "status", "scheme"},
			Values: []string{"nomatch.example.com", "404", "http"},
		}
		labels_one_match = &labelset{
			Names:  []string{"host", "status", "scheme"},
			Values: []string{"www.example.com", "200", "http"},
		}
		labels_two_matches = &labelset{
			Names:  []string{"host", "status", "scheme"},
			Values: []string{"www.example.com", "200", "https"},
		}
		labels_metric_rename = &labelset{
			Names:  []string{"host", "scheme", "prefix"},
			Values: []string{"www.example2.com", "https", "/"},
		}

		histRules = &HistogramRuleList{
			HistogramRule{
				Metric: metric,
				Name:   "time",
				Labels: map[string]string{
					"host":   "www.example.com",
					"status": "200",
				},
			},
			HistogramRule{
				Metric: metric,
				Name:   "time",
				Labels: map[string]string{
					"host":   "www.example.com",
					"scheme": "https",
				},
			},
			HistogramRule{
				Metric: metric,
				Name:   "time",
				Labels: map[string]string{
					"host":   "www.example.com",
					"status": "200",
					"foo":    "bar",
				},
			},
			HistogramRule{
				Metric: metric,
				Name:   "time_scheme_prefix",
				Labels: map[string]string{
					"host":   "www.example2.com",
					"scheme": "https",
					"prefix": ".*",
				},
			},
		}

		tests = []struct {
			labels *labelset
			length int
			ok     bool
			name   string
		}{
			{labels_no_matches, 0, false, "time"},
			{labels_one_match, 1, true, "time"},
			{labels_two_matches, 2, true, "time"},
			{labels_metric_rename, 1, true, "time_scheme_prefix"},
		}
	)

	for _, tt := range tests {
		histos, ok := parseHistograms(metric, tt.labels, histRules)
		assert.Equal(t, tt.length, len(histos))
		assert.Equal(t, tt.ok, ok)

		if ok && len(histos) == 1 {
			assert.Equal(t, tt.name, histos[0].Name)
		}
	}
}
