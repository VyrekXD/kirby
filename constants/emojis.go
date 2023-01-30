package constants

import "regexp"

type emojis struct {
	Star       string
	SecondStar string
	ThirdStar  string

	Error string
	Good  string
	Info  string
}

var Emojis = emojis{
	Star:       "⭐",
	SecondStar: "🌟",
	ThirdStar:  "🌟",
	Error:      "<:error:1066874219947888650>",
	Good:       "<:good:1066874222049230979>",
	Info:       "<:info:1066874226956566538>",
}

var DiscordEmojiRegex = regexp.MustCompile(`(?i)<(:[^\s:]+:)\w+>`)
var DiscordEmojiIdRegex = regexp.MustCompile(`:\w+>`)
var CleanIdRegex = regexp.MustCompile(`(:|>)`)
