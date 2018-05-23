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
		labels0 = &labelset{
			Names:  []string{"host", "status", "scheme"},
			Values: []string{"www.example.com", "404", "http"},
		}
		labels1 = &labelset{
			Names:  []string{"host", "status", "scheme"},
			Values: []string{"www.example.com", "200", "http"},
		}
		labels2 = &labelset{
			Names:  []string{"host", "status", "scheme"},
			Values: []string{"www.example.com", "200", "https"},
		}
		histRules = &HistogramRuleList{
			HistogramRule{
				Labels: map[string]string{
					"host":   "www.example.com",
					"status": "200",
				},
			},
			HistogramRule{
				Labels: map[string]string{
					"host":   "www.example.com",
					"scheme": "https",
				},
			},
			HistogramRule{
				Labels: map[string]string{
					"host":   "www.example.com",
					"status": "200",
					"foo":    "bar",
				},
			},
		}
		tests = []struct {
			labels *labelset
			expLen int
			expOk  bool
		}{
			{labels0, 0, false},
			{labels1, 1, true},
			{labels2, 2, true},
		}
	)

	for _, tt := range tests {
		matches, ok := matchHistogramRules(tt.labels, histRules)
		assert.Equal(t, tt.expLen, len(matches))
		assert.Equal(t, tt.expOk, ok)
	}
}
