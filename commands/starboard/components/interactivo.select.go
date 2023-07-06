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

	starboards := []models.Starboard{}
	err = models.StarboardColl().
		SimpleFind(&starboards, models.Starboard{GuildId: e.GuildID().String()})
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

	menuData := e.ChannelSelectMenuInteractionData()
	if len(menuData.Channels()) < 1 {
		starboard.DeleteTempStarboard(data)
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
			Msg("Error ocurred when trying to respond in \"starboard:interactivo:channel\"")

		return nil
	}

	err = e.CreateModal(discord.ModalCreate{
		CustomID: starboard.ModalId + "/" + data.ID.Hex(),
		Title:    *cmdPack.Get("starboardCreating"),
		Components: []discord.ContainerComponent{
			discord.ActionRowComponent{
				discord.TextInputComponent{
					CustomID:    starboard.NameInputId,
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
					CustomID:    starboard.RequiredInputId,
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
		starboard.DeleteTempStarboard(data)
		e.UpdateMessage(discord.MessageUpdate{
			Content:    cmdPack.Getf("errUnexpected", err),
			Embeds:     json.Ptr([]discord.Embed{}),
			Components: json.Ptr([]discord.ContainerComponent{}),
		})

		return nil
	}

	utils.WaitDo(time.Second*50, func() {
		find := &models.TempStarboard{}
		err := models.TempStarboardColl().First(data, find)

		if err == nil && find.Required == 0 {
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
