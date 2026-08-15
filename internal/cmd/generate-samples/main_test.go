package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sumup/sumup-cli/internal/codesamples"
)

func TestRunWritesCatalogToStdout(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	err := run("", "v1.2.3", &stdout)

	require.NoError(t, err)
	var catalog codesamples.Catalog
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &catalog))
	assert.Equal(t, "v1.2.3", catalog.SDK.Version)
	assert.NotEmpty(t, catalog.Samples)
	assert.Equal(t, byte('\n'), stdout.Bytes()[stdout.Len()-1])
}

func TestWriteSamplesCreatesOutputDirectory(t *testing.T) {
	t.Parallel()

	filename := filepath.Join(t.TempDir(), "nested", "samples.json")
	err := writeSamples(filename, []byte("samples\n"), &bytes.Buffer{})

	require.NoError(t, err)
	contents, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Equal(t, "samples\n", string(contents))
}

func TestRunRequiresCLIVersion(t *testing.T) {
	t.Parallel()

	err := run("", "", &bytes.Buffer{})

	require.EqualError(t, err, "cli version is required")
}
