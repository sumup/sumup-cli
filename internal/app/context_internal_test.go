package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeLocale(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "returns empty string for empty input", input: "", want: ""},
		{name: "trims whitespace and replaces underscore", input: " en_US.UTF-8 ", want: "en-US"},
		{name: "strips locale modifier", input: "sr_RS@latin", want: "sr-RS"},
		{name: "returns plain locale unchanged", input: "de-DE", want: "de-DE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, normalizeLocale(tt.input))
		})
	}
}
