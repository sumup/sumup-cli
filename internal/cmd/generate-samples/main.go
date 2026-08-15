package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sumup/sumup-cli/internal/codesamples"
)

func main() {
	var outputPath string
	var cliVersion string
	flag.StringVar(&outputPath, "out", "", "output JSON path; defaults to stdout")
	flag.StringVar(&cliVersion, "cli-version", "", "SumUp CLI version represented by the samples")
	flag.Parse()

	if err := run(outputPath, cliVersion, os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "generate samples: %v\n", err)
		os.Exit(1)
	}
}

func run(outputPath, cliVersion string, stdout io.Writer) error {
	catalog, err := codesamples.Generate(cliVersion)
	if err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return fmt.Errorf("encode samples: %w", err)
	}
	encoded = append(encoded, '\n')
	return writeSamples(outputPath, encoded, stdout)
}

func writeSamples(outputPath string, encoded []byte, stdout io.Writer) error {
	if outputPath == "" {
		if _, err := stdout.Write(encoded); err != nil {
			return fmt.Errorf("write samples: %w", err)
		}
		return nil
	}

	directory := filepath.Dir(outputPath)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create output directory %q: %w", directory, err)
	}
	if err := os.WriteFile(outputPath, encoded, 0o644); err != nil {
		return fmt.Errorf("write samples %q: %w", outputPath, err)
	}
	return nil
}
