package display

import (
	"os"

	"golang.org/x/term"
)

func terminalWidth() (int, bool) {
	fd := int(os.Stdout.Fd())
	if !term.IsTerminal(fd) {
		return 0, false
	}
	width, _, err := term.GetSize(fd)
	if err != nil || width <= 0 {
		return 0, false
	}
	return width, true
}
