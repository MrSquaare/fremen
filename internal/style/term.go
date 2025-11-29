package style

import "runtime"

const (
	ColorRed    = "\033[91m"
	ColorGreen  = "\033[92m"
	ColorYellow = "\033[93m"
	ColorBlue   = "\033[94m"
	ColorCyan   = "\033[96m"
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
)

type OutputPreferences struct {
	Color bool
	Emoji bool
}

var (
	ConsoleSupportsColor    = detectConsoleColorSupport()
	GlobalOutputPreferences = OutputPreferences{Color: true, Emoji: true}
)

func detectConsoleColorSupport() bool {
	if runtime.GOOS == "windows" {
		return false
	}
	return true
}

func getColorSequence(code string) string {
	if !ConsoleSupportsColor || !GlobalOutputPreferences.Color {
		return ""
	}
	return code
}

func StyledText(text, colorCode string) string {
	prefix := getColorSequence(colorCode)
	if prefix == "" {
		return text
	}
	suffix := getColorSequence(ColorReset)
	return prefix + text + suffix
}

func GetEmoji(symbol string) string {
	if !GlobalOutputPreferences.Emoji {
		return ""
	}
	return symbol
}

func EmojiText(symbol, text string) string {
	e := GetEmoji(symbol)
	if e != "" {
		return e + " " + text
	}
	return text
}
