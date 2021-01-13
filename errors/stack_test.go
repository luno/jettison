package errors

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"path"
	"testing"

	"github.com/luno/jettison/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var writeGoldenFiles = flag.Bool("write-golden-files", false,
	"Whether or not to overwrite golden files with test output.")

//go:generate go test . -write-golden-files

// TestStack tests the stack trace including line numbers.
// Adding anything to this file might break the test.
func TestStack(t *testing.T) {
	err := stack(5)
	je, ok := err.(*JettisonError)
	require.True(t, ok)

	bb, err := json.MarshalIndent(je.Hops[0].StackTrace, "", "  ")
	require.NoError(t, err)

	verifyOutput(t, "log_"+t.Name(), internal.StripTestStacks(t, bb))
}

func stack(i int) error {
	if i == 0 {
		return New("stack")
	}
	return stack(i - 1)
}

func verifyOutput(t *testing.T, goldenFileName string, output []byte) {
	t.Helper()
	flag.Parse()
	goldenFilePath := path.Join("testdata", goldenFileName+".golden")

	if *writeGoldenFiles {
		err := ioutil.WriteFile(goldenFilePath, output, 0777)
		require.NoError(t, err)

		// Nothing to check if we're writing.
		return
	}

	contents, err := ioutil.ReadFile(goldenFilePath)
	require.NoError(t, err, "Error reading golden file %s: %v", goldenFilePath, err)

	assert.Equal(t, string(contents), string(output))
}
