package app

import (
	"strconv"
	"strings"
)

// compareVersions returns -1/0/1 for Go version strings like "go1.26.2",
// "go1.21.6", "go1.26rc1". Stable releases sort above their RC/beta variants
// of the same base, and lexicographically by numeric components.
func compareVersions(a, b string) int {
	pa, ra := parseVersion(a)
	pb, rb := parseVersion(b)
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var ai, bi int
		if i < len(pa) {
			ai = pa[i]
		}
		if i < len(pb) {
			bi = pb[i]
		}
		if ai != bi {
			if ai < bi {
				return -1
			}
			return 1
		}
	}
	// numeric portions equal — stable (empty suffix) > rc/beta (non-empty)
	switch {
	case ra == "" && rb == "":
		return 0
	case ra == "":
		return 1
	case rb == "":
		return -1
	}
	if ra == rb {
		return 0
	}
	if ra < rb {
		return -1
	}
	return 1
}

// parseVersion splits "go1.26.2" → ([]int{1,26,2}, "") and "go1.26rc1" →
// ([]int{1,26}, "rc1").
func parseVersion(v string) ([]int, string) {
	v = strings.TrimPrefix(v, "go")
	var parts []int
	var rest string
	cur := ""
	for i := 0; i < len(v); i++ {
		c := v[i]
		if c >= '0' && c <= '9' {
			cur += string(c)
		} else if c == '.' {
			if cur != "" {
				n, _ := strconv.Atoi(cur)
				parts = append(parts, n)
				cur = ""
			}
		} else {
			if cur != "" {
				n, _ := strconv.Atoi(cur)
				parts = append(parts, n)
				cur = ""
			}
			rest = v[i:]
			break
		}
	}
	if cur != "" {
		n, _ := strconv.Atoi(cur)
		parts = append(parts, n)
	}
	return parts, rest
}
