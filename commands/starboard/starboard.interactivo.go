package starboard

import (
	"strconv"
	"time"

	"golang.org/x/exp/slices"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/vyrekxd/kirby/commands/starboard/cmd_util"
	"github.com/vyrekxd/kirby/constants"
	"github.com/vyrekxd/kirby/database/models"
	"github.com/vyrekxd/kirby/langs"
	"github.com/vyrekxd/kirby/utils"
)

func starboardInteractivo(ctx *handler.CommandEvent) error {
	go func() {
		guildData := models.GuildConfig{Lang: "es-MX"}
		starboards := []models.Starboard{}
		err := models.GuildConfigColl().
			FindByID(ctx.GuildID().String(), &guildData)
		if err == mongo.ErrNoDocuments {
			ctx.UpdateInteractionResponse(discord.MessageUpdate{
				Content: langs.Pack(guildData.Lang).
					Command("starboard").
					SubCommand("manual").
					Getf("errNoGuildData", err),
			})

			return
		} else {
			err := models.StarboardColl().SimpleFind(&starboards, bson.M{"guild_id": ctx.GuildID().String()})
			if err != nil && err != mongo.ErrNoDocuments {
				ctx.UpdateInteractionResponse(discord.MessageUpdate{
					Content: langs.Pack(guildData.Lang).Command("starboard").SubCommand("manual").Getf("errFindGuildStarboards", err),
				})

				return
			}
		}

		if slices.Contains(constants.CurrentServersInSetup, ctx.GuildID().String()) {
			ctx.UpdateInteractionResponse(discord.MessageUpdate{
				Content: langs.Pack(guildData.Lang).Command("starboard").Get("alreadyOnSetup"),
			})
	
			return
		} else {
			constants.CurrentServersInSetup = append(constants.CurrentServersInSetup, ctx.GuildID().String())
	
			defer func() {
				i := slices.Index(constants.CurrentServersInSetup, ctx.GuildID().String())
	
				constants.CurrentServersInSetup = slices.Delete(
					constants.CurrentServersInSetup,
					slices.Index(constants.CurrentServersInSetup, ctx.GuildID().String()),
					i + 1,
				)
			}()
		}

		cmdPack := langs.Pack(guildData.Lang).
			Command("starboard").
			SubCommand("interactivo")

		msg, err := ctx.UpdateInteractionResponse(discord.MessageUpdate{
			Embeds: json.Ptr([]discord.Embed{
				{
					Author: json.Ptr(discord.EmbedAuthor{
						Name:    ctx.User().Username,
						IconURL: *ctx.User().AvatarURL(),
					}),
					Title:       *cmdPack.Get("starboardCreating"),
					Color:       constants.Colors.Main,
					Description: *cmdPack.Get("selectChannel"),
					Timestamp:   json.Ptr(time.Now()),
				},
			}),
			Components: json.Ptr([]discord.ContainerComponent{
				discord.NewActionRow(
					discord.ChannelSelectMenuComponent{
						CustomID:  "starboard:channel",
						MaxValues: 1,
						ChannelTypes: []discord.ComponentType{
							discord.ComponentType(discord.ChannelTypeGuildText),
							discord.ComponentType(discord.ChannelTypeGuildNews),
						},
					},
				),
			}),
		})
		if err != nil {
			log.Error().
				Err(err).
				Msg("Error ocurred when trying to respond in \"starboard:interactivo\"")
			return
		}

		log.Print("before getchannel")
		newCtx, channel, err := cmd_util.GetChannel(ctx, msg, cmdPack)
		if err != nil {
			return
		}

		_, err = ctx.Client().Rest().UpdateMessage(
			msg.ChannelID,
			msg.ID,
			discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name:    ctx.User().Username,
							IconURL: *ctx.User().AvatarURL(),
						}),
						Title:       *cmdPack.Get("starboardCreating"),
						Color:       constants.Colors.Main,
						Description: *cmdPack.Get("selectEmoji"),
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
			},
		)
		if err != nil {
			log.Error().
				Err(err).
				Msg(`Error when trying to update message in "starboard:interactivo"`)
			return
		}

		emoji, err := cmd_util.GetEmoji(newCtx, msg, cmdPack)
		if err != nil {
			return
		}
		
		ctx.Client().Rest().UpdateMessage(
			msg.ChannelID,
			msg.ID,
			discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name:    ctx.User().Username,
							IconURL: *ctx.User().AvatarURL(),
						}),
						Title: *cmdPack.Get("starboardCreating"),
						Color: constants.Colors.Main,
						Description: *cmdPack.Get("selectName") +
							*cmdPack.Get("optionalParam"),
						Timestamp: json.Ptr(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardData") +
									*cmdPack.Getf("starboardDataChannel", channel.ID) +
									*cmdPack.Getf("starboardDataEmoji", emoji),
							},
						},
					},
				}),
				Components: json.Ptr([]discord.ContainerComponent{
					discord.NewActionRow(
						discord.NewSecondaryButton(*cmdPack.Get("skip"), "starboard:skip"),
					),
				}),
			},
		)

		name, err := cmd_util.GetName(newCtx, msg, cmdPack)
		if err != nil {
			return
		}

		ctx.Client().Rest().UpdateMessage(
			msg.ChannelID,
			msg.ID,
			discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name:    ctx.User().Username,
							IconURL: *ctx.User().AvatarURL(),
						}),
						Title:       *cmdPack.Get("starboardCreating"),
						Color:       constants.Colors.Main,
						Description: *cmdPack.Get("selectRequired"),
						Timestamp:   json.Ptr(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardData") +
									*cmdPack.Getf("starboardDataChannel", channel.ID) +
									*cmdPack.Getf("starboardDataEmoji", emoji) +
									*cmdPack.Getf("starboardDataName", name),
							},
						},
					},
				}),
				Components: json.Ptr([]discord.ContainerComponent{}),
			},
		)

		required, err := cmd_util.GetInt(newCtx, msg, cmdPack)
		if err != nil {
			return
		} else if required <= 0 {
			ctx.Client().Rest().UpdateMessage(
				msg.ChannelID,
				msg.ID,
				discord.MessageUpdate{
					Content:    cmdPack.Get("errNoValidRequired"),
					Embeds:     json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				},
			)

			return
		}

		ctx.Client().Rest().UpdateMessage(
			msg.ChannelID,
			msg.ID,
			discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name:    ctx.User().Username,
							IconURL: *ctx.User().AvatarURL(),
						}),
						Title:       *cmdPack.Get("starboardCreating"),
						Color:       constants.Colors.Main,
						Description: *cmdPack.Get("selectRequiredToS"),
						Timestamp:   json.Ptr(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardData") +
									*cmdPack.Getf("starboardDataChannel", channel.ID) +
									*cmdPack.Getf("starboardDataEmoji", emoji) +
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
				Components: json.Ptr([]discord.ContainerComponent{}),
			},
		)

		requiredToStay, err := cmd_util.GetInt(newCtx, msg, cmdPack)
		if err != nil {
			return
		} else if requiredToStay < 0 || requiredToStay >= required {
			ctx.Client().Rest().UpdateMessage(
				msg.ChannelID,
				msg.ID,
				discord.MessageUpdate{
					Content:    cmdPack.Get("errNoValidRequiredToS"),
					Embeds:     json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				},
			)

			return
		}

		ctx.Client().Rest().UpdateMessage(
			msg.ChannelID,
			msg.ID,
			discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name:    ctx.User().Username,
							IconURL: *ctx.User().AvatarURL(),
						}),
						Title: *cmdPack.Get("starboardCreating"),
						Color: constants.Colors.Main,
						Description: *cmdPack.Get("selectBotsReact") +
							*cmdPack.Get("optionalParam") +
							*cmdPack.Get("useButtons"),
						Timestamp: json.Ptr(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardData") +
									*cmdPack.Getf("starboardDataChannel", channel.ID) +
									*cmdPack.Getf("starboardDataEmoji", emoji) +
									*cmdPack.Getf("starboardDataName", name),
							},
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardRequisites") +
									*cmdPack.Getf("starboardRequisitesRequired", required) +
									*cmdPack.Getf("starboardRequisitesRequiredToS", requiredToStay),
							},
						},
					},
				}),
				Components: json.Ptr([]discord.ContainerComponent{
					discord.NewActionRow(
						discord.NewPrimaryButton(*langs.Pack(guildData.Lang).GetGlobal("yes"), "starboard:yes"),
						discord.NewPrimaryButton(*langs.Pack(guildData.Lang).GetGlobal("no"), "starboard:no"),
						discord.NewSecondaryButton(*cmdPack.Get("skip"), "starboard:skip"),
					),
				}),
			},
		)

		botsReact, err := cmd_util.GetBool(newCtx, msg, cmdPack)
		if err != nil {
			return
		} 

		ctx.Client().Rest().UpdateMessage(
			msg.ChannelID,
			msg.ID,
			discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name:    ctx.User().Username,
							IconURL: *ctx.User().AvatarURL(),
						}),
						Title: *cmdPack.Get("starboardCreating"),
						Color: constants.Colors.Main,
						Description: *cmdPack.Get("selectBotsMsg") +
							*cmdPack.Get("optionalParam") +
							*cmdPack.Get("useButtons"),
						Timestamp: json.Ptr(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardData") +
									*cmdPack.Getf("starboardDataChannel", channel.ID) +
									*cmdPack.Getf("starboardDataEmoji", emoji) +
									*cmdPack.Getf("starboardDataName", name),
							},
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardRequisites") +
									*cmdPack.Getf("starboardRequisitesRequired", required) +
									*cmdPack.Getf("starboardRequisitesRequiredToS", requiredToStay) +
									*cmdPack.Getf("starboardRequisitesBotsReact", utils.ReadableBool(&botsReact, "Si.", "No.")),
							},
						},
					},
				}),
				Components: json.Ptr([]discord.ContainerComponent{
					discord.NewActionRow(
						discord.NewPrimaryButton(*langs.Pack(guildData.Lang).GetGlobal("yes"), "starboard:yes"),
						discord.NewPrimaryButton(*langs.Pack(guildData.Lang).GetGlobal("no"), "starboard:no"),
						discord.NewSecondaryButton(*cmdPack.Get("skip"), "starboard:skip"),
					),
				}),
			},
		)

		botsMsg, err := cmd_util.GetBool(newCtx, msg, cmdPack)
		if err != nil {
			return
		} 

		ctx.Client().Rest().UpdateMessage(
			msg.ChannelID,
			msg.ID,
			discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name:    ctx.User().Username,
							IconURL: *ctx.User().AvatarURL(),
						}),
						Title: *cmdPack.Get("starboardCreating"),
						Color: constants.Colors.Main,
						Description: *cmdPack.Get("selectEmbedColor") +
						*cmdPack.Get("optionalParam"),
						Timestamp: json.Ptr(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardData") +
									*cmdPack.Getf("starboardDataChannel", channel.ID) +
									*cmdPack.Getf("starboardDataEmoji", emoji) +
									*cmdPack.Getf("starboardDataName", name),
							},
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardRequisites") +
									*cmdPack.Getf("starboardRequisitesRequired", required) +
									*cmdPack.Getf("starboardRequisitesRequiredToS", requiredToStay) +
									*cmdPack.Getf("starboardRequisitesBotsReact", utils.ReadableBool(&botsReact, "Si.", "No.")) +
									*cmdPack.Getf("starboardRequisitesBotsMsg", utils.ReadableBool(&botsMsg, "Si.", "No.")),
							},
						},
					},
				}),
				Components: json.Ptr([]discord.ContainerComponent{}),
			},
		)
	
		embedColor, err := cmd_util.GetColor(newCtx, msg, cmdPack, guildData.Lang)
		if err != nil {
			return
		} 

		ctx.Client().Rest().UpdateMessage(
			msg.ChannelID,
			msg.ID,
			discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name:    ctx.User().Username,
							IconURL: *ctx.User().AvatarURL(),
						}),
						Title: *cmdPack.Get("starboardCreating"),
						Color: constants.Colors.Main,
						Description: *cmdPack.Get("starboardConfirm"),
						Timestamp: json.Ptr(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardData") +
									*cmdPack.Getf("starboardDataChannel", channel.ID) +
									*cmdPack.Getf("starboardDataEmoji", emoji) +
									*cmdPack.Getf("starboardDataName", name),
							},
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardRequisites") +
									*cmdPack.Getf("starboardRequisitesRequired", required) +
									*cmdPack.Getf("starboardRequisitesRequiredToS", requiredToStay) +
									*cmdPack.Getf("starboardRequisitesBotsReact", utils.ReadableBool(&botsReact, "Si.", "No.")) +
									*cmdPack.Getf("starboardRequisitesBotsMsg", utils.ReadableBool(&botsMsg, "Si.", "No.")),
							},
							{
								Name:   "\u0020",
								Value:  "\u0020",
								Inline: json.Ptr(true),
							},
							{
								Name:   "\u0020",
								Value:  *langs.Pack(guildData.Lang).Command("starboard").SubCommand("manual").Getf("starboardMessages", embedColor),
								Inline: json.Ptr(true),
							},
						},
					},
				}),
				Components: json.Ptr([]discord.ContainerComponent{
					discord.NewActionRow(
						discord.NewPrimaryButton(*langs.Pack(guildData.Lang).GetGlobal("yes"), "starboard:yes"),
						discord.NewPrimaryButton(*langs.Pack(guildData.Lang).GetGlobal("no"), "starboard:no"),
					),
				}),
			},
		)

		confirm, err := cmd_util.GetBool(newCtx, msg, cmdPack)
		if err != nil {
			return
		}

		if !confirm {
			ctx.Client().Rest().UpdateMessage(
				msg.ChannelID,
				msg.ID,
				discord.MessageUpdate{
					Content:    cmdPack.Get("terminated"),
					Embeds:     json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				},
			)
		}

		starboard := models.Starboard{
			GuildId:         ctx.GuildID().String(),
			Name:            name,
			ChannelId:       channel.ID.String(),
			Emoji:           emoji,
			Required:        required,
			RequiredToStay:  requiredToStay,
			BotsReact:       botsReact,
			BotsMessages:    botsMsg,
			ChannelListType: false,
		}

		if embedColor != 0 {
			starboard.EmbedColor = embedColor
		}

		em, err := utils.ParseEmoji(ctx.Client(), *ctx.GuildID(), starboard.Emoji)
		if err != nil {
			ctx.UpdateInteractionResponse(discord.MessageUpdate{
				Content: cmdPack.Getf("errFindEmoji", err.Error()),
			})
	
			return
		}

		pcolor := ""
		if starboard.EmbedColor != 0 {
			pcolor = strconv.FormatInt(int64(starboard.EmbedColor), 16)
		} else {
			pcolor = strconv.FormatInt(int64(constants.Colors.Main), 16)
		}
	
		err = models.StarboardColl().Create(&starboard)
		if err != nil {
			ctx.UpdateInteractionResponse(discord.MessageUpdate{
				Content: cmdPack.Getf("errCreateStarboard", err.Error()),
			})
	
			return
		}

		ctx.Client().Rest().UpdateMessage(
			msg.ChannelID,
			msg.ID,
			discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name:    ctx.User().Username,
							IconURL: *ctx.User().AvatarURL(),
						}),
						Title: *cmdPack.Get("starboardCreated"),
						Color: constants.Colors.Main,
						Description: *cmdPack.Get("starboardDesc"),
						Timestamp: json.Ptr(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardData") +
									*cmdPack.Getf("starboardDataChannel", channel.ID) +
									*cmdPack.Getf("starboardDataEmoji", em) +
									*cmdPack.Getf("starboardDataName", name),
							},
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardRequisites") +
									*cmdPack.Getf("starboardRequisitesRequired", required) +
									*cmdPack.Getf("starboardRequisitesRequiredToS", requiredToStay) +
									*cmdPack.Getf("starboardRequisitesBotsReact", utils.ReadableBool(&botsReact, "Si.", "No.")) +
									*cmdPack.Getf("starboardRequisitesBotsMsg", utils.ReadableBool(&botsMsg, "Si.", "No.")),
							},
							{
								Name:   "\u0020",
								Value:  "\u0020",
								Inline: json.Ptr(true),
							},
							{
								Name:   "\u0020",
								Value:  *langs.Pack(guildData.Lang).Command("starboard").SubCommand("manual").Getf("starboardMessages", pcolor),
								Inline: json.Ptr(true),
							},
						},
					},
				}),
				Components: json.Ptr([]discord.ContainerComponent{
				}),
			},
		)

		}()

	return nil
}