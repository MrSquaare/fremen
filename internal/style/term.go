package style

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/mattn/go-colorable"
)

type Color string

const (
	ColorRed    Color = "\033[91m"
	ColorGreen  Color = "\033[92m"
	ColorYellow Color = "\033[93m"
	ColorBlue   Color = "\033[94m"
	ColorCyan   Color = "\033[96m"
	ColorReset  Color = "\033[0m"
	ColorBold   Color = "\033[1m"
)

type TermStyler struct {
	Color  bool
	Emoji  bool
	Writer io.Writer
}

func NewTermStyler(color, emoji bool) *TermStyler {
	var writer io.Writer
	if runtime.GOOS == "windows" {
		writer = colorable.NewColorable(os.Stdout)
	} else {
		writer = os.Stdout
	}
	return &TermStyler{
		Color:  color,
		Emoji:  emoji,
		Writer: writer,
	}
}

func (t *TermStyler) PrintLnColor(text string, colorCode Color) {
	out := text
	if t.Color {
		out = string(colorCode) + text + string(ColorReset)
	}
	fmt.Fprintln(t.Writer, out)
}

func (t *TermStyler) EmojiText(symbol, text string) string {
	out := text
	if t.Emoji {
		out = symbol + " " + text
	}

	return out
}
