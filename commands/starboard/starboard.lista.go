package starboard

import (
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"

	"github.com/rs/zerolog/log"

	"github.com/vyrekxd/kirby/constants"
	"github.com/vyrekxd/kirby/database/models"
	"github.com/vyrekxd/kirby/langs"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	GetIndex = "/{index}"
	ReturnId = BaseComponentId + "return"
	NextId   = BaseComponentId + "next"
)

func StarboardLista(ctx *handler.CommandEvent) error {
	err := ctx.DeferCreateMessage(false)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error ocurred when trying to defer message in \"starboard:lista\"")

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
					Getf("errNoGuildData", err),
			})

		return nil
	} else if err != nil && err != mongo.ErrNoDocuments {
		ctx.Client().Rest().UpdateInteractionResponse(ctx.ApplicationID(), ctx.Token(), discord.MessageUpdate{
			Content: langs.Pack(guildData.Lang).
				Command("starboard").
				Getf("errUnexpected", err),
		})

		return nil
	}

	cmdPack := langs.Pack(guildData.Lang).
		Command("starboard").
		SubCommand("lista")
	starboards := []models.Starboard{}
	err = models.StarboardColl().SimpleFind(&starboards, bson.M{"guild_id": ctx.GuildID().String()})
	if err != nil {
		ctx.Client().Rest().UpdateInteractionResponse(ctx.ApplicationID(), ctx.Token(), discord.MessageUpdate{
			Content: langs.Pack(guildData.Lang).
				Command("starboard").
				Getf("errUnexpected", err),
		})

		return nil
	} else if len(starboards) == 0 {
		ctx.Client().Rest().UpdateInteractionResponse(ctx.ApplicationID(), ctx.Token(), discord.MessageUpdate{
			Content: cmdPack.Get("errNoStarboards"),
		})

		return nil
	}

	first := starboards
	if len(starboards) > 10 {
		first = starboards[:10]
	}

	text := []string{}

	for i, starboard := range first {
		text = append(text, *cmdPack.Getf("embedStarboardData", i+1, starboard.Emoji, starboard.Name, starboard.ChannelId, starboard.ID.Hex()))
	}

	message := discord.MessageUpdate{
		Embeds: &[]discord.Embed{
			{
				Title: *cmdPack.Get("embedTitle"),
				Author: &discord.EmbedAuthor{
					Name:    ctx.User().Username,
					IconURL: *ctx.User().AvatarURL(),
				},
				Color:       constants.Colors.Main,
				Description: strings.Join(text, "\n"),
				Footer: &discord.EmbedFooter{
					Text: *cmdPack.Getf("embedFooter", len(first), len(starboards)),
				},
				Timestamp: json.Ptr(time.Now()),
			},
		},
	}

	if len(starboards) > 10 {
		message.Components = &[]discord.ContainerComponent{
			discord.ActionRowComponent{
				discord.ButtonComponent{
					Style: discord.ButtonStyleSecondary,
					Emoji: &discord.ComponentEmoji{
						Name: "⬅️",
					},
					CustomID: ReturnId + "/" + "0",
					Disabled: true,
				},
				discord.ButtonComponent{
					Style: discord.ButtonStyleSecondary,
					Emoji: &discord.ComponentEmoji{
						Name: "➡️",
					},
					CustomID: NextId + "/" + "1",
				},
			},
		}
	}

	_, err = ctx.UpdateInteractionResponse(message)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error ocurred when trying to defer message in \"starboard:lista\"")

		return nil
	}

	return nil
}
