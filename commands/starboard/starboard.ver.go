package starboard

import (
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"
	"github.com/rs/zerolog/log"
	"github.com/vyrekxd/kirby/constants"
	"github.com/vyrekxd/kirby/database/models"
	"github.com/vyrekxd/kirby/langs"
	"github.com/vyrekxd/kirby/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func StarboardVer(ctx *handler.CommandEvent) error {
	err := ctx.DeferCreateMessage(true)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error ocurred when trying to defer message in \"starboard:ver\"")

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
		SubCommand("ver")
	data := ctx.SlashCommandInteractionData()
	channel, okC := data.OptChannel("canal")
	idS, okS := data.OptString("id")
	if !okC && !okS {
		ctx.Client().Rest().UpdateInteractionResponse(ctx.ApplicationID(), ctx.Token(), discord.MessageUpdate{
			Content: cmdPack.Get("errNoArgs"),
		})

		return nil
	}

	starboard := &models.Starboard{}
	if okC {
		err = models.StarboardColl().First(bson.M{"guild_id": ctx.GuildID().String(), "channel_id": channel.ID.String()}, starboard)
		if err != nil {
			ctx.Client().Rest().UpdateInteractionResponse(ctx.ApplicationID(), ctx.Token(), discord.MessageUpdate{
				Content: langs.Pack(guildData.Lang).
					Command("starboard").
					Getf("errUnexpected", err),
			})

			return nil
		}
	} else {
		id, err := primitive.ObjectIDFromHex(idS)
		if err != nil {
			ctx.Client().Rest().UpdateInteractionResponse(ctx.ApplicationID(), ctx.Token(), discord.MessageUpdate{
				Content: cmdPack.Get("errNoValidID"),
			})

			return nil
		}

		err = models.StarboardColl().FindByID(id, starboard)
		if err != nil {
			ctx.Client().Rest().UpdateInteractionResponse(ctx.ApplicationID(), ctx.Token(), discord.MessageUpdate{
				Content: langs.Pack(guildData.Lang).
					Command("starboard").
					Getf("errUnexpected", err),
			})

			return nil
		}
	}

	starLangPack := langs.Pack(guildData.Lang).
		Command("starboard").
		SubCommand("interactivo")
	emoji, err := utils.ParseEmoji(starboard.Emoji, *ctx.GuildID(), ctx.Client())
	if err != nil {
		models.StarboardColl().Delete(starboard)
		ctx.Client().Rest().UpdateInteractionResponse(ctx.ApplicationID(), ctx.Token(), discord.MessageUpdate{
			Content: cmdPack.Get("errNoValidEmoji"),
		})

		return nil
	}

	levels := &starboard.Levels
	if levels.FirstEmoji == "" {
		levels.FirstEmoji = constants.Emojis.Star
	} else {
		firstEmoji, err := utils.ParseEmoji(levels.FirstEmoji, *ctx.GuildID(), ctx.Client())
		if err != nil {
			levels.FirstEmoji = ""
			models.StarboardColl().Update(starboard)

			levels.FirstEmoji = constants.Emojis.Star
		}

		levels.FirstEmoji = firstEmoji
	}
	if levels.SecondEmoji == "" {
		levels.SecondEmoji = constants.Emojis.SecondStar
	} else {
		secondEmoji, err := utils.ParseEmoji(levels.SecondEmoji, *ctx.GuildID(), ctx.Client())
		if err != nil {
			levels.SecondEmoji = ""
			models.StarboardColl().Update(starboard)

			levels.SecondEmoji = constants.Emojis.SecondStar
		}

		levels.SecondEmoji = secondEmoji
	}
	if levels.ThirdEmoji == "" {
		levels.ThirdEmoji = constants.Emojis.ThirdStar
	} else {
		thirdEmoji, err := utils.ParseEmoji(levels.ThirdEmoji, *ctx.GuildID(), ctx.Client())
		if err != nil {
			levels.ThirdEmoji = ""
			models.StarboardColl().Update(starboard)

			levels.ThirdEmoji = constants.Emojis.ThirdStar
		}

		levels.ThirdEmoji = thirdEmoji
	}

	message := discord.MessageUpdate{
		Embeds: &[]discord.Embed{
			{
				Author: &discord.EmbedAuthor{
					Name:    ctx.User().Username,
					IconURL: *ctx.User().AvatarURL(),
				},
				Title: starboard.Name,
				Color: starboard.EmbedColor,
				Fields: []discord.EmbedField{
					{
						Name: "\u0020",
						Value: *starLangPack.Get("starboardData") +
							*starLangPack.Getf("starboardDataChannel", starboard.ChannelId) +
							*starLangPack.Getf("starboardDataName", starboard.Name) +
							*starLangPack.Getf("starboardDataEmoji", emoji),
						Inline: json.Ptr(true),
					},
					{
						Name: "\u0020",
						Value: *starLangPack.Get("starboardRequisites") +
							*starLangPack.Getf("starboardRequisitesRequired", starboard.Required) +
							*starLangPack.Getf("starboardRequisitesBotsReact", utils.ReadableBool(
								&starboard.BotsReact,
								*langs.Pack(guildData.Lang).GetGlobal("yes"),
								*langs.Pack(guildData.Lang).GetGlobal("no"),
							)) +
							*starLangPack.Getf("starboardRequisitesBotsMsg", utils.ReadableBool(
								&starboard.BotsMessages,
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
							Getf("starboardCustom", utils.ToString(starboard.EmbedColor), utils.ToString(starboard.EmbedColor)),
						Inline: json.Ptr(true),
					},
					{
						Name: "\u0020",
						Value: *cmdPack.Get("starboardLevels") +
							*cmdPack.Getf("starboardLevelsFirst", levels.FirstEmoji) +
							*cmdPack.Getf("starboardLevelsSecond",
								levels.Second|starboard.Required+constants.DifferenceBetweenLevels,
								levels.SecondEmoji) +
							*cmdPack.Getf("starboardLevelsThird",
								levels.Third|starboard.Required+constants.DifferenceBetweenLevels*2,
								levels.ThirdEmoji),
						Inline: json.Ptr(true),
					},
					{
						Name:   "\u0020",
						Value:  "\u0020",
						Inline: json.Ptr(true),
					},
				},
				Timestamp: json.Ptr(time.Now()),
				Footer: &discord.EmbedFooter{
					Text: *cmdPack.Getf("starboardFooter", starboard.ID.Hex()),
				},
			},
		},
	}

	if len(starboard.ChannelList) != 0 {
		e := *message.Embeds
		e[0].Fields = append(e[0].Fields, discord.EmbedField{
			Name: "\u0020",
			Value: *cmdPack.Get("starboardList") +
				*cmdPack.Getf("starboardListType",
					utils.ReadableBool(&starboard.ChannelListType, "White", "Black")+"list",
				) +
				*cmdPack.Getf("starboardListChannels",
					strings.Join(utils.Map(starboard.ChannelList, func(id string) string {
						return fmt.Sprintf("<#%v>", id)
					}), "\n"),
				),
			Inline: json.Ptr(true),
		})

		message.Embeds = &e
	}

	_, err = ctx.UpdateInteractionResponse(message)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error ocurred when trying to update message in \"starboard:ver\"")

		return nil
	}

	return nil
}
