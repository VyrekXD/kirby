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

		h.Component("/"+starboard.SelectChannelId, starboard.SelectChannel)

		h.Modal("/"+starboard.ModalId, starboard.Modal)

		h.Component("/"+starboard.YesButtonId, starboard.YesButton)
		h.Component("/"+starboard.NoButtonId, starboard.NoButton)
		h.Component("/"+starboard.OmitButtonId, starboard.OmitButton)
	})

	return h
}
