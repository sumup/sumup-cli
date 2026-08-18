package manpages

import (
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"

	docs "github.com/urfave/cli-docs/v3"
	"github.com/urfave/cli/v3"
)

// NewCommand returns a hidden command that writes man pages for release packaging.
func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "manpages",
		Usage:     "Generate man pages for the SumUp CLI",
		UsageText: "sumup manpages [-o man] [--gzip] [--text]",
		Hidden:    true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Write man pages to the given folder",
				Value:   "man",
			},
			&cli.BoolFlag{
				Name:  "gzip",
				Usage: "Write a gzipped man page (sumup.1.gz)",
				Value: true,
			},
			&cli.BoolFlag{
				Name:  "text",
				Usage: "Also write an uncompressed man page (sumup.1)",
			},
		},
		Action: generate,
	}
}

func generate(_ context.Context, cmd *cli.Command) error {
	manpage, err := docs.ToManWithSection(cmd.Root(), 1)
	if err != nil {
		return fmt.Errorf("generate man page: %w", err)
	}

	dir := cmd.String("output")
	man1Dir := filepath.Join(dir, "man1")
	if err := os.MkdirAll(man1Dir, 0o755); err != nil {
		return fmt.Errorf("create man page directory: %w", err)
	}

	if cmd.Bool("text") {
		path := filepath.Join(man1Dir, "sumup.1")
		if err := os.WriteFile(path, []byte(manpage), 0o644); err != nil {
			return fmt.Errorf("write man page: %w", err)
		}
	}

	if cmd.Bool("gzip") {
		path := filepath.Join(man1Dir, "sumup.1.gz")
		file, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("create gzipped man page: %w", err)
		}
		gzWriter := gzip.NewWriter(file)
		if _, err := gzWriter.Write([]byte(manpage)); err != nil {
			_ = gzWriter.Close()
			_ = file.Close()
			return fmt.Errorf("write gzipped man page: %w", err)
		}
		if err := gzWriter.Close(); err != nil {
			_ = file.Close()
			return fmt.Errorf("close gzip writer: %w", err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("close gzipped man page: %w", err)
		}
	}

	_, err = fmt.Fprintf(cmd.Writer, "Wrote man pages to %s\n", dir)
	return err
}
