package display

import (
	"io"
	"os"

	"golang.org/x/term"
)

func terminalWidth(w io.Writer) (int, bool) {
	file, ok := writerOrStdout(w).(*os.File)
	if !ok {
		return 0, false
	}

	fd := int(file.Fd())
	if !term.IsTerminal(fd) {
		return 0, false
	}
	width, _, err := term.GetSize(fd)
	if err != nil || width <= 0 {
		return 0, false
	}
	return width, true
}
