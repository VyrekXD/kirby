package starboard

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"
	"github.com/disgoorg/snowflake/v2"

	"github.com/forPelevin/gomoji"

	"github.com/vyrekxd/kirby/constants"
	"github.com/vyrekxd/kirby/database/models"
	"github.com/vyrekxd/kirby/langs"
	"github.com/vyrekxd/kirby/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func starboardManual(ctx *handler.CommandEvent) error {
	guildData := models.GuildConfig{}
	starboards := []models.Starboard{}
	err := models.GuildConfigColl().FindByID(ctx.GuildID().String(), &guildData)
	if err == mongo.ErrNoDocuments {
		guildData = models.GuildConfig{DefaultModel: models.DefaultModel{ID: ctx.GuildID().String()}, Lang: "es-MX"}
		err := models.GuildConfigColl().Create(&guildData)
		if err != nil {
			ctx.UpdateInteractionResponse(discord.MessageUpdate{
				Content: langs.Pack(guildData.Lang).Command("starboard").SubCommand("manual").Getf("errCreateGuild", err),
			})

			return nil
		}
	} else {
		err := models.StarboardColl().SimpleFind(&starboards, bson.M{"guild_id": ctx.GuildID().String()})
		if err != nil && err != mongo.ErrNoDocuments {
			ctx.UpdateInteractionResponse(discord.MessageUpdate{
				Content: langs.Pack(guildData.Lang).Command("starboard").SubCommand("manual").Getf("errFindGuildStarboards", err),
			})

			return nil
		}
	}

	cmdPack := langs.Pack(guildData.Lang).Command("starboard").SubCommand("manual")

	data := ctx.SlashCommandInteractionData()
	channel, ok := data.OptChannel("canal")
	if !ok {
		ctx.UpdateInteractionResponse(discord.MessageUpdate{
			Content: cmdPack.Get("noChannel"),
		})

		return nil
	}

	for _, s := range starboards {
		if s.ChannelId == channel.ID.String() {
			ctx.UpdateInteractionResponse(discord.MessageUpdate{
				Content: cmdPack.Get("channelAlreadyUsed"),
			})

			return nil
		} else {
			for _, cid := range s.ChannelList {
				if cid == channel.ID.String() {
					ctx.UpdateInteractionResponse(discord.MessageUpdate{
						Content: cmdPack.Get("channelInList"),
					})

					return nil
				}
			}
		}
	}

	emoji, ok := data.OptString("emoji")
	if !ok {
		ctx.UpdateInteractionResponse(discord.MessageUpdate{
			Content: cmdPack.Get("noEmoji"),
		})

		return nil
	} else if constants.DiscordEmojiRegex.FindString(fmt.Sprint(emoji)) != "" {
		res := constants.DiscordEmojiRegex.FindAllString(fmt.Sprint(emoji), 2)
		if len(res) > 1 {
			ctx.UpdateInteractionResponse(discord.MessageUpdate{
				Content: cmdPack.Get("noValidEmoji"),
			})

			return nil
		}

		emoji = fmt.Sprint(emoji)
		emoji = constants.CleanIdRegex.ReplaceAllString(constants.DiscordEmojiIdRegex.FindString(emoji), "")
	} else if res := gomoji.FindAll(emoji); res == nil && len(res) > 1 {
		ctx.UpdateInteractionResponse(discord.MessageUpdate{
			Content: cmdPack.Get("noValidEmoji"),
		})

		return nil
	}

	if r := models.StarboardColl().FindOne(context.TODO(), bson.M{"emoji": emoji, "guild_id": ctx.GuildID().String()}); r.Err() != nil && r.Err() != mongo.ErrNoDocuments {
		ctx.UpdateInteractionResponse(discord.MessageUpdate{
			Content: cmdPack.Get("noUniqueEmoji"),
		})

		return nil
	}

	required, ok := data.OptInt("requeridos")
	if !ok {
		ctx.UpdateInteractionResponse(discord.MessageUpdate{
			Content: cmdPack.Get("noRequired"),
		})

		return nil
	} else if required <= 0 {
		ctx.UpdateInteractionResponse(discord.MessageUpdate{
			Content: cmdPack.Get("noValidRequired"),
		})

		return nil
	}

	requiredToStay, ok := data.OptInt("requeridos-para-quedarse")
	if !ok {
		ctx.UpdateInteractionResponse(discord.MessageUpdate{
			Content: cmdPack.Get("noRequiredToStay"),
		})

		return nil
	} else if requiredToStay < 0 || requiredToStay >= required {
		ctx.UpdateInteractionResponse(discord.MessageUpdate{
			Content: cmdPack.Get("noValidRequiredToStay"),
		})

		return nil
	}

	name, ok := data.OptString("nombre")
	if !ok {
		name = channel.Name
	} else if len(name) > 100 {
		ctx.UpdateInteractionResponse(discord.MessageUpdate{
			Content: cmdPack.Get("noValidName"),
		})

		return nil
	}

	botsCanReact, ok := data.OptBool("bots-reacciones")
	if !ok {
		botsCanReact = false
	}

	botsMessages, ok := data.OptBool("bots-mensajes")
	if !ok {
		botsMessages = false
	}

	parsedEmbedColor := 0
	embedColor, ok := data.OptString("embed-color")
	if ok && constants.HexColorRegex.FindString(embedColor) == "" {
		ctx.UpdateInteractionResponse(discord.MessageUpdate{
			Content: cmdPack.Get("noValidHex"),
		})

		return nil
	} else if ok && len(embedColor) == 4 {
		fembedColor := ""

		for _, d := range strings.Split(strings.ToLower(strings.Replace(embedColor, "#", "", 1)), "") {
			fembedColor += strings.Repeat(d, 2)
		}

		parsedInt, err := strconv.ParseInt(fembedColor, 16, 64)
		if err != nil {
			ctx.UpdateInteractionResponse(discord.MessageUpdate{
				Content: cmdPack.Get("noValidParsedHex"),
			})

			return nil
		}

		parsedEmbedColor = int(parsedInt)
	} else if ok {
		embedColor = strings.ToLower(strings.Replace(embedColor, "#", "", 1))

		parsedInt, err := strconv.ParseInt(embedColor, 16, 64)
		if err != nil {
			ctx.UpdateInteractionResponse(discord.MessageUpdate{
				Content: cmdPack.Get("noValidParsedHex"),
			})

			return nil
		}

		parsedEmbedColor = int(parsedInt)
	}

	listType, ok := data.OptBool("lista-tipo")
	if !ok {
		listType = false
	}

	var listChannels []*discord.GuildChannel
	listChannelsString, ok := data.OptString("canales-lista")
	if !ok && listType {
		ctx.UpdateInteractionResponse(discord.MessageUpdate{
			Content: cmdPack.Get("noChannelsOnWhite"),
		})

		return nil
	} else if ok {
		listChannelsSplit := strings.Split(
			strings.Join(
				strings.Split(
					listChannelsString,
					"\x20",
				), ",",
			), ",",
		)

		for _, id := range listChannelsSplit {
			ch, ok := ctx.Client().Caches().Channel(snowflake.MustParse(id))
			if ok {
				ctx.UpdateInteractionResponse(discord.MessageUpdate{
					Content: json.Ptr(fmt.Sprintf(*cmdPack.Get("noValidChannelId"), id)),
				})

				return nil
			}

			listChannels = append(listChannels, &ch)
		}
	}

	starboard := models.Starboard{
		GuildId:         ctx.GuildID().String(),
		Name:            name,
		ChannelId:       channel.ID.String(),
		Emoji:           emoji,
		Required:        required,
		RequiredToStay:  requiredToStay,
		BotsReact:       botsCanReact,
		BotsMessages:    botsMessages,
		ChannelListType: listType,
	}

	if embedColor != "" {
		starboard.EmbedColor = parsedEmbedColor
	}

	if len(listChannels) != 0 {
		starboard.ChannelList = []string{}

		for _, c := range listChannels {
			starboard.ChannelList = append(starboard.ChannelList, (*c).ID().String())
		}
	}

	err = models.StarboardColl().Create(&starboard)
	if err != nil {
		ctx.UpdateInteractionResponse(discord.MessageUpdate{
			Content: cmdPack.Getf("errCreateStarboard", err.Error()),
		})

		return nil
	}

	em, err := utils.ParseEmoji(ctx.Client(), *ctx.GuildID(), starboard.Emoji)
	if err != nil {
		ctx.UpdateInteractionResponse(discord.MessageUpdate{
			Content: cmdPack.Getf("errFindEmoji", err.Error()),
		})
	}

	pcolor := ""
	if parsedEmbedColor != 0 {
		pcolor = strconv.FormatInt(int64(parsedEmbedColor), 16)
	} else {
		pcolor = strconv.FormatInt(int64(constants.Colors.Main), 16)
	}

	ctx.UpdateInteractionResponse(discord.MessageUpdate{
		Embeds: json.Ptr([]discord.Embed{
			{
				Author: json.Ptr(discord.EmbedAuthor{
					Name:    ctx.User().Username,
					IconURL: *ctx.User().AvatarURL(),
				}),
				Title: *cmdPack.Get("starboardCreated"),
				Color: parsedEmbedColor | constants.Colors.Main,
				Fields: []discord.EmbedField{
					{
						Name:   "\u0020",
						Value:  *cmdPack.Getf("starboardData", starboard.Name, channel.ID, em),
						Inline: json.Ptr(true),
					},
					{
						Name:   "\u0020",
						Value:  *cmdPack.Getf("starboardRequisites", starboard.Required, starboard.RequiredToStay, utils.ReadableBool(&starboard.BotsReact, "Si.", "No."), utils.ReadableBool(&starboard.BotsMessages, "Si.", "No.")),
						Inline: json.Ptr(true),
					},
					{
						Name:   "\u0020",
						Value:  "\u0020",
						Inline: json.Ptr(true),
					},
					{
						Name:   "\u0020",
						Value:  *cmdPack.Getf("starboardMessages", pcolor),
						Inline: json.Ptr(true),
					},
				},
			},
		}),
	})

	return nil
}