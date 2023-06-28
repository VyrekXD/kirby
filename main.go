package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	snowflake "github.com/disgoorg/snowflake/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/vyrekxd/kirby/commands"
	"github.com/vyrekxd/kirby/config"
	"github.com/vyrekxd/kirby/database"
	"github.com/vyrekxd/kirby/events"
	"github.com/vyrekxd/kirby/langs"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.Kitchen,
		FormatLevel: func(i interface{}) string {
			if i == nil {
				i = "log"
			}

			return strings.ToUpper(
				fmt.Sprintf("| %v |", strings.ToUpper(fmt.Sprint(i))),
			)
		},
	})
}

func main() {
	err := database.Connect()
	if err != nil {
		return
	}

	log.Info().Msg("Connected to MongoDB 📁")

	err = langs.Load()
	if err != nil {
		log.Panic().Err(err).Send()
	}

	client, err := disgo.New(config.Token,
		bot.WithGatewayConfigOpts(
			config.Intents,
		),
		bot.WithCacheConfigOpts(
			cache.WithCaches(
				cache.FlagGuilds|cache.FlagMembers|cache.FlagMessages|cache.FlagChannels|cache.FlagEmojis,
			),
		),
	)
	if err != nil {
		log.Panic().Err(err).Msg("Error when trying to create client: ")
	}

	client.AddEventListeners(events.GetEvents(client)...)
	client.AddEventListeners(commands.HandleCommands(client))

	if os.Getenv("APP_ENV") == "development" {
		for _, id := range config.DevServersId {
			_, err := client.Rest().
				SetGuildCommands(client.ApplicationID(), snowflake.MustParse(id), commands.CommandsData)
			if err != nil {
				log.Panic().
					Err(err).
					Msg("Error when creating commands (development mode): ")
			}
		}
	} else {
		_, err := client.Rest().SetGlobalCommands(client.ApplicationID(), commands.CommandsData)
		if err != nil {
			log.Panic().Err(err).Msg("Error when creating commands: ")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err = client.OpenGateway(ctx); err != nil {
		log.Panic().Err(err).Msg("Error when trying to connect to gateway: ")
	}

	defer client.Close(context.TODO())

	log.Info().Msg("Kirby is now running 🚀.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
