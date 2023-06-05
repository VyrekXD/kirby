package commands

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/vyrekxd/kirby/commands/information"
	"github.com/vyrekxd/kirby/commands/starboard"
)

var (
	CommandsData = []discord.ApplicationCommandCreate{
		information.Ping,
		starboard.Starboard,
	}
)

// Main client
func HandleCommands(_ bot.Client) handler.Router {
	h := handler.New()

	h.Command("/"+information.Ping.Name, information.PingHandler)
	h.Command("/"+starboard.Starboard.Name, starboard.StarboardHandler)

	return h
}
