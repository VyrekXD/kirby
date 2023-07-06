package msgevents

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/vyrekxd/kirby/database/models"
	"go.mongodb.org/mongo-driver/bson"
)

func MessageReactionRemoveEmoji(c bot.Client) bot.EventListener {
	return bot.NewListenerFunc(func(e *events.MessageReactionRemoveEmoji) {
		go func() {
			if e.GuildID == nil {
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

			var emoji string
			if e.Emoji.ID != nil {
				emoji = e.Emoji.ID.String()
			} else {
				emoji = e.Emoji.Reaction()
			}

			starboard := &models.Starboard{}
			err = models.StarboardColl().First(bson.M{"guild_id": e.GuildID.String(), "emoji": emoji}, starboard)
			if err != nil {
				return
			}

			starboardMsg := &models.StarboardMessage{}
			err = models.StarboardMessageColl().First(bson.M{"_id": e.MessageID.String(), "starboard_id": starboard.ID.Hex()}, starboardMsg)
			if err != nil {
				return
			}

			models.StarboardColl().Delete(starboardMsg)
			c.Rest().DeleteMessage(
				snowflake.MustParse(starboard.ChannelId),
				snowflake.MustParse(starboardMsg.StarboardMsgId),
			)
		}()
	})
}
