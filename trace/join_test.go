package trace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMerge(t *testing.T) {
	type trace struct {
		trace  []string
		binary string
	}
	testCases := []struct {
		name         string
		traces       []trace
		expFullTrace []string
	}{
		{name: "empty"},
		{
			name: "single trace",
			traces: []trace{
				{trace: []string{"a", "b", "c"}, binary: "bin"},
			},
			expFullTrace: []string{"a", "b", "c"},
		},
		{
			name: "multiple traces from same binary",
			traces: []trace{
				{trace: []string{"one", "two", "three"}, binary: "bin"},
				{trace: []string{"a", "b", "c"}, binary: "bin"},
			},
			expFullTrace: []string{
				"a",
				"b",
				"c",
				"bin -> bin",
				"one",
				"two",
				"three",
			},
		},
		{
			name: "trace from one to the next",
			traces: []trace{
				{trace: []string{"from_a"}, binary: "a"},
				{trace: []string{"from_b"}, binary: "b"},
				{trace: []string{"from_c"}, binary: "c"},
			},
			expFullTrace: []string{
				"from_c",
				"b -> c",
				"from_b",
				"a -> b",
				"from_a",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var m Merge
			for _, tr := range tc.traces {
				m.Add(tr.trace, tr.binary)
			}
			assert.Equal(t, tc.expFullTrace, m.FullTrace())
		})
	}
}
