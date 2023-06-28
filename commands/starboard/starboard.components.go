package starboard

import (
	"context"
	"strconv"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"
	"github.com/disgoorg/snowflake/v2"

	"github.com/rs/zerolog/log"

	"github.com/vyrekxd/kirby/constants"
	"github.com/vyrekxd/kirby/database/models"
	"github.com/vyrekxd/kirby/langs"
	"github.com/vyrekxd/kirby/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func SelectChannel(e *handler.ComponentEvent) error {
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

	// how can i convert the string ID to objectid

	starboards := []models.Starboard{}
	err = models.StarboardColl().
		SimpleFind(&starboards, bson.M{"guild_id": e.GuildID().String()})
	if err != nil && err != mongo.ErrNoDocuments {
		e.UpdateMessage(discord.MessageUpdate{
			Content: langs.Pack(guildData.Lang).
				Command("starboard").
				SubCommand("manual").
				Getf("errFindGuildStarboards", err),
		})

		return nil
	}

	data := &models.TempStarboard{}
	err = models.TempStarboardColl().First(models.TempStarboard{
		GuildId: e.GuildID().String(),
	}, data)
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

	if data.UserId == e.User().ID.String() {
		return nil
	}

	menuData := e.ChannelSelectMenuInteractionData()
	if len(menuData.Channels()) < 1 {
		DeleteTempStarboard(data)
		e.UpdateMessage(discord.MessageUpdate{
			Content:    cmdPack.Get("errNoChannel"),
			Embeds:     json.Ptr([]discord.Embed{}),
			Components: json.Ptr([]discord.ContainerComponent{}),
		})

		return nil
	}

	channel := menuData.Channels()[0]
	for _, s := range starboards {
		if s.ChannelId == channel.ID.String() {
			e.UpdateMessage(discord.MessageUpdate{
				Content:    cmdPack.Get("channelAlreadyUsed"),
				Embeds:     json.Ptr([]discord.Embed{}),
				Components: json.Ptr([]discord.ContainerComponent{}),
			})

			return nil
		} else {
			for _, cid := range s.ChannelList {
				if cid == channel.ID.String() {
					e.UpdateMessage(discord.MessageUpdate{
						Content:    cmdPack.Get("channelInList"),
						Embeds:     json.Ptr([]discord.Embed{}),
						Components: json.Ptr([]discord.ContainerComponent{}),
					})

					return nil
				}
			}
		}
	}

	data.ChannelId = channel.ID.String()
	data.Phase = models.PhaseSelectChannel
	_, err = models.TempStarboardColl().
		UpdateOne(context.Background(), bson.M{"_id": data.ID}, bson.M{"$set": data})
	if err != nil {
		DeleteTempStarboard(data)
		e.UpdateMessage(discord.MessageUpdate{
			Content:    cmdPack.Getf("errCantUpdate", err),
			Embeds:     json.Ptr([]discord.Embed{}),
			Components: json.Ptr([]discord.ContainerComponent{}),
		})

		return nil
	}

	_, err = e.Client().Rest().UpdateMessage(
		snowflake.MustParse(data.MsgChannelId),
		snowflake.MustParse(data.MessageId),
		discord.MessageUpdate{
			Embeds: json.Ptr([]discord.Embed{
				{
					Author: json.Ptr(discord.EmbedAuthor{
						Name:    e.User().Username,
						IconURL: *e.User().AvatarURL(),
					}),
					Title:       *cmdPack.Get("starboardCreating"),
					Color:       constants.Colors.Main,
					Description: *cmdPack.Get("respondModal"),
					Timestamp:   json.Ptr(time.Now()),
					Fields: []discord.EmbedField{
						{
							Name: "\u0020",
							Value: *cmdPack.Get("starboardData") +
								*cmdPack.Getf("starboardDataChannel", channel.ID),
						},
					},
				},
			}),
			Components: json.Ptr([]discord.ContainerComponent{}),
		})
	if err != nil {
		DeleteTempStarboard(data)
		log.Error().
			Err(err).
			Msg("Error ocurred when trying to respond in \"starboard:interactivo:channel\"")

		return nil
	}

	err = e.CreateModal(discord.ModalCreate{
		CustomID: ModalId,
		Title:    *cmdPack.Get("starboardCreating"),
		Components: []discord.ContainerComponent{
			discord.ActionRowComponent{
				discord.TextInputComponent{
					CustomID:    NameInputId,
					Label:       *cmdPack.Get("labelName") + *cmdPack.Get("labelOptional"),
					Placeholder: *cmdPack.Get("placeholderName"),
					Style:       discord.TextInputStyleShort,
					MinLength:   json.Ptr(4),
					MaxLength:   25,
					Required:    false,
				},
			},
			discord.ActionRowComponent{
				discord.TextInputComponent{
					CustomID:    RequiredInputId,
					Label:       *cmdPack.Get("labelRequired"),
					Placeholder: *cmdPack.Get("placeholderRequired"),
					Style:       discord.TextInputStyleShort,
					MinLength:   json.Ptr(1),
					MaxLength:   2,
					Required:    true,
				},
			},
		},
	})
	if err != nil {
		DeleteTempStarboard(data)
		e.UpdateMessage(discord.MessageUpdate{
			Content:    cmdPack.Getf("errUnexpected", err),
			Embeds:     json.Ptr([]discord.Embed{}),
			Components: json.Ptr([]discord.ContainerComponent{}),
		})

		return nil
	}

	utils.WaitDo(time.Second*50, func() {
		find := &models.TempStarboard{}
		models.TempStarboardColl().First(data, find)

		if find.Required == 0 {
			err := models.TempStarboardColl().Delete(find)
			if err != nil {
				log.Error().
					Err(err).
					Msg("Error ocurred when trying to delete document in \"starboard:interactivo\"")
			}

			_, errM := e.Client().Rest().UpdateMessage(
				snowflake.MustParse(find.MsgChannelId),
				snowflake.MustParse(find.MessageId),
				discord.MessageUpdate{
					Content:    cmdPack.Get("errTimeout"),
					Embeds:     json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				},
			)
			if errM != nil {
				log.Error().
					Err(err).
					Msg("Error ocurred when trying to edit message in \"starboard:interactivo\"")
			}
		}
	})

	return nil
}

func Modal(e *handler.ModalEvent) error {
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
	res := models.TempStarboardColl().
		FindOne(context.Background(), bson.M{"guild_id": e.GuildID().String()})
	err = res.Decode(data)
	if err != nil {
		DeleteTempStarboard(data)
		e.UpdateMessage(discord.MessageUpdate{
			Content:    cmdPack.Getf("errUnexpected", err),
			Embeds:     json.Ptr([]discord.Embed{}),
			Components: json.Ptr([]discord.ContainerComponent{}),
		})

		return nil

	}
	if data.UserId == e.User().ID.String() {
		return nil
	}

	name := e.Data.Text(NameInputId)
	if name == "" {
		ch, err := e.Client().
			Rest().
			GetChannel(snowflake.MustParse(data.ChannelId))
		if err != nil {
			DeleteTempStarboard(data)
			e.UpdateMessage(discord.MessageUpdate{
				Content:    cmdPack.Getf("errUnexpected", err),
				Embeds:     json.Ptr([]discord.Embed{}),
				Components: json.Ptr([]discord.ContainerComponent{}),
			})

			return nil
		}

		name = ch.Name()
	}

	requiredStr := e.Data.Text(RequiredInputId)
	required, err := strconv.Atoi(requiredStr)
	if err != nil {
		DeleteTempStarboard(data)
		e.UpdateMessage(discord.MessageUpdate{
			Content:    cmdPack.Get("errNoValidNumber"),
			Embeds:     json.Ptr([]discord.Embed{}),
			Components: json.Ptr([]discord.ContainerComponent{}),
		})

		return nil
	}

	if required < 1 {
		DeleteTempStarboard(data)
		e.UpdateMessage(discord.MessageUpdate{
			Content:    cmdPack.Get("errNoValidRequired"),
			Embeds:     json.Ptr([]discord.Embed{}),
			Components: json.Ptr([]discord.ContainerComponent{}),
		})

		return nil
	}

	data.Name = name
	data.Required = required
	data.Phase = models.PhaseModal
	_, err = models.TempStarboardColl().
		UpdateOne(context.TODO(), bson.M{"guild_id": e.GuildID().String()}, bson.M{"$set": data})
	if err != nil {
		DeleteTempStarboard(data)
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
							*cmdPack.Getf("starboardDataName", name),
					},
					{
						Name: "\u0020",
						Value: *cmdPack.Get("starboardRequisites") +
							*cmdPack.Getf("starboardRequisitesRequired", required),
					},
				},
			},
		}),
		Components: json.Ptr([]discord.ContainerComponent{
			discord.NewActionRow(
				discord.NewPrimaryButton(
					*langs.Pack(guildData.Lang).GetGlobal("yes"),
					YesButtonId,
				),
				discord.NewPrimaryButton(
					*langs.Pack(guildData.Lang).GetGlobal("no"),
					NoButtonId,
				),
				discord.NewSecondaryButton(
					*cmdPack.Get("skip"),
					SkipButtonId,
				),
			),
		}),
	})

	if err != nil {
		DeleteTempStarboard(data)
		log.Error().
			Err(err).
			Msg("Error ocurred when trying to respond in \"starboard:interactivo:channel\"")

		return nil
	}

	utils.WaitDo(time.Second*50, func() {
		find := &models.TempStarboard{}
		models.TempStarboardColl().First(data, find)

		if find.Phase != models.PhaseBotsMessages {
			err := models.TempStarboardColl().Delete(find)
			if err != nil {
				log.Error().
					Err(err).
					Msg("Error ocurred when trying to delete document in \"starboard:interactivo\"")
			}

			_, errM := e.Client().Rest().UpdateMessage(
				snowflake.MustParse(find.MsgChannelId),
				snowflake.MustParse(find.MessageId),
				discord.MessageUpdate{
					Content:    cmdPack.Get("errTimeout"),
					Embeds:     json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				},
			)
			if errM != nil {
				log.Error().
					Err(err).
					Msg("Error ocurred when trying to edit message in \"starboard:interactivo\"")
			}
		}
	})

	return nil
}

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
	err = models.TempStarboardColl().First(models.TempStarboard{
		GuildId: e.GuildID().String(),
	}, data)
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

	if data.UserId == e.User().ID.String() {
		return nil
	}

	switch data.Phase {
	case models.PhaseModal:
		{
			data.BotsMessages = true
			data.Phase = models.PhaseBotsMessages
			_, err = models.TempStarboardColl().
				UpdateOne(context.Background(), bson.M{"_id": data.ID}, bson.M{"$set": data})
			if err != nil {
				DeleteTempStarboard(data)
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
							},
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardRequisites") +
									*cmdPack.Getf("starboardRequisitesRequired", data.Required) +
									*cmdPack.Getf("starboardRequisitesBotsReact", utils.ReadableBool(
										&data.BotsMessages,
										*langs.Pack(guildData.Lang).GetGlobal("yes"),
										*langs.Pack(guildData.Lang).GetGlobal("no"),
									),
									),
							},
						},
					},
				}),
				Components: json.Ptr([]discord.ContainerComponent{
					discord.NewActionRow(
						discord.NewPrimaryButton(
							*langs.Pack(guildData.Lang).GetGlobal("yes"),
							YesButtonId,
						),
						discord.NewPrimaryButton(
							*langs.Pack(guildData.Lang).GetGlobal("no"),
							NoButtonId,
						),
						discord.NewSecondaryButton(
							*cmdPack.Get("skip"),
							SkipButtonId,
						),
					),
				}),
			})
			if err != nil {
				DeleteTempStarboard(data)
				log.Error().
					Err(err).
					Msg("Error ocurred when trying to respond in \"starboard:interactivo:yesbutton\"")

				return nil
			}

			utils.WaitDo(time.Second*50, func() {
				find := &models.TempStarboard{}
				models.TempStarboardColl().First(data, find)

				if find.Phase != models.PhaseBotsReact {
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
							Msg("Error ocurred when trying to edit message in \"starboard:interactivo\"")
					}
				}
			})

			return nil
		}

	case models.PhaseBotsMessages:
		{
			data.BotsReact = true
			data.Phase = models.PhaseBotsReact
			_, err = models.TempStarboardColl().
				UpdateOne(context.Background(), bson.M{"_id": data.ID}, bson.M{"$set": data})
			if err != nil {
				DeleteTempStarboard(data)
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
							},
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardRequisites") +
									*cmdPack.Getf("starboardRequisitesRequired", data.Required) +
									*cmdPack.Getf("starboardRequisitesBotsReact", utils.ReadableBool(
										&data.BotsMessages,
										*langs.Pack(guildData.Lang).GetGlobal("yes"),
										*langs.Pack(guildData.Lang).GetGlobal("no"),
									)) +
									*cmdPack.Getf("starboardRequisitesBotsMsg", utils.ReadableBool(
										&data.BotsMessages,
										*langs.Pack(guildData.Lang).GetGlobal("yes"),
										*langs.Pack(guildData.Lang).GetGlobal("no"),
									)),
							},
						},
					},
				}),
				Components: json.Ptr([]discord.ContainerComponent{
					discord.NewActionRow(
						discord.NewPrimaryButton(
							*langs.Pack(guildData.Lang).GetGlobal("yes"),
							YesButtonId,
						),
						discord.NewPrimaryButton(
							*langs.Pack(guildData.Lang).GetGlobal("no"),
							NoButtonId,
						),
						discord.NewSecondaryButton(
							*cmdPack.Get("skip"),
							SkipButtonId,
						),
					),
				}),
			})
			if err != nil {
				DeleteTempStarboard(data)
				log.Error().
					Err(err).
					Msg("Error ocurred when trying to respond in \"starboard:interactivo:yesbutton\"")

				return nil
			}

			utils.WaitDo(time.Second*50, func() {
				find := &models.TempStarboard{}
				models.TempStarboardColl().First(data, find)

				if find.Emoji == "" {
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
							Msg("Error ocurred when trying to edit message in \"starboard:interactivo\"")
					}
				}
			})

			return nil
		}

	default:
		{
			DeleteTempStarboard(data)
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
	err = models.TempStarboardColl().First(models.TempStarboard{
		GuildId: e.GuildID().String(),
	}, data)
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

	if data.UserId == e.User().ID.String() {
		return nil
	}

	switch data.Phase {
	case models.PhaseModal:
		{
			data.BotsMessages = false
			data.Phase = models.PhaseBotsMessages
			_, err = models.TempStarboardColl().
				UpdateOne(context.Background(), bson.M{"_id": data.ID}, bson.M{"$set": data})
			if err != nil {
				DeleteTempStarboard(data)
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
							},
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardRequisites") +
									*cmdPack.Getf("starboardRequisitesRequired", data.Required) +
									*cmdPack.Getf("starboardRequisitesBotsReact", utils.ReadableBool(
										&data.BotsMessages,
										*langs.Pack(guildData.Lang).GetGlobal("yes"),
										*langs.Pack(guildData.Lang).GetGlobal("no"),
									)),
							},
						},
					},
				}),
				Components: json.Ptr([]discord.ContainerComponent{
					discord.NewActionRow(
						discord.NewPrimaryButton(
							*langs.Pack(guildData.Lang).GetGlobal("yes"),
							YesButtonId,
						),
						discord.NewPrimaryButton(
							*langs.Pack(guildData.Lang).GetGlobal("no"),
							NoButtonId,
						),
						discord.NewSecondaryButton(
							*cmdPack.Get("skip"),
							SkipButtonId,
						),
					),
				}),
			})
			if err != nil {
				DeleteTempStarboard(data)
				log.Error().
					Err(err).
					Msg("Error ocurred when trying to respond in \"starboard:interactivo:yesbutton\"")

				return nil
			}

			utils.WaitDo(time.Second*50, func() {
				find := &models.TempStarboard{}
				models.TempStarboardColl().First(data, find)

				if find.Phase != models.PhaseBotsReact {
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
							Msg("Error ocurred when trying to edit message in \"starboard:interactivo\"")
					}
				}
			})

			return nil
		}

	case models.PhaseBotsMessages:
		{
			data.BotsReact = true
			data.Phase = models.PhaseBotsReact
			_, err = models.TempStarboardColl().
				UpdateOne(context.Background(), bson.M{"_id": data.ID}, bson.M{"$set": data})
			if err != nil {
				DeleteTempStarboard(data)
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
							},
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardRequisites") +
									*cmdPack.Getf("starboardRequisitesRequired", data.Required) +
									*cmdPack.Getf("starboardRequisitesBotsReact", utils.ReadableBool(
										&data.BotsMessages,
										*langs.Pack(guildData.Lang).GetGlobal("yes"),
										*langs.Pack(guildData.Lang).GetGlobal("no"),
									)) +
									*cmdPack.Getf("starboardRequisitesBotsMsg", utils.ReadableBool(
										&data.BotsMessages,
										*langs.Pack(guildData.Lang).GetGlobal("yes"),
										*langs.Pack(guildData.Lang).GetGlobal("no"),
									)),
							},
						},
					},
				}),
				Components: json.Ptr([]discord.ContainerComponent{
					discord.NewActionRow(
						discord.NewPrimaryButton(
							*langs.Pack(guildData.Lang).GetGlobal("yes"),
							YesButtonId,
						),
						discord.NewPrimaryButton(
							*langs.Pack(guildData.Lang).GetGlobal("no"),
							NoButtonId,
						),
						discord.NewSecondaryButton(
							*cmdPack.Get("skip"),
							SkipButtonId,
						),
					),
				}),
			})
			if err != nil {
				DeleteTempStarboard(data)
				log.Error().
					Err(err).
					Msg("Error ocurred when trying to respond in \"starboard:interactivo:yesbutton\"")

				return nil
			}

			utils.WaitDo(time.Second*50, func() {
				find := &models.TempStarboard{}
				models.TempStarboardColl().First(data, find)

				if find.Emoji == "" {
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
							Msg("Error ocurred when trying to edit message in \"starboard:interactivo\"")
					}
				}
			})

			return nil
		}

	default:
		{
			DeleteTempStarboard(data)
			e.UpdateMessage(discord.MessageUpdate{
				Content:    cmdPack.Get("errNoValidPhase"),
				Embeds:     json.Ptr([]discord.Embed{}),
				Components: json.Ptr([]discord.ContainerComponent{}),
			})
		}
	}

	return nil
}

func OmitButton(e *handler.ComponentEvent) error {
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
	err = models.TempStarboardColl().First(models.TempStarboard{
		GuildId: e.GuildID().String(),
	}, data)
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
			data.Phase = models.PhaseBotsMessages
			_, err = models.TempStarboardColl().
				UpdateOne(context.Background(), bson.M{"_id": data.ID}, bson.M{"$set": data})
			if err != nil {
				DeleteTempStarboard(data)
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
							},
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardRequisites") +
									*cmdPack.Getf("starboardRequisitesRequired", data.Required) +
									*cmdPack.Getf("starboardRequisitesBotsReact", utils.ReadableBool(
										&data.BotsMessages,
										*langs.Pack(guildData.Lang).GetGlobal("yes"),
										*langs.Pack(guildData.Lang).GetGlobal("no"),
									)),
							},
						},
					},
				}),
				Components: json.Ptr([]discord.ContainerComponent{
					discord.NewActionRow(
						discord.NewPrimaryButton(
							*langs.Pack(guildData.Lang).GetGlobal("yes"),
							YesButtonId,
						),
						discord.NewPrimaryButton(
							*langs.Pack(guildData.Lang).GetGlobal("no"),
							NoButtonId,
						),
						discord.NewSecondaryButton(
							*cmdPack.Get("skip"),
							SkipButtonId,
						),
					),
				}),
			})
			if err != nil {
				DeleteTempStarboard(data)
				log.Error().
					Err(err).
					Msg("Error ocurred when trying to respond in \"starboard:interactivo:yesbutton\"")

				return nil
			}

			utils.WaitDo(time.Second*50, func() {
				find := &models.TempStarboard{}
				models.TempStarboardColl().First(data, find)

				if find.Phase != models.PhaseBotsReact {
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
							Msg("Error ocurred when trying to edit message in \"starboard:interactivo\"")
					}
				}
			})

			return nil
		}

	case models.PhaseBotsMessages:
		{
			data.Phase = models.PhaseBotsReact
			_, err = models.TempStarboardColl().
				UpdateOne(context.Background(), bson.M{"_id": data.ID}, bson.M{"$set": data})
			if err != nil {
				DeleteTempStarboard(data)
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
							},
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardRequisites") +
									*cmdPack.Getf("starboardRequisitesRequired", data.Required) +
									*cmdPack.Getf("starboardRequisitesBotsReact", utils.ReadableBool(
										&data.BotsMessages,
										*langs.Pack(guildData.Lang).GetGlobal("yes"),
										*langs.Pack(guildData.Lang).GetGlobal("no"),
									)) +
									*cmdPack.Getf("starboardRequisitesBotsMsg", utils.ReadableBool(
										&data.BotsMessages,
										*langs.Pack(guildData.Lang).GetGlobal("yes"),
										*langs.Pack(guildData.Lang).GetGlobal("no"),
									)),
							},
						},
					},
				}),
				Components: json.Ptr([]discord.ContainerComponent{
					discord.NewActionRow(
						discord.NewPrimaryButton(
							*langs.Pack(guildData.Lang).GetGlobal("yes"),
							YesButtonId,
						),
						discord.NewPrimaryButton(
							*langs.Pack(guildData.Lang).GetGlobal("no"),
							NoButtonId,
						),
						discord.NewSecondaryButton(
							*cmdPack.Get("skip"),
							SkipButtonId,
						),
					),
				}),
			})
			if err != nil {
				DeleteTempStarboard(data)
				log.Error().
					Err(err).
					Msg("Error ocurred when trying to respond in \"starboard:interactivo:yesbutton\"")

				return nil
			}

			utils.WaitDo(time.Second*50, func() {
				find := &models.TempStarboard{}
				models.TempStarboardColl().First(data, find)

				if find.Emoji == "" {
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
							Msg("Error ocurred when trying to edit message in \"starboard:interactivo\"")
					}
				}
			})

			return nil
		}

	default:
		{
			DeleteTempStarboard(data)
			e.UpdateMessage(discord.MessageUpdate{
				Content:    cmdPack.Get("errNoValidPhase"),
				Embeds:     json.Ptr([]discord.Embed{}),
				Components: json.Ptr([]discord.ContainerComponent{}),
			})
		}
	}

	return nil
}
