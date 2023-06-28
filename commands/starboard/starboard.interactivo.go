package starboard

import (
	"context"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"
	"github.com/rs/zerolog/log"

	"github.com/vyrekxd/kirby/constants"
	"github.com/vyrekxd/kirby/database/models"
	"github.com/vyrekxd/kirby/langs"
	"github.com/vyrekxd/kirby/utils"

	"go.mongodb.org/mongo-driver/mongo"
)

const (
	BaseComponentId = "starboard:"
	SelectChannelId = BaseComponentId + "channel"
	ModalId         = BaseComponentId + "modal"
	NameInputId     = ModalId + "nameInput"
	RequiredInputId = ModalId + "requiredInput"
	YesButtonId     = BaseComponentId + "yes"
	NoButtonId      = BaseComponentId + "no"
	SkipButtonId    = BaseComponentId + "skip"
	OmitButtonId    = BaseComponentId + "omit"
)

func DeleteTempStarboard(tempStarboard *models.TempStarboard) {
	err := models.TempStarboardColl().Delete(tempStarboard)
	if err != nil {
		log.Error().Err(err).Msg("Error deleting temp starboard")
	}
}

func StarboardInteractivo(ctx *handler.CommandEvent) error {
	err := ctx.DeferCreateMessage(false)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error ocurred when trying to defer message in \"starboard:interactivo\"")

		return nil
	}

	guildData := models.GuildConfig{Lang: "es-MX"}
	err = models.GuildConfigColl().
		FindByID(ctx.GuildID().String(), &guildData)
	if err == mongo.ErrNoDocuments {
		ctx.Client().
			Rest().
			UpdateInteractionResponse(ctx.ApplicationID(), ctx.Token(), discord.MessageUpdate{
				Content: langs.Pack(guildData.Lang).
					Command("starboard").
					SubCommand("manual").
					Getf("errNoGuildData", err),
			})

		return nil
	} else if err != nil && err != mongo.ErrNoDocuments {
		ctx.Client().Rest().UpdateInteractionResponse(ctx.ApplicationID(), ctx.Token(), discord.MessageUpdate{
			Content: langs.Pack(guildData.Lang).
				Command("starboard").
				SubCommand("manual").
				Getf("errUnexpected", err),
		})

		return nil
	}

	findTempStar := &models.TempStarboard{}
	err = models.TempStarboardColl().
		First(models.TempStarboard{GuildId: ctx.GuildID().String()}, findTempStar)
	if err != nil && err != mongo.ErrNoDocuments {
		ctx.Client().
			Rest().
			UpdateInteractionResponse(ctx.ApplicationID(), ctx.Token(), discord.MessageUpdate{
				Content: langs.Pack(guildData.Lang).
					Command("starboard").
					SubCommand("manual").
					Getf("errUnexpected", err),
			})

		return nil
	} else if err != nil && err == mongo.ErrNoDocuments {
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
		})
		if err != nil {
			log.Error().
				Err(err).
				Msg("Error ocurred when trying to respond in \"starboard:interactivo\"")

			return nil
		}

		tempStarboard := &models.TempStarboard{
			GuildId:      ctx.GuildID().String(),
			MsgChannelId: msg.ChannelID.String(),
			UserId:       ctx.User().ID.String(),
			MessageId:    msg.ID.String(),
		}

		_, err = models.TempStarboardColl().InsertOne(context.Background(), tempStarboard)
		if err != nil {
			_, err = ctx.Client().Rest().UpdateMessage(
				msg.ChannelID,
				msg.ID, discord.MessageUpdate{
					Content: langs.Pack(guildData.Lang).
						Command("starboard").
						Getf("errUnexpected", err),
				})

			return nil
		}

		ctx.UpdateInteractionResponse(discord.MessageUpdate{
			Components: json.Ptr([]discord.ContainerComponent{
				discord.NewActionRow(
					discord.ChannelSelectMenuComponent{
						CustomID:     SelectChannelId + "/" + tempStarboard.ID.String(),
						MaxValues:    1,
						ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildText},
					},
				),
			}),
		})

		utils.WaitDo(time.Second*50, func() {
			data := &models.TempStarboard{}
			models.TempStarboardColl().First(tempStarboard, data)

			if data.ChannelId == "" {
				DeleteTempStarboard(tempStarboard)

				_, errM := ctx.Client().Rest().UpdateMessage(
					msg.ChannelID,
					msg.ID,
					discord.MessageUpdate{
						Content:    cmdPack.Get("errTimeout"),
						Embeds:     json.Ptr([]discord.Embed{}),
						Components: json.Ptr([]discord.ContainerComponent{}),
					},
				)
				if errM != nil {
					log.Error().Err(err).Msg("Error ocurred when trying to edit message in \"starboard:interactivo\"")
				}
			}
		})

		return nil
	} else {
		ctx.Client().Rest().UpdateInteractionResponse(ctx.ApplicationID(), ctx.Token(), discord.MessageUpdate{
			Content: langs.Pack(guildData.Lang).
				Command("starboard").
				Getf("alreadyOnSetup", err),
		})

		return nil
	}
}
