package mtesting

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"coolstercodes/modules/modulir"
)

// NewContext is a convenience helper to create a new modulir.Context suitable
// for use in the test suite.
func NewContext() *modulir.Context {
	return modulir.NewContext(&modulir.Args{Log: &modulir.Logger{Level: modulir.LevelInfo}})
}

// WriteTempFile writes the given data to a temporary file. It returns the path
// to the temporary file which should be removed with `defer os.Remove(path)`.
func WriteTempFile(t *testing.T, data []byte) string {
	t.Helper()

	tempFile, err := os.CreateTemp(t.TempDir(), "modulir")
	require.NoError(t, err)

	_, err = tempFile.Write(data)
	require.NoError(t, err)

	err = tempFile.Close()
	require.NoError(t, err)

	return tempFile.Name()
}
