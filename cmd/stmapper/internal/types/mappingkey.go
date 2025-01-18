package types

import (
	"slices"
	"strings"
)

type MappingKeyFunc func(name, tag string) string

// NameMappingKeyFunc 根据字段名称进行映射
func NameMappingKeyFunc(name, tag string) string {
	return name
}

// TagMappingKeyFunc 根据tag进行映射
func TagMappingKeyFunc(name, tag, targetTag string) string {
	tag = strings.Trim(tag, "`")
	tags := strings.Split(tag, " ")
	idx := slices.IndexFunc(tags, func(s string) bool {
		return strings.Contains(s, targetTag)
	})
	if idx == -1 {
		return ""
	}
	// stmapper:"id"
	_, a, found := strings.Cut(tags[idx], ":")
	if !found {
		return ""
	}
	a = strings.Trim(a, "\"")
	b, _, found := strings.Cut(a, ",")
	if !found {
		return a
	}

	return b
}
