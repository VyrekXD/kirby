package msgevents

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/vyrekxd/kirby/database/models"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/exp/slices"
)

func MessageReactionRemoveAll(c bot.Client) bot.EventListener {
	return bot.NewListenerFunc(func(e *events.MessageReactionRemoveAll) {
		guildData := models.GuildConfig{
			Lang: "es-MX",
		}
		err := models.GuildConfigColl().
			FindByID(e.GuildID.String(), &guildData)
		if err != nil {
			return
		}

		allStarboards := []models.Starboard{}
		err = models.StarboardColl().SimpleFind(&allStarboards, bson.M{"guild_id": e.GuildID.String()})
		if err != nil {
			return
		}
		starboard := &models.Starboard{}

		for _, starb := range allStarboards {
			if !starb.ChannelListType {
				if starb.ChannelList != nil && !slices.Contains(starb.ChannelList, e.ChannelID.String()) {
					starboard = &starb
					break
				}
			} else {
				if starb.ChannelList != nil && slices.Contains(starb.ChannelList, e.ChannelID.String()) {
					starboard = &starb
					break
				}
			}
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
	})
}
