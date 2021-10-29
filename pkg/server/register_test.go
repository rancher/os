package server

import (
	"testing"

	"gotest.tools/assert"
)

func TestBuildName(t *testing.T) {
	data := map[string]interface{}{
		"level1A": map[string]interface{}{
			"level2A": "level2AValue",
			"level2B": map[string]interface{}{
				"level3A": "level3AValue",
			},
		},
		"level1B": "level1BValue",
	}

	testCase := []struct {
		Format string
		Output string
	}{
		{
			Format: "${level1B}",
			Output: "level1bvalue",
		},
		{
			Format: "${level1B",
			Output: "m-level1b",
		},
		{
			Format: "a${level1B",
			Output: "a-level1b",
		},
		{
			Format: "${}",
			Output: "m",
		},
		{
			Format: "${",
			Output: "m-",
		},
		{
			Format: "a${",
			Output: "a-",
		},
		{
			Format: "${level1A}",
			Output: "m",
		},
		{
			Format: "a${level1A}c",
			Output: "ac",
		},
		{
			Format: "a${level1A}",
			Output: "a",
		},
		{
			Format: "${level1A}c",
			Output: "c",
		},
		{
			Format: "a${level1A/level2A}c",
			Output: "alevel2avaluec",
		},
		{
			Format: "a${level1A/level2B/level3A}c",
			Output: "alevel3avaluec",
		},
		{
			Format: "a${level1A/level2B/level3A}c${level1B}",
			Output: "alevel3avalueclevel1bvalue",
		},
	}

	for _, testCase := range testCase {
		assert.Equal(t, testCase.Output, buildName(data, testCase.Format))
	}
}
