package events

import (
	"github.com/disgoorg/disgo/bot"
	msgevents "github.com/vyrekxd/kirby/events/message_events"
)

func GetEvents(c bot.Client) []bot.EventListener {
	return []bot.EventListener{
		Ready(c),
		msgevents.MessageCreate(c),
		msgevents.MessageDelete(c),
		msgevents.MessageReactionAdd(c),
		msgevents.MessageReactionRemove(c),
		msgevents.MessageReactionRemoveAll(c),
		msgevents.MessageReactionRemoveEmoji(c),
	}
}
