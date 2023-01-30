package starboard

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"
)

func starboardInteractivo(ctx *handler.CommandEvent) error {
	ctx.UpdateInteractionResponse(discord.MessageUpdate{
		Content: json.Ptr("Alrato joven."),
	})

	return nil
}
