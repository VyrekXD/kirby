package commands

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/vyrekxd/kirby/commands/information"
	"github.com/vyrekxd/kirby/commands/starboard"
)

var CommandsData = []discord.ApplicationCommandCreate{
	information.Ping,
	starboard.Starboard,
}

// Main client
func HandleCommands(_ bot.Client) handler.Router {
	h := handler.New()

	h.Group(func(r handler.Router) {
		r.Command("/"+information.Ping.Name, information.PingHandler)
	})

	h.Group(func(r handler.Router) {
		r.Route("/"+starboard.Starboard.Name, func(r handler.Router) {
			r.Use(starboard.StarboardMiddleware)

			r.Command(
				"/"+starboard.Starboard.Options[0].OptionName(),
				starboard.StarboardInteractivo,
			)
			r.Command(
				"/"+starboard.Starboard.Options[1].OptionName(),
				starboard.StarboardManual,
			)
		})

		h.Component(
			"/"+starboard.SelectChannelId+"/{id}",
			starboard.SelectChannel,
		)

		h.Modal("/"+starboard.ModalId+starboard.GetId, starboard.Modal)

		h.Component(
			"/"+starboard.YesButtonId+starboard.GetId,
			starboard.YesButton,
		)
		h.Component(
			"/"+starboard.NoButtonId+starboard.GetId,
			starboard.NoButton,
		)
		h.Component(
			"/"+starboard.SkipButtonId+starboard.GetId,
			starboard.SkipButton,
		)
	})

	return h
}
