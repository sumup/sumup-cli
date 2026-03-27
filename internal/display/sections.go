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
func RenderSections(w io.Writer, sections []Section) error {
	rendered := 0
	for _, section := range sections {
		if strings.TrimSpace(section.Title) == "" {
			continue
		}
		if len(section.Pairs) == 0 && len(section.Lines) == 0 {
			continue
		}

		if rendered > 0 {
			if err := writeln(w); err != nil {
				return err
			}
		}

		if err := writeln(w, section.Title); err != nil {
			return err
		}
		if len(section.Pairs) > 0 {
			if err := DataList(w, section.Pairs); err != nil {
				return err
			}
		}
		for _, line := range section.Lines {
			if err := writeln(w, line); err != nil {
				return err
			}
		}
		rendered++
	}
	return nil
}
