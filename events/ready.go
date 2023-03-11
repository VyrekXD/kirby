package events

import (
	"context"
	"math/rand"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
	"github.com/vyrekxd/kirby/constants"
)

var scheduler *gocron.Scheduler

func init() {
	scheduler = gocron.NewScheduler(time.UTC)
}

func Ready(c bot.Client) bot.EventListener {
	return bot.NewListenerFunc(func(e *events.Ready) {
		log.Info().Msgf("Logged as: %v 👤", e.User.Username)

		presence := constants.Presences[rand.Intn(len(constants.Presences))]

		if err := c.SetPresence(
			context.TODO(),
			gateway.WithWatchingActivity(presence),
			gateway.WithOnlineStatus(discord.OnlineStatusIdle),
		); err != nil {
			log.Panic().
				Err(err).
				Msgf("An error ocurred trying to set presence: ")
		}

		scheduler.Every("3m").Do(func() {
			presence := constants.Presences[rand.Intn(len(constants.Presences))]

			if err := c.SetPresence(
				context.TODO(),
				gateway.WithWatchingActivity(presence),
				gateway.WithOnlineStatus(discord.OnlineStatusIdle),
			); err != nil {
				log.Panic().
					Err(err).
					Msgf("An error ocurred trying to set presence: ")
			}
		})
	})
}
