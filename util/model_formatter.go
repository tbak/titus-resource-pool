package util

// Data structures and functions for pretty formatting resource pools, nodes, pods, machine types, etc.

import (
	"k8s.io/apimachinery/pkg/util/json"
)

const (
	FormatCompact    FormatDetailsLevel = 0
	FormatEssentials FormatDetailsLevel = 1
	FormatDetails    FormatDetailsLevel = 2
)

type FormatDetailsLevel int

type FormatterOptions struct {
	Level FormatDetailsLevel
}

func ToJSONString(value interface{}) string {
	bytes, err := json.Marshal(value)
	if err != nil {
		return "<formatting error>"
	}
	return string(bytes)
}
