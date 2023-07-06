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
	"github.com/vyrekxd/kirby/database/models"
	"github.com/vyrekxd/kirby/langs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/slices"
)

func MessageReactionAdd(c bot.Client) bot.EventListener {
	return bot.NewListenerFunc(func(e *events.MessageReactionAdd) {
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

		starboard := &models.Starboard{}
		err = models.StarboardColl().First(bson.M{
			"guild_id": e.GuildID.String(),
			"emoji":    emoji,
		}, starboard)
		if err != nil {
			return
		}

		if starboard.ChannelListType == models.BlackList &&
			starboard.ChannelList != nil &&
			len(starboard.ChannelList) > 0 &&
			slices.Contains(starboard.ChannelList, e.ChannelID.String()) {
			return
		} else if starboard.ChannelListType == models.WhiteList &&
			starboard.ChannelList != nil &&
			len(starboard.ChannelList) > 0 &&
			!slices.Contains(starboard.ChannelList, e.ChannelID.String()) {
			return
		}

		starboardMsg := &models.StarboardMessage{}
		err = models.StarboardMessageColl().First(bson.M{"_id": msg.ID.String()}, starboardMsg)
		if err != nil && err != mongo.ErrNoDocuments {
			return
		} else if err == mongo.ErrNoDocuments {
			starboardMsg = &models.StarboardMessage{
				DefaultModel: models.DefaultModel{ID: msg.ID.String()},
				GuildId:      e.GuildID.String(),
				ChannelId:    e.ChannelID.String(),
				StarboardId:  starboard.ID,
				Users:        []string{},
			}

			err = models.StarboardMessageColl().Create(starboardMsg)
			if err != nil {
				return
			}
		}

		if slices.Contains(starboardMsg.Users, e.UserID.String()) {
			return
		}

		starboardMsg.Users = append(starboardMsg.Users, e.UserID.String())

		if starboard.Required > len(starboardMsg.Users) {
			models.StarboardMessageColl().Update(starboardMsg)
			return
		} else if starboardMsg.StarboardMsgId == "" && starboard.Required <= len(starboardMsg.Users) {
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

			msgCreate := discord.MessageCreate{
				Content: fmt.Sprintf(`%v | **%v**`, e.Emoji.Reaction(), len(starboardMsg.Users)),
				Embeds: []discord.Embed{
					{
						Author: &discord.EmbedAuthor{
							Name:    msg.Author.Username,
							IconURL: *msg.Author.AvatarURL(),
						},
						Color:       starboard.EmbedColor,
						Title:       starboard.Name,
						Description: strings.Join(contents, "\n"),
						Timestamp:   json.Ptr(time.Now()),
					},
				},
				Components: []discord.ContainerComponent{
					discord.ActionRowComponent{
						discord.ButtonComponent{
							Style: discord.ButtonStyleLink,
							Label: *langs.Pack(guildData.Lang).Command("reaction_events").Get("jump_to_message"),
							URL:   msg.JumpURL(),
						},
					},
				},
			}

			if len(images) > 0 {
				msgCreate.Embeds[0].Image = &discord.EmbedResource{
					URL: images[0].URL,
				}
			}

			newMsg, err := c.Rest().CreateMessage(
				snowflake.MustParse(starboard.ChannelId),
				msgCreate,
			)
			if err != nil {
				return
			}

			starboardMsg.StarboardMsgId = newMsg.ID.String()
			models.StarboardMessageColl().Update(starboardMsg)
		} else if starboard.Required <= len(starboardMsg.Users) {
			c.Rest().UpdateMessage(
				snowflake.MustParse(starboard.ChannelId),
				snowflake.MustParse(starboardMsg.StarboardMsgId),
				discord.MessageUpdate{
					Content: json.Ptr(fmt.Sprintf(`%v | **%v**`, e.Emoji.Reaction(), len(starboardMsg.Users))),
				},
			)
			models.StarboardMessageColl().Update(starboardMsg)
			return
		}
	})
}
