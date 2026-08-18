package manpages_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-cli/internal/commands/manpages"
)

func TestGenerateWritesGzippedManPage(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()
	var stdout bytes.Buffer

	root := &cli.Command{
		Name:   "sumup",
		Usage:  "Command line tool for the SumUp API",
		Writer: &stdout,
		Commands: []*cli.Command{
			manpages.NewCommand(),
			{
				Name:  "version",
				Usage: "Print CLI build information",
			},
		},
	}

	err := root.Run(t.Context(), []string{"sumup", "manpages", "--output", outputDir, "--gzip", "--text"})
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Wrote man pages to "+outputDir)

	textPath := filepath.Join(outputDir, "man1", "sumup.1")
	text, err := os.ReadFile(textPath)
	require.NoError(t, err)
	assert.Contains(t, string(text), "sumup")

	gzPath := filepath.Join(outputDir, "man1", "sumup.1.gz")
	gzFile, err := os.Open(gzPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = gzFile.Close() })

	reader, err := gzip.NewReader(gzFile)
	require.NoError(t, err)
	t.Cleanup(func() { _ = reader.Close() })

	decompressed, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, text, decompressed)
}

func TestCommandIsHidden(t *testing.T) {
	t.Parallel()

	command := manpages.NewCommand()
	assert.True(t, command.Hidden)
	assert.Equal(t, "manpages", command.Name)
}
