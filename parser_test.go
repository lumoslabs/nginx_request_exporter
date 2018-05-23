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
		}
		tests = []struct {
			src      string
			expected string
		}{
			{"some iPhone device", "ios"},
			{"some stupid browser", "web"},
		}
	)

	for _, tt := range tests {
		assert.Equal(t, tt.expected, parseRule(tt.src, defaultValue, rules))

	}
}
