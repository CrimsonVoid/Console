package styles

import (
	"fmt"
)

type Attrib string

const (
	Reset      Attrib = "\x1b[0m"
	Bright     Attrib = "\x1b[1m"
	Dim        Attrib = "\x1b[2m"
	Underscore Attrib = "\x1b[4m"
	Blink      Attrib = "\x1b[5m"
	Reverse    Attrib = "\x1b[7m"
	Hidden     Attrib = "\x1b[8m"
)

// Paints the text with `Attrib` and formats the text with Sprintf
func (self Attrib) Paint(format string, a ...interface{}) string {
	return string(self) + fmt.Sprintf(format, a...) + string(Reset)
}
