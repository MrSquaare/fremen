package style

var NoEmoji = false

func EmojiText(symbol, text string) string {
	if NoEmoji {
		return text
	}
	return symbol + " " + text
}
