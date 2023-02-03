package starboard

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"
	"github.com/forPelevin/gomoji"
	"github.com/rs/zerolog/log"
	"github.com/vyrekxd/kirby/constants"
	"github.com/vyrekxd/kirby/database/models"
	"github.com/vyrekxd/kirby/langs"
	"github.com/vyrekxd/kirby/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	errNoChannel = errors.New("noChannel")
	errNoValidEmoji = errors.New("noValidEmoji")
)

func starboardInteractivo(ctx *handler.CommandEvent) error {
	go func(){
		guildData := models.GuildConfig{Lang: "es-MX"}
		starboards := []models.Starboard{}
		err := models.GuildConfigColl().FindByID(ctx.GuildID().String(), &guildData)
		if err == mongo.ErrNoDocuments {
			ctx.UpdateInteractionResponse(discord.MessageUpdate{
				Content: langs.Pack(guildData.Lang).Command("starboard").SubCommand("manual").Getf("errNoGuildData", err),
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

		cmdPack := langs.Pack(guildData.Lang).Command("starboard").SubCommand("interactivo")
	
		msg, err := ctx.UpdateInteractionResponse(discord.MessageUpdate{
			Embeds: json.Ptr([]discord.Embed{
				{
					Author: json.Ptr(discord.EmbedAuthor{
						Name: ctx.User().Username,
						IconURL: *ctx.User().AvatarURL(),
					}),
					Title: *cmdPack.Get("starboardCreating"),
					Color: constants.Colors.Main,
					Description: *cmdPack.Get("selectChannel"),
					Timestamp: json.Ptr(time.Now()),
				},
			}),
			Components: json.Ptr([]discord.ContainerComponent{
				discord.NewActionRow(
					discord.ChannelSelectMenuComponent{
						CustomID: "starboard:channel",
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
			log.Error().Err(err).Msg("Error ocurred when trying to respond in \"starboard:interactivo\"")
			return
		}

		timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 20 * time.Second)
		errCh := make(chan error)
		eventCh := make(chan utils.CollectedEvent[events.ComponentInteractionCreate])
		utils.NewCollector(
			ctx.Client(),
			timeoutCtx,
			func(e events.ComponentInteractionCreate) bool {
				return e.ChannelSelectMenuInteractionData().CustomID() == "starboard:channel" &&
				e.User().ID == ctx.User().ID &&
				e.ChannelID() == ctx.ChannelID()
			},
			cancelTimeout,
			eventCh,
			errCh,
		)
		err, ev := <- errCh, <- eventCh
		switch err {
			case utils.ErrTimeout: {
				ctx.Client().Rest().UpdateMessage(
					msg.ChannelID,
					msg.ID,
					discord.MessageUpdate{
						Content: cmdPack.Get("errTimeout"),
						Embeds: json.Ptr([]discord.Embed{}),
						Components: json.Ptr([]discord.ContainerComponent{}),
					},
				)

				return
			}
			default: {
				if err != nil {
					if !reflect.ValueOf(ev.Data).IsZero() {
						ev.Data.DeferUpdateMessage()
					}

					ctx.Client().Rest().UpdateMessage(
						msg.ChannelID,
						msg.ID,
						discord.MessageUpdate{
							Content: cmdPack.Get("errCollector"),
							Embeds: json.Ptr([]discord.Embed{}),
							Components: json.Ptr([]discord.ContainerComponent{}),
						},
					)

					return
				}
			}
		}

		menu := ev.Data.ChannelSelectMenuInteractionData()
		if len(menu.Channels()) == 0 {
			ev.Data.DeferUpdateMessage()

			ctx.Client().Rest().UpdateMessage(
				msg.ChannelID,
				msg.ID,
				discord.MessageUpdate{
					Content: cmdPack.Get("errNoChannel"),
					Embeds: json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				},
			)

			return
		}

		ev.Data.DeferUpdateMessage()

		ctx.Client().Rest().UpdateMessage(
			msg.ChannelID,
			msg.ID,
			discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name: ctx.User().Username,
							IconURL: *ctx.User().AvatarURL(),
						}),
						Title: *cmdPack.Get("starboardCreating"),
						Color: constants.Colors.Main,
						Description: *cmdPack.Get("selectEmoji"),
						Timestamp: json.Ptr(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardData") + *cmdPack.Getf("starboardDataChannel", menu.Channels()[0].ID),
							},
						},
					},
				}),
			},
		)

		emoji, err := getEmoji(ctx)
		switch err {
			case utils.ErrTimeout: {
				ctx.Client().Rest().UpdateMessage(
					msg.ChannelID,
					msg.ID,
					discord.MessageUpdate{
						Content: cmdPack.Get("errTimeout"),
						Embeds: json.Ptr([]discord.Embed{}),
						Components: json.Ptr([]discord.ContainerComponent{}),
					},
				)

				return
			}
			case errNoValidEmoji: {
				ev.Data.DeferUpdateMessage()

				ctx.Client().Rest().UpdateMessage(
					msg.ChannelID,
					msg.ID,
					discord.MessageUpdate{
						Content: cmdPack.Get("errNoValidEmoji"),
						Embeds: json.Ptr([]discord.Embed{}),
						Components: json.Ptr([]discord.ContainerComponent{}),
					},
				)

				return
			}
			default: {
				if err != nil {
					ctx.Client().Rest().UpdateMessage(
						msg.ChannelID,
						msg.ID,
						discord.MessageUpdate{
							Content: cmdPack.Get("errCollector"),
							Embeds: json.Ptr([]discord.Embed{}),
							Components: json.Ptr([]discord.ContainerComponent{}),
						},
					)

					return
				}
			}
		}

		

		ctx.Client().Rest().UpdateMessage(
			msg.ChannelID,
			msg.ID,
			discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name: ctx.User().Username,
							IconURL: *ctx.User().AvatarURL(),
						}),
						Title: *cmdPack.Get("starboardCreating"),
						Color: constants.Colors.Main,
						Description: *cmdPack.Get("selectEmoji"),
						Timestamp: json.Ptr(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardData") + 
								*cmdPack.Getf("starboardDataChannel", menu.Channels()[0].ID) + 
								*cmdPack.Getf("starboardDataEmoji", emoji),
							},
						},
					},
				}),
			},
		)
	}()

	return nil
}

func getEmoji(ctx *handler.CommandEvent) (string, error) {
	timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 20 * time.Second)

	errC := make(chan error)
	ch := make(chan utils.CollectedEvent[events.MessageCreate])
	go utils.NewCollector(
		ctx.Client(),
		timeoutCtx,
		func(e events.MessageCreate) bool {
			return ctx.User().ID == e.Message.Author.ID &&
			ctx.ChannelID() == e.ChannelID 
		},
		cancelTimeout,
		ch,
		errC,
	)
	err, ev := <- errC, <- ch
	if err != nil {
		return "", nil
	}

	emoji := ev.Data.Message.Content
	if res := constants.DiscordEmojiRegex.FindAllString(fmt.Sprint(emoji), 2); len(res) != 0 {
		if len(res) > 1 {
			return "", errNoValidEmoji
		}

		emoji = constants.CleanIdRegex.ReplaceAllString(
			constants.DiscordEmojiIdRegex.FindString(fmt.Sprint(emoji)),
			"",
		)
	} else if res := gomoji.FindAll(emoji); res == nil && len(res) > 1 {
		return "", errNoValidEmoji
	} else {
		return "", errNoValidEmoji
	}

	return emoji, nil
}

func getChannel(ctx *handler.CommandEvent) (events.ComponentInteractionCreate, discord.ResolvedChannel, error) {
	timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 20 * time.Second)

	log.Print("it enter get channel")

	errC := make(chan error)
	ch := make(chan utils.CollectedEvent[events.ComponentInteractionCreate])
	go utils.NewCollector(
		ctx.Client(),
		timeoutCtx,
		func(e events.ComponentInteractionCreate) bool {
			return e.ChannelSelectMenuInteractionData().CustomID() == "starboard:channel" 
			// e.User().ID == ctx.User().ID &&
			// e.ChannelID() == ctx.ChannelID()
		},
		cancelTimeout,
		ch,
		errC,
	)
	err, ev := <- errC, <- ch
	if err != nil {
		return events.ComponentInteractionCreate{}, discord.ResolvedChannel{}, err
	}

	log.Print("it gets out of the new collector")

	menu := ev.Data.ChannelSelectMenuInteractionData()
	if len(menu.Channels()) == 0 {
		return events.ComponentInteractionCreate{}, discord.ResolvedChannel{}, errNoChannel
	}

	log.Print("returns the channel data")

	return ev.Data, menu.Channels()[0], nil
}
