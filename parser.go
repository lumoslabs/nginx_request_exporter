// Copyright 2016 Markus Lindenberg
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/scanner"

	"github.com/prometheus/common/log"
)

type metric struct {
	Name  string
	Value float64
}

type labelset struct {
	Names  []string
	Values []string
}

type histogram struct {
	Name   string
	Labels *labelset
}

func (l *labelset) Equals(labels []string) bool {
	if len(l.Names) != len(labels) {
		return false
	}
	for i := range l.Names {
		if l.Names[i] != labels[i] {
			return false
		}
	}
	return true
}

func (l *labelset) Get(key string) (string, bool) {
	for i, k := range l.Names {
		if k == key {
			return l.Values[i], true
		}
	}
	return "", false
}

func (l *labelset) Set(key, val string) {
	if _, exists := l.Get(key); exists {
		l.Delete(key)
	}
	l.Names = append(l.Names, key)
	l.Values = append(l.Values, val)
}

func (l *labelset) Delete(key string) {
	for i, k := range l.Names {
		if k == key {
			l.Names = append(l.Names[:i], l.Names[i+1:]...)
			l.Values = append(l.Values[:i], l.Values[i+1:]...)
		}
	}
}

func parseMessage(src string) (metrics []metric, labels *labelset, err error) {
	metrics = make([]metric, 0)
	labels = &labelset{
		Names:  make([]string, 0),
		Values: make([]string, 0),
	}

	var s scanner.Scanner
	s.Init(strings.NewReader(src))
	var tok rune
	for {
		tok = s.Scan()
		if tok == scanner.EOF {
			return
		} else if tok != scanner.Ident {
			err = fmt.Errorf("Ident expected at %v, got %s", s.Pos(), scanner.TokenString(tok))
			return
		}
		name := s.TokenText()

		tok = s.Scan()
		if tok == ':' {
			// Metric
			tok = s.Scan()
			if tok == scanner.Float || tok == scanner.Int {
				var value float64
				value, err = strconv.ParseFloat(s.TokenText(), 64)
				if err != nil {
					return
				}
				metrics = append(metrics, metric{
					Name:  name,
					Value: value,
				})
			} else {
				err = fmt.Errorf("Float or Int expected at %v, got %s", s.Pos(), scanner.TokenString(tok))
				return
			}

		} else if tok == '=' {
			// Label
			tok = s.Scan()
			var value string
			if tok == scanner.Ident || tok == scanner.Float || tok == scanner.Int {
				value = s.TokenText()
			} else if tok == scanner.String {
				value, err = strconv.Unquote(s.TokenText())
				if err != nil {
					return
				}
			} else {
				err = fmt.Errorf("Ident or String expected at %v, got %s", s.Pos(), scanner.TokenString(tok))
			}
			labels.Names = append(labels.Names, name)
			labels.Values = append(labels.Values, value)
		} else {
			err = fmt.Errorf(": or = expected at %v, got %s", s.Pos(), scanner.TokenString(tok))
			return
		}
	}
	return
}

func parseRule(src, defaultValue string, rules *RuleList) string {
	if rules == nil {
		return defaultValue
	}

	for _, r := range *rules {
		var regex string

		if r.Regex == "" {
			regex = fmt.Sprintf("^%s", r.Value)
		} else {
			regex = r.Regex
		}

		if ok, er := regexp.MatchString(regex, src); ok {
			return r.Value
		} else if er != nil {
			log.Error(er)
		}
	}
	return defaultValue
}

func parseHistograms(metricName string, srcLabels *labelset, rules *HistogramRuleList) (histos []*histogram, ok bool) {
	histos = make([]*histogram, 0)

	if rules == nil {
		return
	}

	for _, rule := range *rules {
		if metricName == rule.Metric {
			if h, hOk := newHistogram(srcLabels, &rule); hOk {
				histos = append(histos, h)
				ok = true
			}
		}
	}

	return
}

func newHistogram(srcLabels *labelset, rule *HistogramRule) (h *histogram, ok bool) {
	h = &histogram{
		Name: rule.Name,
		Labels: &labelset{
			Names:  make([]string, 0),
			Values: make([]string, 0),
		},
	}

	for _, name := range keys(rule.Labels) {
		if val, valOk := srcLabels.Get(name); valOk {
			regex := rule.Labels[name]
			if match, er := regexp.MatchString(regex, val); match {
				h.Labels.Names = append(h.Labels.Names, name)
				h.Labels.Values = append(h.Labels.Values, val)
			} else {
				if er != nil {
					log.Error(er)
				}
				return
			}
		} else {
			return
		}
	}

	ok = true
	return
}

func keys(src map[string]string) []string {
	dst := make([]string, 0, len(src))
	for k, _ := range src {
		dst = append(dst, k)
	}
	sort.Strings(dst)
	return dst
}
