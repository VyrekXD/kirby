package cmd_util

import (
	"context"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func defaultContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
}

func skipButton(
	ctx events.ComponentInteractionCreate,
) (<-chan *events.ComponentInteractionCreate, func()) {
	return bot.NewEventCollector(
		ctx.Client(),
		func(e *events.ComponentInteractionCreate) bool {
			return e.Type() == discord.ComponentTypeButton &&
				e.ButtonInteractionData().CustomID() == collSkipId &&
				ctx.User().ID == e.Message.Author.ID &&
				ctx.Channel().ID() == e.ChannelID()
		},
	)
}
