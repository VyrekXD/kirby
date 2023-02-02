package starboard

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"
	"github.com/rs/zerolog/log"
	"github.com/vyrekxd/kirby/constants"
	"github.com/vyrekxd/kirby/database/models"
	"github.com/vyrekxd/kirby/langs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	errTimeout = errors.New("timeout")
	errNoChannel = errors.New("noChannel")
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
					Title: *cmdPack.Get("starboardData"),
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

		errCh := make(chan error)
		ctxCh := make(chan events.ComponentInteractionCreate)
		chanCh := make(chan discord.ResolvedChannel)
		go GetChannel(ctx, errCh, ctxCh, chanCh)
		err, newCtx, ch := <-errCh, <-ctxCh, <-chanCh

		switch err {
			case errTimeout: {
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
			case errNoChannel: {
				newCtx.CreateMessage(discord.MessageCreate{
					Content: *cmdPack.Get("errNoChannel"),
					Components: []discord.ContainerComponent{},
				})

				return
			}
			default: {
				if err != nil {
					if reflect.ValueOf(newCtx).IsZero() {
						ctx.CreateMessage(discord.MessageCreate{
							Content: *cmdPack.Getf("errCollector", err.Error()),
							Components: []discord.ContainerComponent{},
						})
	
						return
					} else {
						newCtx.CreateMessage(discord.MessageCreate{
							Content: *cmdPack.Getf("errCollector", err.Error()),
							Components: []discord.ContainerComponent{},
						})
	
						return
					}
				}
			}
		}

		newCtx.DeferUpdateMessage()

		ctx.Client().Rest().UpdateMessage(
			msg.ChannelID,
			msg.ID,
			discord.MessageUpdate{
				Content: json.Ptr(fmt.Sprintf("Channel Name: %v\nInteracion ID: %v", ch.ID, newCtx.ID())),
				Embeds: json.Ptr([]discord.Embed{}),
				Components: json.Ptr([]discord.ContainerComponent{}),
			},
		)

	}()

	return nil
}

func GetChannel(
	ctx *handler.CommandEvent,
	errCh chan error,
	ctxCh chan events.ComponentInteractionCreate,
	chanCh chan discord.ResolvedChannel,
) {
	defer close(errCh)
	defer close(ctxCh)
	defer close(chanCh)

	collector, cancel := bot.NewEventCollector(
		ctx.Client(),
		func(e *events.ComponentInteractionCreate) bool {
			return e.ChannelSelectMenuInteractionData().CustomID() == "starboard:channel" &&
			e.User().ID == ctx.User().ID
		},
	)
	timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 20 * time.Second)
	defer cancelTimeout()

	for {
		select {
			case <- timeoutCtx.Done(): {
				cancel()

				errCh <- errTimeout
				ctxCh <- events.ComponentInteractionCreate{}
				chanCh <- discord.ResolvedChannel{}

				return
			}
			case componentEvent := <- collector: {
				cancel()

				menu := componentEvent.ChannelSelectMenuInteractionData()
				if len(menu.Channels()) <= 0 {
					errCh <- errNoChannel
					ctxCh <- *componentEvent
					chanCh <- discord.ResolvedChannel{}

					return
				}

				errCh <- nil
				ctxCh <- *componentEvent
				chanCh <- menu.Channels()[0]

				return
			}
		}
	}
}
