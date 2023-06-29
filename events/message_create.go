package events

import (
	"context"
	"fmt"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/json"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/disgoorg/snowflake/v2"
	"github.com/forPelevin/gomoji"
	"github.com/vyrekxd/kirby/commands/starboard"
	"github.com/vyrekxd/kirby/constants"
	"github.com/vyrekxd/kirby/database/models"
	"github.com/vyrekxd/kirby/langs"
	"github.com/vyrekxd/kirby/utils"
)

func MessageCreate(c bot.Client) bot.EventListener {
	return bot.NewListenerFunc(func(e *events.MessageCreate) {
		if e.Message.Author.Bot {
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

		cmdPack := langs.Pack(guildData.Lang).
			Command("starboard").
			SubCommand("interactivo")
		data := &models.TempStarboard{}
		cur := models.TempStarboardColl().FindOne(context.TODO(), bson.M{"guild_id": e.GuildID.String()})
		err = cur.Decode(data)
		if err != nil {
			return
		}
		if e.Message.Author.ID.String() != data.UserId {
			return
		}

		content := e.Message.Content

		switch data.Phase {
		case models.PhaseBotsMessages:
			{
				if res := constants.DiscordEmojiRegex.FindString(fmt.Sprint(content)); res != "" {
					content = constants.CleanIdRegex.ReplaceAllString(
						constants.DiscordEmojiIdRegex.FindString(
							fmt.Sprint(content),
						),
						"",
					)
				} else if res := gomoji.FindAll(content); len(res) > 1 || len(res) == 0 {
					starboard.DeleteTempStarboard(data)
					_, err = c.Rest().UpdateMessage(
						snowflake.MustParse(data.MsgChannelId),
						snowflake.MustParse(data.MessageId),
						discord.MessageUpdate{
							Content:    cmdPack.Get("errNoValidEmoji"),
							Embeds:     json.Ptr([]discord.Embed{}),
							Components: json.Ptr([]discord.ContainerComponent{}),
						},
					)
					if err != nil {
						log.Error().Err(err).Msg("Error ocurred when trying to update main message in \"starboard:interactivo:message_create\"")
					}

					return
				}

				data.Emoji = content
				data.Phase = models.PhaseEmoji
				err = models.TempStarboardColl().Update(data)
				if err != nil {
					starboard.DeleteTempStarboard(data)
					c.Rest().UpdateMessage(
						snowflake.MustParse(data.MsgChannelId),
						snowflake.MustParse(data.MessageId),
						discord.MessageUpdate{
							Content:    cmdPack.Getf("errCantUpdate", err),
							Embeds:     json.Ptr([]discord.Embed{}),
							Components: json.Ptr([]discord.ContainerComponent{}),
						})

					return
				}
				_, err = c.Rest().UpdateMessage(
					snowflake.MustParse(data.MsgChannelId),
					snowflake.MustParse(data.MessageId),
					discord.MessageUpdate{
						Embeds: json.Ptr([]discord.Embed{
							{
								Author: json.Ptr(discord.EmbedAuthor{
									Name:    e.Message.Author.Username,
									IconURL: *e.Message.Author.AvatarURL(),
								}),
								Title:       *cmdPack.Get("starboardCreating"),
								Color:       constants.Colors.Main,
								Description: *cmdPack.Get("selectEmbedColor"),
								Timestamp:   json.Ptr(time.Now()),
								Fields: []discord.EmbedField{
									{
										Name: "\u0020",
										Value: *cmdPack.Get("starboardData") +
											*cmdPack.Getf("starboardDataChannel", data.ChannelId) +
											*cmdPack.Getf("starboardDataName", data.Name) +
											*cmdPack.Getf("starboardDataEmoji", content),
										Inline: json.Ptr(true),
									},
									{
										Name: "\u0020",
										Value: *cmdPack.Get("starboardRequisites") +
											*cmdPack.Getf("starboardRequisitesRequired", data.Required) +
											*cmdPack.Getf("starboardRequisitesBotsReact", utils.ReadableBool(
												&data.BotsReact,
												*langs.Pack(guildData.Lang).GetGlobal("yes"),
												*langs.Pack(guildData.Lang).GetGlobal("no"),
											)) +
											*cmdPack.Getf("starboardRequisitesBotsMsg", utils.ReadableBool(
												&data.BotsMessages,
												*langs.Pack(guildData.Lang).GetGlobal("yes"),
												*langs.Pack(guildData.Lang).GetGlobal("no"),
											)),
										Inline: json.Ptr(true),
									},
								},
							},
						}),
						Components: json.Ptr([]discord.ContainerComponent{
							discord.NewActionRow(
								discord.NewSecondaryButton(
									*cmdPack.Get("skip"),
									starboard.SkipButtonId+"/"+data.ID.Hex(),
								),
							),
						}),
					})
				if err != nil {
					starboard.DeleteTempStarboard(data)
					log.Error().Err(err).Msg("Error ocurred when trying to respond in \"starboard:interactivo:message_create\"")

					return
				}

				utils.WaitDo(time.Second*50, func() {
					find := &models.TempStarboard{}
					err := models.TempStarboardColl().First(data, find)

					if err == nil && find.EmbedColor == 0 {
						err := models.TempStarboardColl().Delete(find)
						if err != nil {
							log.Error().Err(err).Msg("Error ocurred when trying to delete document in \"starboard:interactivo\"")
						}

						_, errM := c.Rest().UpdateMessage(
							snowflake.MustParse(data.MsgChannelId),
							snowflake.MustParse(data.MessageId),
							discord.MessageUpdate{
								Content:    cmdPack.Get("errTimeout"),
								Embeds:     json.Ptr([]discord.Embed{}),
								Components: json.Ptr([]discord.ContainerComponent{}),
							},
						)
						if errM != nil {
							log.Error().Err(err).Msg("Error ocurred when trying to edit message in \"starboard:interactivo:message_create\"")
						}
					}
				})

				return
			}
		case models.PhaseEmoji:
			{
				h, err := utils.ToHex(content)
				if err != nil {
					starboard.DeleteTempStarboard(data)
					c.Rest().UpdateMessage(
						snowflake.MustParse(data.MsgChannelId),
						snowflake.MustParse(data.MessageId),
						discord.MessageUpdate{
							Content: langs.Pack(guildData.Lang).
								Command("starboard").
								SubCommand("manual").
								Get("noValidHex"),
							Embeds:     json.Ptr([]discord.Embed{}),
							Components: json.Ptr([]discord.ContainerComponent{}),
						},
					)

					return
				}

				data.EmbedColor = h
				data.Phase = models.PhaseEmbedColor
				err = models.TempStarboardColl().Update(data)
				if err != nil {
					starboard.DeleteTempStarboard(data)
					c.Rest().UpdateMessage(
						snowflake.MustParse(data.MsgChannelId),
						e.Message.ID,
						discord.MessageUpdate{
							Content:    cmdPack.Getf("errCantUpdate", err),
							Embeds:     json.Ptr([]discord.Embed{}),
							Components: json.Ptr([]discord.ContainerComponent{}),
						})

					return
				}

				var emoji string
				emojiId, err := snowflake.Parse(data.Emoji)
				if err != nil {
					emoji = data.Emoji
				} else {
					emojiData, err := c.Rest().GetEmoji(*e.GuildID, emojiId)
					if err != nil {
						starboard.DeleteTempStarboard(data)
						c.Rest().UpdateMessage(
							snowflake.MustParse(data.MsgChannelId),
							e.Message.ID,
							discord.MessageUpdate{
								Content:    cmdPack.Getf("errUnexpected", err),
								Embeds:     json.Ptr([]discord.Embed{}),
								Components: json.Ptr([]discord.ContainerComponent{}),
							})

						return
					}

					emoji = emojiData.Mention()
				}

				_, err = c.Rest().UpdateMessage(
					snowflake.MustParse(data.MsgChannelId),
					snowflake.MustParse(data.MessageId),
					discord.MessageUpdate{
						Embeds: json.Ptr([]discord.Embed{
							{
								Author: json.Ptr(discord.EmbedAuthor{
									Name:    e.Message.Author.Username,
									IconURL: *e.Message.Author.AvatarURL(),
								}),
								Title:       *cmdPack.Get("starboardCreating"),
								Color:       constants.Colors.Main,
								Description: *cmdPack.Get("starboardConfirm"),
								Timestamp:   json.Ptr(time.Now()),
								Fields: []discord.EmbedField{
									{
										Name: "\u0020",
										Value: *cmdPack.Get("starboardData") +
											*cmdPack.Getf("starboardDataChannel", data.ChannelId) +
											*cmdPack.Getf("starboardDataName", data.Name) +
											*cmdPack.Getf("starboardDataEmoji", emoji),
										Inline: json.Ptr(true),
									},
									{
										Name: "\u0020",
										Value: *cmdPack.Get("starboardRequisites") +
											*cmdPack.Getf("starboardRequisitesRequired", data.Required) +
											*cmdPack.Getf("starboardRequisitesBotsReact", utils.ReadableBool(
												&data.BotsReact,
												*langs.Pack(guildData.Lang).GetGlobal("yes"),
												*langs.Pack(guildData.Lang).GetGlobal("no"),
											)) +
											*cmdPack.Getf("starboardRequisitesBotsMsg", utils.ReadableBool(
												&data.BotsMessages,
												*langs.Pack(guildData.Lang).GetGlobal("yes"),
												*langs.Pack(guildData.Lang).GetGlobal("no"),
											)),
										Inline: json.Ptr(true),
									},
									{
										Name:   "\u0020",
										Value:  "\u0020",
										Inline: json.Ptr(true),
									},
									{
										Name: "\u0020",
										Value: *langs.Pack(guildData.Lang).
											Command("starboard").
											SubCommand("manual").
											Getf("starboardCustom", utils.ToString(data.EmbedColor)),
										Inline: json.Ptr(true),
									},
								},
							},
						}),
						Components: json.Ptr([]discord.ContainerComponent{
							discord.NewActionRow(
								discord.NewPrimaryButton(
									*langs.Pack(guildData.Lang).GetGlobal("yes"),
									starboard.YesButtonId+"/"+data.ID.Hex(),
								),
								discord.NewPrimaryButton(
									*langs.Pack(guildData.Lang).GetGlobal("no"),
									starboard.NoButtonId+"/"+data.ID.Hex(),
								),
							),
						}),
					})
				if err != nil {
					starboard.DeleteTempStarboard(data)
					log.Error().
						Err(err).
						Msg("Error ocurred when trying to respond in \"starboard:interactivo:skipbutton\"")

					return
				}

				utils.WaitDo(time.Second*80, func() {
					find := &models.TempStarboard{}
					err := models.TempStarboardColl().First(data, find)

					if err != mongo.ErrNoDocuments {
						err := models.TempStarboardColl().Delete(find)
						if err != nil {
							log.Error().
								Err(err).
								Msg("Error ocurred when trying to delete document in \"starboard:interactivo:skipbutton\"")
						}

						_, errM := e.Client().Rest().UpdateMessage(
							snowflake.MustParse(find.MsgChannelId),
							snowflake.MustParse(find.MessageId),
							discord.MessageUpdate{
								Content: cmdPack.Get("errTimeout"),
								Embeds:  json.Ptr([]discord.Embed{}),
								Components: json.Ptr(
									[]discord.ContainerComponent{},
								),
							},
						)
						if errM != nil {
							log.Error().
								Err(err).
								Msg("Error ocurred when trying to edit message in \"starboard:interactivo:skipbutton\"")
						}
					}
				})

				return
			}
		default:
			{
				starboard.DeleteTempStarboard(data)
				c.Rest().UpdateMessage(
					snowflake.MustParse(data.MsgChannelId),
					e.Message.ID,
					discord.MessageUpdate{
						Content:    cmdPack.Get("errNoValidPhase"),
						Embeds:     json.Ptr([]discord.Embed{}),
						Components: json.Ptr([]discord.ContainerComponent{}),
					})

				return
			}
		}
	})
}
