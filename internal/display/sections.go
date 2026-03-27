package display

import (
	"io"
	"strings"

	"github.com/sumup/sumup-cli/internal/display/attribute"
)

// Section groups related output under a title.
type Section struct {
	Title string
	Pairs []attribute.KeyValue
	Lines []string
}

// RenderSections renders titled output blocks with a blank line between them.
func RenderSections(w io.Writer, sections []Section) {
	rendered := 0
	for _, section := range sections {
		if strings.TrimSpace(section.Title) == "" {
			continue
		}
		if len(section.Pairs) == 0 && len(section.Lines) == 0 {
			continue
		}

		if rendered > 0 {
			writeln(w)
		}

		writeln(w, section.Title)
		if len(section.Pairs) > 0 {
			DataList(w, section.Pairs)
		}
		for _, line := range section.Lines {
			writeln(w, line)
		}
		rendered++
	}
}
