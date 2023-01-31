package starboard

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func starboardInteractivo(ctx *handler.CommandEvent) error {
	ctx.CreateMessage(discord.NewMessageCreateBuilder().
		AddEmbeds(discord.Embed{}).
		AddContainerComponents(
			discord.NewActionRow(),
			discord.NewActionRow(),
		).
		Build(),
	)

}
