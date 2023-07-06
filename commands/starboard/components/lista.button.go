package components

import (
	"strconv"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/rs/zerolog/log"

	"github.com/vyrekxd/kirby/commands/starboard"
	"github.com/vyrekxd/kirby/constants"
	"github.com/vyrekxd/kirby/database/models"
	"github.com/vyrekxd/kirby/langs"
)

func NextButton(e *handler.ComponentEvent) error {
	if e.Message.Author.ID != e.Client().ID() {
		return nil
	}

	err := e.DeferUpdateMessage()
	if err != nil {
		e.Client().Rest().DeleteMessage(
			e.Message.ChannelID,
			e.Message.ID,
		)

		log.Error().
			Err(err).
			Msg("Error ocurred when trying to defer message in \"starboard:lista\"")

		return nil
	}

	guildData := models.GuildConfig{
		Lang: "es-MX",
	}
	err = models.GuildConfigColl().
		FindByID(e.GuildID().String(), &guildData)
	if err != nil {
		e.Client().Rest().DeleteMessage(
			e.Message.ChannelID,
			e.Message.ID,
		)

		return nil
	}

	cmdPack := langs.Pack(guildData.Lang).
		Command("starboard").
		SubCommand("interactivo")

	starboards := []models.Starboard{}
	err = models.StarboardColl().SimpleFind(&starboards, bson.M{"guild_id": e.GuildID().String()})
	if err != nil {
		e.Client().Rest().DeleteMessage(
			e.Message.ChannelID,
			e.Message.ID,
		)

		return nil
	}

	indexStr := e.Variables["index"]
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		e.Client().Rest().DeleteMessage(
			e.Message.ChannelID,
			e.Message.ID,
		)

		return nil
	}
	selected := starboards[(9 * index):]
	if len(selected) > 10 {
		selected = selected[:10]
	}

	text := []string{}

	for i, starboard := range selected {
		text = append(text, *cmdPack.Getf("embedStarboardData", (i+1)+(9*index), starboard.Emoji, starboard.Name, starboard.ChannelId, starboard.ID.Hex()))
	}

	message := discord.MessageUpdate{
		Embeds: &[]discord.Embed{
			{
				Title: *cmdPack.Get("embedTitle"),
				Author: &discord.EmbedAuthor{
					Name:    e.User().Username,
					IconURL: *e.User().AvatarURL(),
				},
				Color:       constants.Colors.Main,
				Description: strings.Join(text, "\n"),
				Footer: &discord.EmbedFooter{
					Text: *cmdPack.Getf("embedFooter", len(selected)+(10*index), len(starboards)),
				},
				Timestamp: json.Ptr(time.Now()),
			},
		},
	}

	if float64(index) == float64(len(starboards)/10) {
		message.Components = &[]discord.ContainerComponent{
			discord.ActionRowComponent{
				discord.ButtonComponent{
					Style: discord.ButtonStyleSecondary,
					Emoji: &discord.ComponentEmoji{
						Name: "⬅️",
					},
					CustomID: starboard.ReturnId + "/" + strconv.Itoa(index-1),
				},
				discord.ButtonComponent{
					Style: discord.ButtonStyleSecondary,
					Emoji: &discord.ComponentEmoji{
						Name: "➡️",
					},
					CustomID: starboard.NextId + "/" + strconv.Itoa(index),
					Disabled: true,
				},
			},
		}
	} else {
		message.Components = &[]discord.ContainerComponent{
			discord.ActionRowComponent{
				discord.ButtonComponent{
					Style: discord.ButtonStyleSecondary,
					Emoji: &discord.ComponentEmoji{
						Name: "⬅️",
					},
					CustomID: starboard.ReturnId + "/" + strconv.Itoa(index-1),
				},
				discord.ButtonComponent{
					Style: discord.ButtonStyleSecondary,
					Emoji: &discord.ComponentEmoji{
						Name: "➡️",
					},
					CustomID: starboard.NextId + "/" + strconv.Itoa(index+1),
				},
			},
		}
	}

	_, err = e.UpdateInteractionResponse(message)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error ocurred when trying to defer message in \"starboard:lista\"")

		return nil
	}

	return nil
}

func ReturnButton(e *handler.ComponentEvent) error {
	if e.Message.Author.ID != e.Client().ID() {
		return nil
	}

	err := e.DeferUpdateMessage()
	if err != nil {
		e.Client().Rest().DeleteMessage(
			e.Message.ChannelID,
			e.Message.ID,
		)

		log.Error().
			Err(err).
			Msg("Error ocurred when trying to defer message in \"starboard:lista\"")

		return nil
	}

	guildData := models.GuildConfig{
		Lang: "es-MX",
	}
	err = models.GuildConfigColl().
		FindByID(e.GuildID().String(), &guildData)
	if err != nil {
		e.Client().Rest().DeleteMessage(
			e.Message.ChannelID,
			e.Message.ID,
		)

		return nil
	}

	cmdPack := langs.Pack(guildData.Lang).
		Command("starboard").
		SubCommand("interactivo")

	starboards := []models.Starboard{}
	err = models.StarboardColl().SimpleFind(&starboards, bson.M{"guild_id": e.GuildID().String()})
	if err != nil {
		e.Client().Rest().DeleteMessage(
			e.Message.ChannelID,
			e.Message.ID,
		)

		return nil
	}

	indexStr := e.Variables["index"]
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		e.Client().Rest().DeleteMessage(
			e.Message.ChannelID,
			e.Message.ID,
		)

		return nil
	}
	selected := starboards[(9 * index):]
	if len(selected) > 10 {
		selected = selected[:10]
	}

	text := []string{}

	for i, starboard := range selected {
		text = append(text, *cmdPack.Getf("embedStarboardData", (i+1)+(9*index), starboard.Emoji, starboard.Name, starboard.ChannelId, starboard.ID.Hex()))
	}

	message := discord.MessageUpdate{
		Embeds: &[]discord.Embed{
			{
				Title: *cmdPack.Get("embedTitle"),
				Author: &discord.EmbedAuthor{
					Name:    e.User().Username,
					IconURL: *e.User().AvatarURL(),
				},
				Color:       constants.Colors.Main,
				Description: strings.Join(text, "\n"),
				Footer: &discord.EmbedFooter{
					Text: *cmdPack.Getf("embedFooter", len(selected)+(10*index), len(starboards)),
				},
				Timestamp: json.Ptr(time.Now()),
			},
		},
	}

	if index == 0 {
		message.Components = &[]discord.ContainerComponent{
			discord.ActionRowComponent{
				discord.ButtonComponent{
					Style: discord.ButtonStyleSecondary,
					Emoji: &discord.ComponentEmoji{
						Name: "⬅️",
					},
					CustomID: starboard.ReturnId + "/" + "0",
					Disabled: true,
				},
				discord.ButtonComponent{
					Style: discord.ButtonStyleSecondary,
					Emoji: &discord.ComponentEmoji{
						Name: "➡️",
					},
					CustomID: starboard.NextId + "/" + strconv.Itoa(index+1),
				},
			},
		}
	} else {
		message.Components = &[]discord.ContainerComponent{
			discord.ActionRowComponent{
				discord.ButtonComponent{
					Style: discord.ButtonStyleSecondary,
					Emoji: &discord.ComponentEmoji{
						Name: "⬅️",
					},
					CustomID: starboard.ReturnId + "/" + strconv.Itoa(index-1),
					Disabled: true,
				},
				discord.ButtonComponent{
					Style: discord.ButtonStyleSecondary,
					Emoji: &discord.ComponentEmoji{
						Name: "➡️",
					},
					CustomID: starboard.NextId + "/" + strconv.Itoa(index+1),
				},
			},
		}
	}

	_, err = e.UpdateInteractionResponse(message)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error ocurred when trying to defer message in \"starboard:lista\"")

		return nil
	}

	return nil
}
