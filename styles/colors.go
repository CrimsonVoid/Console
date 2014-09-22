package styles

import (
	"fmt"
)

type Color int

const (
	Clear   Color = 0 // Do not change existing colors
	Black   Color = 30
	Red     Color = 31
	Green   Color = 32
	Yellow  Color = 33
	Blue    Color = 34
	Magenta Color = 35
	Cyan    Color = 36
	White   Color = 37
)

// Sets the foreground to `Color` and formats the text with Sprintf
func (self Color) Fg(format string, a ...interface{}) string {
	return PaintColors(self, Clear, format, a...)
}

// Sets the background to `Color` and formats the text with Sprintf
func (self Color) Bg(format string, a ...interface{}) string {
	return PaintColors(Clear, self, format, a...)
}

// Sets the foregroudn and background to `Color` and formats the text with
// Sprintf
func (self Color) Paint(format string, a ...interface{}) string {
	return PaintColors(self, self, format, a...)
}

// Sets the foreground to `fg` and the background to `bg` and formats the text
// with Sprintf
func PaintColors(fg, bg Color, format string, a ...interface{}) string {
	return setColors(fg, bg) + fmt.Sprintf(format, a...) + string(Reset)
}

// Returns a string with the just the appropriate color codes
func setColors(fg, bg Color) string {
	color := ""

	if fg != Clear {
		color += fmt.Sprintf("\x1b[%vm", fg)
	}
	if bg != Clear {
		color += fmt.Sprintf("\x1b[%vm", bg+10)
	}

	return color
}
