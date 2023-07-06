package msgevents

import (
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/json"
	"github.com/disgoorg/snowflake/v2"
	"github.com/rs/zerolog/log"
	"github.com/vyrekxd/kirby/database/models"
	"github.com/vyrekxd/kirby/langs"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/exp/slices"
)

func MessageReactionRemove(c bot.Client) bot.EventListener {
	return bot.NewListenerFunc(func(e *events.MessageReactionRemove) {
		go func() {
			if e.GuildID == nil {
				return
			}
			msg, err := c.Rest().GetMessage(e.ChannelID, e.MessageID)
			if err != nil {
				return
			}

			guildData := models.GuildConfig{
				Lang: "es-MX",
			}
			err = models.GuildConfigColl().
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

			log.Print(1)
			starboard := &models.Starboard{}
			err = models.StarboardColl().First(bson.M{"guild_id": e.GuildID.String(), "emoji": emoji}, starboard)
			if err != nil {
				return
			}

			log.Print(2)
			starboardMsg := &models.StarboardMessage{}
			err = models.StarboardMessageColl().First(bson.M{"_id": msg.ID.String()}, starboardMsg)
			if err != nil {
				return
			}

			if !slices.Contains(starboardMsg.Users, e.UserID.String()) {
				return
			}
			log.Print(3)
			i := slices.Index(starboardMsg.Users, e.UserID.String())
			starboardMsg.Users = append(starboardMsg.Users[:i], starboardMsg.Users[i+1:]...)

			if len(starboardMsg.Users) < starboard.Required {
				c.Rest().DeleteMessage(
					snowflake.MustParse(starboard.ChannelId),
					snowflake.MustParse(starboardMsg.StarboardMsgId),
				)

				models.StarboardMessageColl().Delete(starboardMsg)

				return
			}

			log.Print(4)

			text := msg.Content
			if len(text) > 2048 {
				text = text[:2048]
			}

			contents := []string{
				text,
				"\n",
			}

			images := []discord.Attachment{}
			if len(msg.Attachments) > 0 {
				attach := []discord.Attachment{}

				for _, attachment := range msg.Attachments {
					switch *attachment.ContentType {
					case "image/png", "image/jpeg", "image/gif":
						{
							images = append(images, attachment)
							continue
						}
					default:
						{
							attach = append(attach, attachment)
							continue
						}
					}
				}

				if len(attach) > 0 {
					name := attach[0].Filename
					if len(name) > 256 {
						name = name[:256]
					}

					contents = append(
						contents,
						*langs.Pack(guildData.Lang).Command("reaction_events").Get("attachments"),
						fmt.Sprintf("**[%v](%v)**", name, attach[0].URL),
					)
					if len(attach) > 1 {
						contents = append(
							contents,
							*langs.Pack(guildData.Lang).Command("reaction_events").Getf("more_attachments", len(attach)-1),
						)
					}
				}
			}

			err = models.StarboardMessageColl().Update(starboardMsg)
			if err != nil {
				return
			}
			log.Print(5)

			updateMsg := discord.MessageUpdate{
				Content: json.Ptr(fmt.Sprintf(`%v | **%v**`, e.Emoji.Reaction(), len(starboardMsg.Users))),
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: &discord.EmbedAuthor{
							Name:    msg.Author.Username,
							IconURL: *msg.Author.AvatarURL(),
						},
						Title:       starboard.Name,
						Description: strings.Join(contents, "\n"),
						Timestamp:   json.Ptr(time.Now()),
					},
				}),
			}

			if len(images) > 0 {
				e := *updateMsg.Embeds
				e[0].Image = &discord.EmbedResource{
					URL: images[0].URL,
				}

				updateMsg.Embeds = &e
			}

			c.Rest().UpdateMessage(
				snowflake.MustParse(starboard.ChannelId),
				snowflake.MustParse(starboardMsg.StarboardMsgId),
				updateMsg,
			)
			log.Print(6)
		}()
	})
}
