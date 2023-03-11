package events

import "github.com/disgoorg/disgo/bot"

func GetEvents(c bot.Client) []bot.EventListener {
	return []bot.EventListener{
		Ready(c),
	}
}
