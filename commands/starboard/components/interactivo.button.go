package components

import (
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"
	"github.com/disgoorg/snowflake/v2"
	"github.com/rs/zerolog/log"
	"github.com/vyrekxd/kirby/commands/starboard"
	"github.com/vyrekxd/kirby/constants"
	"github.com/vyrekxd/kirby/database/models"
	"github.com/vyrekxd/kirby/langs"
	"github.com/vyrekxd/kirby/utils"
	"go.mongodb.org/mongo-driver/mongo"
)

func YesButton(e *handler.ComponentEvent) error {
	if e.Message.Author.ID != e.Client().ID() {
		return nil
	}

	guildData := models.GuildConfig{
		Lang: "es-MX",
	}
	err := models.GuildConfigColl().
		FindByID(e.GuildID().String(), &guildData)
	if err != nil {
		return nil
	}

	cmdPack := langs.Pack(guildData.Lang).
		Command("starboard").
		SubCommand("interactivo")

	data := &models.TempStarboard{}
	err = models.TempStarboardColl().FindByID(e.Variables["id"], data)
	if err == mongo.ErrNoDocuments {
		e.UpdateMessage(discord.MessageUpdate{
			Content:    cmdPack.Getf("errUnexpected", err),
			Embeds:     json.Ptr([]discord.Embed{}),
			Components: json.Ptr([]discord.ContainerComponent{}),
		})

		return nil

	} else if err != nil && err != mongo.ErrNoDocuments {
		e.UpdateMessage(discord.MessageUpdate{
			Content:    cmdPack.Getf("errUnexpected", err),
			Embeds:     json.Ptr([]discord.Embed{}),
			Components: json.Ptr([]discord.ContainerComponent{}),
		})

		return nil
	}

	if data.UserId != e.User().ID.String() {
		return nil
	}

	switch data.Phase {
	case models.PhaseModal:
		{
			data.BotsReact = true
			data.Phase = models.PhaseBotsReact
			err = models.TempStarboardColl().Update(data)
			if err != nil {
				starboard.DeleteTempStarboard(data)
				e.UpdateMessage(discord.MessageUpdate{
					Content:    cmdPack.Getf("errCantUpdate", err),
					Embeds:     json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				})

				return nil
			}

			err = e.UpdateMessage(discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name:    e.User().Username,
							IconURL: *e.User().AvatarURL(),
						}),
						Title: *cmdPack.Get("starboardCreating"),
						Color: constants.Colors.Main,
						Description: *cmdPack.Get("selectBotsReact") +
							*cmdPack.Get("useButtons"),
						Timestamp: json.Ptr(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardData") +
									*cmdPack.Getf("starboardDataChannel", data.ChannelId) +
									*cmdPack.Getf("starboardDataName", data.Name),
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
									)),
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
						discord.NewSecondaryButton(
							*cmdPack.Get("skip"),
							starboard.SkipButtonId+"/"+data.ID.Hex(),
						),
					),
				}),
			})
			if err != nil {
				starboard.DeleteTempStarboard(data)
				log.Error().
					Err(err).
					Msg("Error ocurred when trying to respond in \"starboard:interactivo:yesbutton\"")

				return nil
			}

			utils.WaitDo(time.Second*50, func() {
				find := &models.TempStarboard{}
				err := models.TempStarboardColl().First(data, find)

				if err == nil && find.Phase != models.PhaseBotsReact {
					err := models.TempStarboardColl().Delete(find)
					if err != nil {
						log.Error().
							Err(err).
							Msg("Error ocurred when trying to delete document in \"starboard:interactivo:yesbutton\"")
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
							Msg("Error ocurred when trying to edit message in \"starboard:interactivo:yesbutton\"")
					}
				}
			})

			return nil
		}

	case models.PhaseBotsReact:
		{
			data.BotsMessages = true
			data.Phase = models.PhaseBotsMessages
			err = models.TempStarboardColl().Update(data)
			if err != nil {
				starboard.DeleteTempStarboard(data)
				e.UpdateMessage(discord.MessageUpdate{
					Content:    cmdPack.Getf("errCantUpdate", err),
					Embeds:     json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				})

				return nil
			}

			err = e.UpdateMessage(discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name:    e.User().Username,
							IconURL: *e.User().AvatarURL(),
						}),
						Title:       *cmdPack.Get("starboardCreating"),
						Color:       constants.Colors.Main,
						Description: *cmdPack.Get("selectEmoji"),
						Timestamp:   json.Ptr(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardData") +
									*cmdPack.Getf("starboardDataChannel", data.ChannelId) +
									*cmdPack.Getf("starboardDataName", data.Name),
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
				Components: json.Ptr([]discord.ContainerComponent{}),
			})
			if err != nil {
				starboard.DeleteTempStarboard(data)
				log.Error().
					Err(err).
					Msg("Error ocurred when trying to respond in \"starboard:interactivo:yesbutton\"")

				return nil
			}

			utils.WaitDo(time.Second*50, func() {
				find := &models.TempStarboard{}
				err := models.TempStarboardColl().First(data, find)

				if err == nil && find.Emoji == "" {
					err := models.TempStarboardColl().Delete(find)
					if err != nil {
						log.Error().
							Err(err).
							Msg("Error ocurred when trying to delete document in \"starboard:interactivo:yesbutton\"")
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
							Msg("Error ocurred when trying to edit message in \"starboard:interactivo:yesbutton\"")
					}
				}
			})

			return nil
		}

	case models.PhaseEmbedColor:
		{
			err = models.TempStarboardColl().Delete(data)
			if err != nil {
				e.UpdateMessage(discord.MessageUpdate{
					Content:    cmdPack.Getf("errUnexpected", err),
					Embeds:     json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				})

				return nil
			}

			starboard := &models.Starboard{
				GuildId:      data.GuildId,
				Name:         data.Name,
				ChannelId:    data.ChannelId,
				Emoji:        data.Emoji,
				Required:     data.Required,
				BotsReact:    data.BotsReact,
				BotsMessages: data.BotsMessages,
				EmbedColor:   data.EmbedColor,
			}
			err = models.StarboardColl().Create(starboard)
			if err != nil {
				e.UpdateMessage(discord.MessageUpdate{
					Content:    cmdPack.Getf("errUnexpected", err),
					Embeds:     json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				})

				return nil
			}

			err = e.UpdateMessage(discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name:    e.Message.Author.Username,
							IconURL: *e.Message.Author.AvatarURL(),
						}),
						Title:       *cmdPack.Get("starboardCreated"),
						Color:       constants.Colors.Main,
						Description: *cmdPack.Get("starboardDesc"),
						Timestamp:   json.Ptr(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardData") +
									*cmdPack.Getf("starboardDataChannel", data.ChannelId) +
									*cmdPack.Getf("starboardDataName", data.Name) +
									*cmdPack.Getf("starboardDataEmoji", data.Emoji),
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
				Components: json.Ptr([]discord.ContainerComponent{}),
			})
			if err != nil {
				models.StarboardColl().Delete(starboard)
				log.Error().
					Err(err).
					Msg("Error ocurred when trying to respond in \"starboard:interactivo:skipbutton\"")

				return nil
			}

			return nil
		}

	default:
		{
			starboard.DeleteTempStarboard(data)
			e.UpdateMessage(discord.MessageUpdate{
				Content:    cmdPack.Get("errNoValidPhase"),
				Embeds:     json.Ptr([]discord.Embed{}),
				Components: json.Ptr([]discord.ContainerComponent{}),
			})
		}
	}

	return nil
}

func NoButton(e *handler.ComponentEvent) error {
	if e.Message.Author.ID != e.Client().ID() {
		return nil
	}

	guildData := models.GuildConfig{
		Lang: "es-MX",
	}
	err := models.GuildConfigColl().
		FindByID(e.GuildID().String(), &guildData)
	if err != nil {
		return nil
	}

	cmdPack := langs.Pack(guildData.Lang).
		Command("starboard").
		SubCommand("interactivo")

	data := &models.TempStarboard{}
	err = models.TempStarboardColl().FindByID(e.Variables["id"], data)
	if err == mongo.ErrNoDocuments {
		e.UpdateMessage(discord.MessageUpdate{
			Content:    cmdPack.Getf("errUnexpected", err),
			Embeds:     json.Ptr([]discord.Embed{}),
			Components: json.Ptr([]discord.ContainerComponent{}),
		})

		return nil

	} else if err != nil && err != mongo.ErrNoDocuments {
		e.UpdateMessage(discord.MessageUpdate{
			Content:    cmdPack.Getf("errUnexpected", err),
			Embeds:     json.Ptr([]discord.Embed{}),
			Components: json.Ptr([]discord.ContainerComponent{}),
		})

		return nil
	}

	if data.UserId != e.User().ID.String() {
		return nil
	}

	switch data.Phase {
	case models.PhaseModal:
		{
			data.BotsReact = false
			data.Phase = models.PhaseBotsReact
			err = models.TempStarboardColl().Update(data)
			if err != nil {
				starboard.DeleteTempStarboard(data)
				e.UpdateMessage(discord.MessageUpdate{
					Content:    cmdPack.Getf("errCantUpdate", err),
					Embeds:     json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				})

				return nil
			}

			err = e.UpdateMessage(discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name:    e.User().Username,
							IconURL: *e.User().AvatarURL(),
						}),
						Title: *cmdPack.Get("starboardCreating"),
						Color: constants.Colors.Main,
						Description: *cmdPack.Get("selectBotsMsg") +
							*cmdPack.Get("useButtons"),
						Timestamp: json.Ptr(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardData") +
									*cmdPack.Getf("starboardDataChannel", data.ChannelId) +
									*cmdPack.Getf("starboardDataName", data.Name),
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
									)),
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
						discord.NewSecondaryButton(
							*cmdPack.Get("skip"),
							starboard.SkipButtonId+"/"+data.ID.Hex(),
						),
					),
				}),
			})
			if err != nil {
				starboard.DeleteTempStarboard(data)
				log.Error().
					Err(err).
					Msg("Error ocurred when trying to respond in \"starboard:interactivo:nobutton\"")

				return nil
			}

			utils.WaitDo(time.Second*50, func() {
				find := &models.TempStarboard{}
				err := models.TempStarboardColl().First(data, find)

				if err == nil && find.Phase != models.PhaseBotsReact {
					err := models.TempStarboardColl().Delete(find)
					if err != nil {
						log.Error().
							Err(err).
							Msg("Error ocurred when trying to delete document in \"starboard:interactivo:nobutton\"")
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
							Msg("Error ocurred when trying to edit message in \"starboard:interactivo:nobutton\"")
					}
				}
			})

			return nil
		}

	case models.PhaseBotsReact:
		{
			data.BotsMessages = true
			data.Phase = models.PhaseBotsMessages
			err = models.TempStarboardColl().Update(data)
			if err != nil {
				starboard.DeleteTempStarboard(data)
				e.UpdateMessage(discord.MessageUpdate{
					Content:    cmdPack.Getf("errCantUpdate", err),
					Embeds:     json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				})

				return nil
			}

			err = e.UpdateMessage(discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name:    e.User().Username,
							IconURL: *e.User().AvatarURL(),
						}),
						Title:       *cmdPack.Get("starboardCreating"),
						Color:       constants.Colors.Main,
						Description: *cmdPack.Get("selectEmoji"),
						Timestamp:   json.Ptr(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardData") +
									*cmdPack.Getf("starboardDataChannel", data.ChannelId) +
									*cmdPack.Getf("starboardDataName", data.Name),
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
				Components: json.Ptr([]discord.ContainerComponent{}),
			})
			if err != nil {
				starboard.DeleteTempStarboard(data)
				log.Error().
					Err(err).
					Msg("Error ocurred when trying to respond in \"starboard:interactivo:nobutton\"")

				return nil
			}

			utils.WaitDo(time.Second*50, func() {
				find := &models.TempStarboard{}
				err := models.TempStarboardColl().First(data, find)

				if err == nil && find.Emoji == "" {
					err := models.TempStarboardColl().Delete(find)
					if err != nil {
						log.Error().
							Err(err).
							Msg("Error ocurred when trying to delete document in \"starboard:interactivo:nobutton\"")
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
							Msg("Error ocurred when trying to edit message in \"starboard:interactivo:nobutton\"")
					}
				}
			})

			return nil
		}

	case models.PhaseEmbedColor:
		{
			err = models.TempStarboardColl().Delete(data)
			if err != nil {
				e.UpdateMessage(discord.MessageUpdate{
					Content:    cmdPack.Getf("errUnexpected", err),
					Embeds:     json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				})

				return nil
			}

			e.UpdateMessage(discord.MessageUpdate{
				Content:    cmdPack.Getf("starboardCancel", err),
				Embeds:     json.Ptr([]discord.Embed{}),
				Components: json.Ptr([]discord.ContainerComponent{}),
			})

			return nil
		}

	default:
		{
			starboard.DeleteTempStarboard(data)
			e.UpdateMessage(discord.MessageUpdate{
				Content:    cmdPack.Get("errNoValidPhase"),
				Embeds:     json.Ptr([]discord.Embed{}),
				Components: json.Ptr([]discord.ContainerComponent{}),
			})
		}
	}

	return nil
}

func SkipButton(e *handler.ComponentEvent) error {
	if e.Message.Author.ID != e.Client().ID() {
		return nil
	}

	guildData := models.GuildConfig{
		Lang: "es-MX",
	}
	err := models.GuildConfigColl().
		FindByID(e.GuildID().String(), &guildData)
	if err != nil {
		return nil
	}

	cmdPack := langs.Pack(guildData.Lang).
		Command("starboard").
		SubCommand("interactivo")

	data := &models.TempStarboard{}
	err = models.TempStarboardColl().FindByID(e.Variables["id"], data)
	if err == mongo.ErrNoDocuments {
		e.UpdateMessage(discord.MessageUpdate{
			Content:    cmdPack.Getf("errUnexpected", err),
			Embeds:     json.Ptr([]discord.Embed{}),
			Components: json.Ptr([]discord.ContainerComponent{}),
		})

		return nil

	} else if err != nil && err != mongo.ErrNoDocuments {
		e.UpdateMessage(discord.MessageUpdate{
			Content:    cmdPack.Getf("errUnexpected", err),
			Embeds:     json.Ptr([]discord.Embed{}),
			Components: json.Ptr([]discord.ContainerComponent{}),
		})

		return nil
	}

	switch data.Phase {
	case models.PhaseModal:
		{
			data.Phase = models.PhaseBotsReact
			err = models.TempStarboardColl().Update(data)
			if err != nil {
				starboard.DeleteTempStarboard(data)
				e.UpdateMessage(discord.MessageUpdate{
					Content:    cmdPack.Getf("errCantUpdate", err),
					Embeds:     json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				})

				return nil
			}

			err = e.UpdateMessage(discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name:    e.User().Username,
							IconURL: *e.User().AvatarURL(),
						}),
						Title: *cmdPack.Get("starboardCreating"),
						Color: constants.Colors.Main,
						Description: *cmdPack.Get("selectBotsMsg") +
							*cmdPack.Get("useButtons"),
						Timestamp: json.Ptr(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardData") +
									*cmdPack.Getf("starboardDataChannel", data.ChannelId) +
									*cmdPack.Getf("starboardDataName", data.Name),
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
									)),
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
						discord.NewSecondaryButton(
							*cmdPack.Get("skip"),
							starboard.SkipButtonId+"/"+data.ID.Hex(),
						),
					),
				}),
			})
			if err != nil {
				starboard.DeleteTempStarboard(data)
				log.Error().
					Err(err).
					Msg("Error ocurred when trying to respond in \"starboard:interactivo:skipbutton\"")

				return nil
			}

			utils.WaitDo(time.Second*50, func() {
				find := &models.TempStarboard{}
				err := models.TempStarboardColl().First(data, find)

				if err == nil && find.Phase != models.PhaseBotsReact {
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

			return nil
		}

	case models.PhaseBotsReact:
		{
			data.Phase = models.PhaseBotsMessages
			err = models.TempStarboardColl().Update(data)
			if err != nil {
				starboard.DeleteTempStarboard(data)
				e.UpdateMessage(discord.MessageUpdate{
					Content:    cmdPack.Getf("errCantUpdate", err),
					Embeds:     json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				})

				return nil
			}

			err = e.UpdateMessage(discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name:    e.User().Username,
							IconURL: *e.User().AvatarURL(),
						}),
						Title:       *cmdPack.Get("starboardCreating"),
						Color:       constants.Colors.Main,
						Description: *cmdPack.Get("selectEmoji"),
						Timestamp:   json.Ptr(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardData") +
									*cmdPack.Getf("starboardDataChannel", data.ChannelId) +
									*cmdPack.Getf("starboardDataName", data.Name),
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
				Components: json.Ptr([]discord.ContainerComponent{}),
			})
			if err != nil {
				starboard.DeleteTempStarboard(data)
				log.Error().
					Err(err).
					Msg("Error ocurred when trying to respond in \"starboard:interactivo:skipbutton\"")

				return nil
			}

			utils.WaitDo(time.Second*50, func() {
				find := &models.TempStarboard{}
				err := models.TempStarboardColl().First(data, find)

				if err == nil && find.Emoji == "" {
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

			return nil
		}

	case models.PhaseEmoji:
		{
			data.EmbedColor = constants.Colors.Main
			data.Phase = models.PhaseEmbedColor
			err = models.TempStarboardColl().Update(data)
			if err != nil {
				starboard.DeleteTempStarboard(data)
				e.UpdateMessage(discord.MessageUpdate{
					Content:    cmdPack.Getf("errCantUpdate", err),
					Embeds:     json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				})

				return nil
			}

			err = e.UpdateMessage(discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name:    e.User().Username,
							IconURL: *e.User().AvatarURL(),
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
									*cmdPack.Getf("starboardDataEmoji", data.Emoji),
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

				return nil
			}

			utils.WaitDo(time.Second*50, func() {
				find := &models.TempStarboard{}
				err := models.TempStarboardColl().First(data, find)

				if err == nil && find.Phase != models.PhaseConfirm {
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

			return nil
		}

	default:
		{
			starboard.DeleteTempStarboard(data)
			e.UpdateMessage(discord.MessageUpdate{
				Content:    cmdPack.Get("errNoValidPhase"),
				Embeds:     json.Ptr([]discord.Embed{}),
				Components: json.Ptr([]discord.ContainerComponent{}),
			})
		}
	}

	return nil
}
