package starboard

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"
	"github.com/forPelevin/gomoji"
	"github.com/rs/zerolog/log"
	"github.com/vyrekxd/kirby/constants"
	"github.com/vyrekxd/kirby/database/models"
	"github.com/vyrekxd/kirby/langs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	errNoChannel = errors.New("noChannel")
	errNoValidEmoji = errors.New("noValidEmoji")
	errTimeout = errors.New("timeout")
	errNoValidRequired = errors.New("noValidRequired")
	errNoValidRequiredToS = errors.New("noValidRequiredToS")
)

type modalExtraData struct {
	Required int
	RequiredToStay int
	Name string
}

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

		errCh := make(chan error)
		ctxCh := make(chan events.ComponentInteractionCreate)
		chanCh := make(chan discord.ResolvedChannel)
		go getChannel(ctx, errCh, ctxCh, chanCh)
		err, newCtx, channel := <-errCh, <-ctxCh, <-chanCh
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
				newCtx.DeferUpdateMessage()

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
			default: {
				if err != nil {
					if !reflect.ValueOf(newCtx).IsZero() {
						newCtx.DeferUpdateMessage()
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
						Description: *cmdPack.Get("selectExtraData"),
						Timestamp: json.Ptr(time.Now()),
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

		err = newCtx.CreateModal(discord.ModalCreate{
			CustomID: "starboard:modal",
		})
		if err != nil {
			ctx.Client().Rest().UpdateMessage(
				msg.ChannelID,
				msg.ID,
				discord.MessageUpdate{
					Content: cmdPack.Get("errModal"),
					Embeds: json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				},
			)

			return
		}

		errCh2 := make(chan error)
		mctxCh := make(chan events.ModalSubmitInteractionCreate)
		dataCh := make(chan modalExtraData)
		go getExtraData(ctx, errCh, mctxCh, dataCh)
		err, mctx, mdata := <- errCh2, <- mctxCh, <- dataCh
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

			case errNoValidRequired: {
				mctx.DeferUpdateMessage()

				ctx.Client().Rest().UpdateMessage(
					msg.ChannelID,
					msg.ID,
					discord.MessageUpdate{
						Content: cmdPack.Get("errNoValidRequired"),
						Embeds: json.Ptr([]discord.Embed{}),
						Components: json.Ptr([]discord.ContainerComponent{}),
					},
				)

				return
			}

			case errNoValidRequiredToS: {
				mctx.DeferUpdateMessage()

				ctx.Client().Rest().UpdateMessage(
					msg.ChannelID,
					msg.ID,
					discord.MessageUpdate{
						Content: cmdPack.Get("errNoValidRequiredToS"),
						Embeds: json.Ptr([]discord.Embed{}),
						Components: json.Ptr([]discord.ContainerComponent{}),
					},
				)

				return
			}

			default: {
				if err != nil {
					if !reflect.ValueOf(newCtx).IsZero() {
						mctx.DeferUpdateMessage()
					}

					ctx.Client().Rest().UpdateMessage(
						msg.ChannelID,
						msg.ID,
						discord.MessageUpdate{
							Content: cmdPack.Get("errModal"),
							Embeds: json.Ptr([]discord.Embed{}),
							Components: json.Ptr([]discord.ContainerComponent{}),
						},
					)

					return
				}
			}
		}

		errCh = make(chan error)
		strCh := make(chan string)
		go getEmoji(ctx, errCh, strCh)
		err, emoji := <- errCh, <- strCh
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
			case errNoValidEmoji: {
				newCtx.DeferUpdateMessage()

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
					if !reflect.ValueOf(newCtx).IsZero() {
						newCtx.DeferUpdateMessage()
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
						Description: *cmdPack.Get(""),
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
			},
		)
	}()

	return nil
}

func getExtraData(
	ctx *handler.CommandEvent,
	errCh chan error,
	modalCtx chan events.ModalSubmitInteractionCreate,
	dataCh chan modalExtraData,
) {
	defer close(errCh)
	defer close(dataCh)

	collector, cancel := bot.NewEventCollector(
		ctx.Client(),
		func(e *events.ModalSubmitInteractionCreate) bool {
			return e.Data.CustomID == "starboard:modal" &&
			e.User().ID == ctx.User().ID &&
			e.ChannelID() == ctx.ChannelID()
		},
	)
	timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancelTimeout()

	for {
		select {
			case <- timeoutCtx.Done(): {
				cancel()

				errCh <- errTimeout
				modalCtx <- events.ModalSubmitInteractionCreate{}
				dataCh <- modalExtraData{}

				return
			}
			case modalEvent := <- collector: {
				cancel()


				requiredStr := modalEvent.Data.Text("starboard:modal:required")
				requiredToStayStr := modalEvent.Data.Text("starboard:modal:requiredtostay")
				nameStr := modalEvent.Data.Text("starboard:modal:name")

				required, err := strconv.Atoi(requiredStr)
				if err != nil {
					errCh <- errNoValidRequired
					modalCtx <- events.ModalSubmitInteractionCreate{}
					dataCh <- modalExtraData{}
				}

				requiredToStay, err := strconv.Atoi(requiredToStayStr)
				if err != nil {
					errCh <- errNoValidRequiredToS
					modalCtx <- events.ModalSubmitInteractionCreate{}
					dataCh <- modalExtraData{}
				}

				errCh <- nil
				modalCtx <- *modalEvent
				dataCh <- modalExtraData{
					Required: required,
					RequiredToStay: requiredToStay,
					Name: nameStr,
				}

				return
			}
		}
	}
}

func getEmoji(
	ctx *handler.CommandEvent,
	errCh chan error,
	strCh chan string,
) {
	defer close(errCh)
	defer close(strCh)

	collector, cancel := bot.NewEventCollector(
		ctx.Client(),
		func(e *events.MessageCreate) bool {
			return ctx.User().ID == e.Message.Author.ID &&
			ctx.ChannelID() == e.ChannelID 
		},
	)
	timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancelTimeout()

	for {
		select {
			case <- timeoutCtx.Done(): {
				cancel()

				errCh <- errTimeout
				strCh <- ""

				return
			}
			case msgEvent := <- collector: {
				cancel()

				emoji := msgEvent.Message.Content
				if res := constants.DiscordEmojiRegex.FindAllString(fmt.Sprint(emoji), 2); len(res) != 0 {
					if len(res) > 1 {
						errCh <- errNoValidEmoji
						strCh <- ""
					}
			
					emoji = constants.CleanIdRegex.ReplaceAllString(
						constants.DiscordEmojiIdRegex.FindString(fmt.Sprint(emoji)),
						"",
					)
				} else if res := gomoji.FindAll(emoji); len(res) > 1 || len(res) == 0 {
					errCh <- errNoValidEmoji
					strCh <- ""
				}

				errCh <- nil
				strCh <- emoji

				return
			}
		}
	}
}

func getChannel(
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
	timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 30 * time.Second)
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
