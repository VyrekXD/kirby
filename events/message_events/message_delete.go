package msgevents

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/vyrekxd/kirby/database/models"
	"go.mongodb.org/mongo-driver/bson"
)

func MessageDelete(c bot.Client) bot.EventListener {
	return bot.NewListenerFunc(func(e *events.MessageDelete) {
		go func() {
			if e.GuildID == nil {
				return
			} else if len(e.Message.Reactions) == 0 {
				return
			}

			guildData := models.GuildConfig{
				Lang: "es-MX",
			}
			err := models.GuildConfigColl().
				FindByID(e.GuildID.String(), &guildData)
			if err != nil {
				return
			}

			data := &models.StarboardMessage{}
			err = models.StarboardMessageColl().First(bson.M{
				"guild_id":   e.GuildID.String(),
				"channel_id": e.ChannelID.String(),
				"_id":        e.MessageID.String(),
			}, data)
			if err != nil {
				return
			}

			models.StarboardMessageColl().Delete(data)
		}()
	})
}
