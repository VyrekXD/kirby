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
	guildData := models.GuildConfig{Lang: "es-MX"}
	starboards := []models.Starboard{}
	err := models.GuildConfigColl().FindByID(ctx.GuildID().String(), &guildData)
	if err == mongo.ErrNoDocuments {
		ctx.UpdateInteractionResponse(discord.MessageUpdate{
			Content: langs.Pack(guildData.Lang).Command("starboard").SubCommand("manual").Getf("errNoGuildData", err),
		})

		return nil
	} else {
		err := models.StarboardColl().SimpleFind(&starboards, bson.M{"guild_id": ctx.GuildID().String()})
		if err != nil && err != mongo.ErrNoDocuments {
			ctx.UpdateInteractionResponse(discord.MessageUpdate{
				Content: langs.Pack(guildData.Lang).Command("starboard").SubCommand("manual").Getf("errFindGuildStarboards", err),
			})

			return nil
		}
	}

	cmdPack := langs.Pack(guildData.Lang).Command("starboard").SubCommand("interactivo")
	
	_, err = ctx.UpdateInteractionResponse(discord.MessageUpdate{
		Embeds: json.Ptr([]discord.Embed{
			{
				Author: json.Ptr(discord.EmbedAuthor{
					Name: ctx.User().Username,
					IconURL: *ctx.User().AvatarURL(),
				}),
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
		return nil
	}

	channel, newCtx, err := GetChannel(ctx)
	if err != nil && err == errTimeout {
		ctx.CreateMessage(discord.NewMessageCreateBuilder().
		SetContent(*cmdPack.Get("errTimeout")).
		Build())
		
		return nil
	} else if err != nil && err == errNoChannel {
		if reflect.ValueOf(newCtx).IsZero() {
			ctx.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent(*cmdPack.Get("errNoChannel")).
			Build())
		
			return nil
		}
		
		newCtx.CreateMessage(discord.NewMessageCreateBuilder().
		SetContent(*cmdPack.Get("errNoChannel")).
		Build())

		return nil
	} else if err != nil {
		if reflect.ValueOf(newCtx).IsZero() {
			ctx.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent(*cmdPack.Getf("errCollector", err.Error())).
			Build())
		
			return nil
		}
		
		newCtx.CreateMessage(discord.NewMessageCreateBuilder().
		SetContent(*cmdPack.Getf("errCollector", err.Error())).
		Build())

		return nil
	}

	newCtx.CreateMessage(discord.NewMessageCreateBuilder().
	SetContent(fmt.Sprintf("Channel Name: %v\nInteraction ID: %v", channel.Name, newCtx.ID())).
	Build())

	return nil
}

func GetChannel(ctx *handler.CommandEvent,) (discord.ResolvedChannel, events.ComponentInteractionCreate, error) {
	ch := make(chan discord.ResolvedChannel)
	defer close(ch)

	err := make(chan error)
	defer close(err)

	newCtx := make(chan events.ComponentInteractionCreate)
	close(newCtx)

	go func() {
		eventCh, stop := bot.NewEventCollector(ctx.Client(), func(e *events.ComponentInteractionCreate) bool {
			return e.ChannelSelectMenuInteractionData().CustomID() == "starboard:channel" && e.User().ID == ctx.User().ID
		})
		defer stop()

		timeoutCtx, stopCtx := context.WithTimeout(context.Background(), 20 * time.Second)
		defer stopCtx()

		for {
			select {
				case <- timeoutCtx.Done(): {
					ch <- discord.ResolvedChannel{}
					newCtx <- events.ComponentInteractionCreate{}
					err <- errTimeout

					return
				}
				case compEvent := <- eventCh: {
					smenu := compEvent.ChannelSelectMenuInteractionData()

					if len(smenu.Channels()) <= 0 {
						ch <- discord.ResolvedChannel{}
						newCtx <- *compEvent
						err <- errNoChannel

						return
					}

					ch <- smenu.Channels()[0]
					newCtx <- *compEvent
					err <- bson.ErrDecodeToNil

					return
				}
			}
		}
	}()

	return <-ch, <-newCtx, <-err
}