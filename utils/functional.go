package utils

import (
	"strings"
)

func Filter(values []string, predicate func(string) bool) []string {
	result := make([]string, 0)
	for _, value := range values {
		if predicate(value) {
			result = append(result, value)
		}
	}
	return result
}

func Exists(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// 切分, 去掉 empty, trim
func Split(value, sep string) []string {
	result := make([]string, 0)
	for _, term := range strings.Split(value, sep) {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}
		result = append(result, term)
	}
	return result
}
