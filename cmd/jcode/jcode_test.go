package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

var writeGoldenFiles = flag.Bool("write-golden-files", false,
	"Whether or not to overwrite golden files with test output.")

var testFormat = formatParams{
	valid: func(pkg, variable, code string) bool {
		return code == "{code0}"
	},
	gen: func(pkg, variable string) string {
		return "{code1}"
	},
}

func TestInOut(t *testing.T) {
	cases := []struct {
		name string
		msgs []string
	}{
		{
			name: "0",
			msgs: []string{
				"testdata/0.in: ErrInvalidCode1: incorrect jettison code (fixed)",
				"testdata/0.in: ErrMissingCode: missing jettison code (fixed)",
				"testdata/0.in: ErrInvalidCode2: incorrect jettison code (fixed)",
			},
		}, {
			name: "1",
			msgs: []string{
				"testdata/1.in: errMissingCode: missing jettison code (fixed)",
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			in := fmt.Sprintf("testdata/%s.in", test.name)
			out := fmt.Sprintf("testdata/%s.out", test.name)
			res, err := checkFile(in, testFormat, true)
			assert.NoError(t, err)
			assert.False(t, res.pass)
			assert.EqualValues(t, test.msgs, res.msgs)
			verifyOutput(t, out, res.out)
		})
	}
}

func verifyOutput(t *testing.T, golden string, output []byte) {
	flag.Parse()
	if *writeGoldenFiles {
		ioutil.WriteFile(golden, output, 0644)
		// Nothing to check if we're writing.
		return
	}

	contents, err := ioutil.ReadFile(golden)
	if err != nil {
		t.Errorf("Error reading golden file %s: %v", golden, err)
	}

	assert.Equal(t, string(contents), string(output))
}

func TestBase64(t *testing.T) {
	uuid := fmtBase64("", "")
	assert.Len(t, uuid, 8)
	assert.True(t, validBase64("", "", uuid))
}

func TestErrHex4(t *testing.T) {
	code := fmtErrHex("", "")
	assert.Len(t, code, 4+16)
	assert.True(t, validErrHex("", "", code))
}
