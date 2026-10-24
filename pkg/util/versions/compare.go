package versions

import (
	"sort"
	"strconv"
	"strings"
)

// compare compares two version strings
// returns -1 if v1 < v2, 1 if v1 > v2, 0 otherwise.
func compare(v1, v2 string) int {
	var (
		currTab  = strings.Split(v1, ".")
		otherTab = strings.Split(v2, ".")
	)

	max := len(currTab)
	if len(otherTab) > max {
		max = len(otherTab)
	}
	for i := 0; i < max; i++ {
		var currInt, otherInt int

		if len(currTab) > i {
			currInt, _ = strconv.Atoi(currTab[i])
		}
		if len(otherTab) > i {
			otherInt, _ = strconv.Atoi(otherTab[i])
		}
		if currInt > otherInt {
			return 1
		}
		if otherInt > currInt {
			return -1
		}
	}
	return 0
}

// LessThan checks if a version is less than another
func LessThan(v, other string) bool {
	return compare(v, other) == -1
}

// SortEngines orders engine names such as docker-29.10.0 by version, so
// that 29.10 comes after 29.9 (a plain string sort puts it before).
func SortEngines(engines []string) {
	sort.Slice(engines, func(i, j int) bool {
		return LessThan(strings.TrimPrefix(engines[i], "docker-"), strings.TrimPrefix(engines[j], "docker-"))
	})
}
