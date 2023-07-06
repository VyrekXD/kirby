package commands

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/vyrekxd/kirby/commands/information"
	"github.com/vyrekxd/kirby/commands/starboard"
	"github.com/vyrekxd/kirby/commands/starboard/components"
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
			r.Command(
				"/"+starboard.Starboard.Options[2].OptionName(),
				starboard.StarboardLista,
			)
			r.Command(
				"/"+starboard.Starboard.Options[3].OptionName(),
				starboard.StarboardVer,
			)
		})

		h.Component(
			"/"+starboard.SelectChannelId+"/{id}",
			components.SelectChannel,
		)

		h.Modal("/"+starboard.ModalId+starboard.GetId, components.Modal)

		h.Component(
			"/"+starboard.YesButtonId+starboard.GetId,
			components.YesButton,
		)
		h.Component(
			"/"+starboard.NoButtonId+starboard.GetId,
			components.NoButton,
		)
		h.Component(
			"/"+starboard.SkipButtonId+starboard.GetId,
			components.SkipButton,
		)

		h.Component(
			"/"+starboard.NextId+starboard.GetIndex,
			components.NextButton,
		)
		h.Component(
			"/"+starboard.ReturnId+starboard.GetIndex,
			components.ReturnButton,
		)
	})

	return h
}
