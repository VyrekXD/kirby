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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/vyrekxd/kirby/constants"
	"github.com/vyrekxd/kirby/database/models"
	"github.com/vyrekxd/kirby/langs"
)

var (
	errNoChannel     = errors.New("noChannel")
	errNoValidEmoji  = errors.New("noValidEmoji")
	errTimeout       = errors.New("timeout")
	errNoValidNumber = errors.New("noValidNumber")
	errNoValidName   = errors.New("noValidName")
)

func starboardInteractivo(ctx *handler.CommandEvent) error {
	go func() {
		guildData := models.GuildConfig{Lang: "es-MX"}
		starboards := []models.Starboard{}
		err := models.GuildConfigColl().
			FindByID(ctx.GuildID().String(), &guildData)
		if err == mongo.ErrNoDocuments {
			ctx.UpdateInteractionResponse(discord.MessageUpdate{
				Content: langs.Pack(guildData.Lang).
					Command("starboard").
					SubCommand("manual").
					Getf("errNoGuildData", err),
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
			Components: json.Ptr([]discord.ContainerComponent{
				discord.NewActionRow(
					discord.ChannelSelectMenuComponent{
						CustomID:  "starboard:channel",
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
			log.Error().
				Err(err).
				Msg("Error ocurred when trying to respond in \"starboard:interactivo\"")
			return
		}

		errCh := make(chan error)
		ctxCh := make(chan events.ComponentInteractionCreate)
		chanCh := make(chan discord.ResolvedChannel)
		go getChannel(ctx, errCh, ctxCh, chanCh)
		err, newCtx, channel := <-errCh, <-ctxCh, <-chanCh
		switch err {
		case errTimeout:
			{
				ctx.Client().Rest().UpdateMessage(
					msg.ChannelID,
					msg.ID,
					discord.MessageUpdate{
						Content:    cmdPack.Get("errTimeout"),
						Embeds:     json.Ptr([]discord.Embed{}),
						Components: json.Ptr([]discord.ContainerComponent{}),
					},
				)

				return
			}
		case errNoChannel:
			{
				newCtx.DeferUpdateMessage()

				ctx.Client().Rest().UpdateMessage(
					msg.ChannelID,
					msg.ID,
					discord.MessageUpdate{
						Content:    cmdPack.Get("errNoChannel"),
						Embeds:     json.Ptr([]discord.Embed{}),
						Components: json.Ptr([]discord.ContainerComponent{}),
					},
				)

				return
			}
		default:
			{
				if err != nil {
					if !reflect.ValueOf(newCtx).IsZero() {
						newCtx.DeferUpdateMessage()
					}

					ctx.Client().Rest().UpdateMessage(
						msg.ChannelID,
						msg.ID,
						discord.MessageUpdate{
							Content: cmdPack.Getf(
								"errCollector",
								err.Error(),
							),
							Embeds: json.Ptr([]discord.Embed{}),
							Components: json.Ptr(
								[]discord.ContainerComponent{},
							),
						},
					)

					return
				}
			}
		}

		_, err = ctx.Client().Rest().UpdateMessage(
			msg.ChannelID,
			msg.ID,
			discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name:    ctx.User().Username,
							IconURL: *ctx.User().AvatarURL(),
						}),
						Title:       *cmdPack.Get("starboardCreating"),
						Color:       constants.Colors.Main,
						Description: *cmdPack.Get("selectRequired"),
						Timestamp:   json.Ptr(time.Now()),
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
		if err != nil {
			log.Error().
				Err(err).
				Msg(`Error when trying to update message in "starboard:interactivo"`)
			return
		}

		errCh = make(chan error)
		strCh := make(chan string)
		go getEmoji(ctx, errCh, strCh)
		err, emoji := <-errCh, <-strCh
		switch err {
		case errTimeout:
			{
				ctx.Client().Rest().UpdateMessage(
					msg.ChannelID,
					msg.ID,
					discord.MessageUpdate{
						Content:    cmdPack.Get("errTimeout"),
						Embeds:     json.Ptr([]discord.Embed{}),
						Components: json.Ptr([]discord.ContainerComponent{}),
					},
				)

				return
			}
		case errNoValidEmoji:
			{
				newCtx.DeferUpdateMessage()

				ctx.Client().Rest().UpdateMessage(
					msg.ChannelID,
					msg.ID,
					discord.MessageUpdate{
						Content:    cmdPack.Get("errNoValidEmoji"),
						Embeds:     json.Ptr([]discord.Embed{}),
						Components: json.Ptr([]discord.ContainerComponent{}),
					},
				)

				return
			}
		default:
			{
				if err != nil {
					if !reflect.ValueOf(newCtx).IsZero() {
						newCtx.DeferUpdateMessage()
					}

					ctx.Client().Rest().UpdateMessage(
						msg.ChannelID,
						msg.ID,
						discord.MessageUpdate{
							Content: cmdPack.Getf(
								"errCollector",
								err.Error(),
							),
							Embeds: json.Ptr([]discord.Embed{}),
							Components: json.Ptr(
								[]discord.ContainerComponent{},
							),
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
							Name:    ctx.User().Username,
							IconURL: *ctx.User().AvatarURL(),
						}),
						Title: *cmdPack.Get("starboardCreating"),
						Color: constants.Colors.Main,
						Description: *cmdPack.Get("selectName") +
							*cmdPack.Get("optionalParam"),
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

		errCh = make(chan error)
		strCh = make(chan string)
		go getName(ctx, errCh, strCh)
		err, name := <-errCh, <-strCh
		switch err {
		case errTimeout:
			{
				ctx.Client().Rest().UpdateMessage(
					msg.ChannelID,
					msg.ID,
					discord.MessageUpdate{
						Content:    cmdPack.Get("errTimeout"),
						Embeds:     json.Ptr([]discord.Embed{}),
						Components: json.Ptr([]discord.ContainerComponent{}),
					},
				)

				return
			}
		case errNoValidName:
			{
				newCtx.DeferUpdateMessage()

				ctx.Client().Rest().UpdateMessage(
					msg.ChannelID,
					msg.ID,
					discord.MessageUpdate{
						Content:    cmdPack.Get("errNoValidName"),
						Embeds:     json.Ptr([]discord.Embed{}),
						Components: json.Ptr([]discord.ContainerComponent{}),
					},
				)

				return
			}
		default:
			{
				if err != nil {
					if !reflect.ValueOf(newCtx).IsZero() {
						newCtx.DeferUpdateMessage()
					}

					ctx.Client().Rest().UpdateMessage(
						msg.ChannelID,
						msg.ID,
						discord.MessageUpdate{
							Content: cmdPack.Getf(
								"errCollector",
								err.Error(),
							),
							Embeds: json.Ptr([]discord.Embed{}),
							Components: json.Ptr(
								[]discord.ContainerComponent{},
							),
						},
					)

					return
				}
			}
		}
		if name == "" {
			name = channel.Name
		}

		ctx.Client().Rest().UpdateMessage(
			msg.ChannelID,
			msg.ID,
			discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name:    ctx.User().Username,
							IconURL: *ctx.User().AvatarURL(),
						}),
						Title:       *cmdPack.Get("starboardCreating"),
						Color:       constants.Colors.Main,
						Description: *cmdPack.Get("selectRequired"),
						Timestamp:   json.Ptr(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardData") +
									*cmdPack.Getf("starboardDataChannel", channel.ID) +
									*cmdPack.Getf("starboardDataEmoji", emoji) +
									*cmdPack.Getf("starboardDataName", name),
							},
						},
					},
				}),
			},
		)

		errCh = make(chan error)
		numCh := make(chan int)
		go getNumber(ctx, errCh, numCh)
		err, required := <-errCh, <-numCh
		switch err {
		case errTimeout:
			{
				ctx.Client().Rest().UpdateMessage(
					msg.ChannelID,
					msg.ID,
					discord.MessageUpdate{
						Content:    cmdPack.Get("errTimeout"),
						Embeds:     json.Ptr([]discord.Embed{}),
						Components: json.Ptr([]discord.ContainerComponent{}),
					},
				)

				return
			}
		case errNoValidNumber:
			{
				newCtx.DeferUpdateMessage()

				ctx.Client().Rest().UpdateMessage(
					msg.ChannelID,
					msg.ID,
					discord.MessageUpdate{
						Content:    cmdPack.Get("errNoValidEmoji"),
						Embeds:     json.Ptr([]discord.Embed{}),
						Components: json.Ptr([]discord.ContainerComponent{}),
					},
				)

				return
			}
		default:
			{
				if err != nil {
					if !reflect.ValueOf(newCtx).IsZero() {
						newCtx.DeferUpdateMessage()
					}

					ctx.Client().Rest().UpdateMessage(
						msg.ChannelID,
						msg.ID,
						discord.MessageUpdate{
							Content: cmdPack.Getf(
								"errCollector",
								err.Error(),
							),
							Embeds: json.Ptr([]discord.Embed{}),
							Components: json.Ptr(
								[]discord.ContainerComponent{},
							),
						},
					)

					return
				}
			}
		}
		if required <= 0 {
			ctx.Client().Rest().UpdateMessage(
				msg.ChannelID,
				msg.ID,
				discord.MessageUpdate{
					Content:    cmdPack.Get("errNoValidRequired"),
					Embeds:     json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				},
			)

			return
		}

		ctx.Client().Rest().UpdateMessage(
			msg.ChannelID,
			msg.ID,
			discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name:    ctx.User().Username,
							IconURL: *ctx.User().AvatarURL(),
						}),
						Title:       *cmdPack.Get("starboardCreating"),
						Color:       constants.Colors.Main,
						Description: *cmdPack.Get("selectRequiredToS"),
						Timestamp:   json.Ptr(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardData") +
									*cmdPack.Getf("starboardDataChannel", channel.ID) +
									*cmdPack.Getf("starboardDataEmoji", emoji) +
									*cmdPack.Getf("starboardDataName", name),
							},
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardRequisites") +
									*cmdPack.Getf("starboardRequisitesRequired", required),
							},
						},
					},
				}),
			},
		)

		errCh = make(chan error)
		numCh = make(chan int)
		go getNumber(ctx, errCh, numCh)
		err, requiredToStay := <-errCh, <-numCh
		switch err {
			case errTimeout:
			{
				ctx.Client().Rest().UpdateMessage(
					msg.ChannelID,
					msg.ID,
					discord.MessageUpdate{
						Content:    cmdPack.Get("errTimeout"),
						Embeds:     json.Ptr([]discord.Embed{}),
						Components: json.Ptr([]discord.ContainerComponent{}),
					},
				)

				return
			}
			case errNoValidNumber:
			{
				newCtx.DeferUpdateMessage()

				ctx.Client().Rest().UpdateMessage(
					msg.ChannelID,
					msg.ID,
					discord.MessageUpdate{
						Content:    cmdPack.Get("errNoValidEmoji"),
						Embeds:     json.Ptr([]discord.Embed{}),
						Components: json.Ptr([]discord.ContainerComponent{}),
					},
				)

				return
			}
			default:
			{
				if err != nil {
					if !reflect.ValueOf(newCtx).IsZero() {
						newCtx.DeferUpdateMessage()
					}

					ctx.Client().Rest().UpdateMessage(
						msg.ChannelID,
						msg.ID,
						discord.MessageUpdate{
							Content: cmdPack.Getf(
								"errCollector",
								err.Error(),
							),
							Embeds: json.Ptr([]discord.Embed{}),
							Components: json.Ptr(
								[]discord.ContainerComponent{},
							),
						},
					)

					return
				}
			}
		}
		if requiredToStay < 0 || requiredToStay >= required {
			ctx.Client().Rest().UpdateMessage(
				msg.ChannelID,
				msg.ID,
				discord.MessageUpdate{
					Content:    cmdPack.Get("errNoValidRequiredToS"),
					Embeds:     json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				},
			)

			return
		}

		ctx.Client().Rest().UpdateMessage(
			msg.ChannelID,
			msg.ID,
			discord.MessageUpdate{
				Embeds: json.Ptr([]discord.Embed{
					{
						Author: json.Ptr(discord.EmbedAuthor{
							Name:    ctx.User().Username,
							IconURL: *ctx.User().AvatarURL(),
						}),
						Title: *cmdPack.Get("starboardCreating"),
						Color: constants.Colors.Main,
						Description: *cmdPack.Get("selectBotsReact") +
							*cmdPack.Get("useButtons"),
						Timestamp: json.Ptr(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardData") +
									*cmdPack.Getf("starboardDataChannel", channel.ID) +
									*cmdPack.Getf("starboardDataEmoji", emoji) +
									*cmdPack.Getf("starboardDataName", name),
							},
							{
								Name: "\u0020",
								Value: *cmdPack.Get("starboardRequisites") +
									*cmdPack.Getf("starboardRequisitesRequired", required) +
									*cmdPack.Getf("starboardRequisitesRequiredToS", requiredToStay),
							},
						},
					},
				}),
			},
		)
		
		errCh = make(chan error)
		boolCh := make(chan bool)
		go getNumber(ctx, errCh, numCh)
		err, botsReact := <- errCh, <- boolCh
		switch err {
			case errTimeout:
			{
				ctx.Client().Rest().UpdateMessage(
					msg.ChannelID,
					msg.ID,
					discord.MessageUpdate{
						Content:    cmdPack.Get("errTimeout"),
						Embeds:     json.Ptr([]discord.Embed{}),
						Components: json.Ptr([]discord.ContainerComponent{}),
					},
				)

				return
			}
			default:
			{
				if err != nil {
					ctx.Client().Rest().UpdateMessage(
						msg.ChannelID,
						msg.ID,
						discord.MessageUpdate{
							Content: cmdPack.Getf(
								"errCollector",
								err.Error(),
							),
							Embeds: json.Ptr([]discord.Embed{}),
							Components: json.Ptr(
								[]discord.ContainerComponent{},
							),
						},
					)

					return
				}
			}
		}
	}()

	return nil
}

func getBool(
	ctx *handler.CommandEvent,
	errCh chan error,
	boolCh chan bool,
) {
	defer close(errCh)
	defer close(boolCh)

	collector, cancel := bot.NewEventCollector(
		ctx.Client(),
		func(e *events.ComponentInteractionCreate) bool {
			return e.Type() == discord.InteractionTypeComponent &&
			(e.ButtonInteractionData().CustomID() == "starboard:yes" || 
			e.ButtonInteractionData().CustomID() == "starboard:no" || 
			e.ButtonInteractionData().CustomID() == "starboard:skip") &&
			ctx.User().ID == e.Message.Author.ID &&
			ctx.ChannelID() == e.ChannelID() 
		},
	)
	timeoutCtx, cancelTimeout := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancelTimeout()

	for {
		select {
		case <-timeoutCtx.Done():
			{
				cancel()

				errCh <- errTimeout
				boolCh <- false

				return
			}
			case buttonEvent := <- collector: {
				cancel()

				button := buttonEvent.ButtonInteractionData()

				switch {
					case button.CustomID() == "starboard:yes": {
						
						errCh <- nil
						boolCh <- true
						
						return
					}
					case button.CustomID() == "starboard:skip":
					case button.CustomID() == "starboard:no": {
						errCh <- nil
						boolCh <- false

						return
					}
				}


				return
			}
		}
	}
}

func getName(
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
	timeoutCtx, cancelTimeout := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancelTimeout()

	skipCollector, skipCancel := bot.NewEventCollector(
		ctx.Client(),
		func(e *events.ComponentInteractionCreate) bool {
			return e.Type() == discord.ComponentTypeButton &&
				e.ButtonInteractionData().CustomID() == "starboard:skip" &&
				ctx.User().ID == e.Message.Author.ID &&
				ctx.ChannelID() == e.ChannelID()
		},
	)

	for {
		select {
		case <-timeoutCtx.Done():
			{
				cancel()
				skipCancel()

				errCh <- errTimeout
				strCh <- ""

				return
			}
		case msgEvent := <-collector:
			{
				cancel()
				skipCancel()

				if len(msgEvent.Message.Content) > 100 {
					errCh <- errNoValidName
					strCh <- ""

					return
				}

				errCh <- nil
				strCh <- msgEvent.Message.Content

				return
			}
		case <-skipCollector:
			{
				cancel()
				skipCancel()

				errCh <- nil
				strCh <- ""

				return
			}
		}
	}
}

func getNumber(
	ctx *handler.CommandEvent,
	errCh chan error,
	numCh chan int,
) {
	defer close(errCh)
	defer close(numCh)

	collector, cancel := bot.NewEventCollector(
		ctx.Client(),
		func(e *events.MessageCreate) bool {
			return ctx.User().ID == e.Message.Author.ID &&
				ctx.ChannelID() == e.ChannelID
		},
	)
	timeoutCtx, cancelTimeout := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancelTimeout()

	for {
		select {
		case <-timeoutCtx.Done():
			{
				cancel()

				errCh <- errTimeout
				numCh <- 0

				return
			}
		case msgEvent := <-collector:
			{
				cancel()

				num, err := strconv.Atoi(msgEvent.Message.Content)
				if err != nil {
					errCh <- errNoValidNumber
					numCh <- 0

					return
				}

				errCh <- nil
				numCh <- num

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
	timeoutCtx, cancelTimeout := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancelTimeout()

	for {
		select {
		case <-timeoutCtx.Done():
			{
				cancel()

				errCh <- errTimeout
				strCh <- ""

				return
			}
		case msgEvent := <-collector:
			{
				cancel()

				emoji := msgEvent.Message.Content
				if res := constants.DiscordEmojiRegex.FindAllString(fmt.Sprint(emoji), 2); len(
					res,
				) != 0 {
					if len(res) > 1 {
						errCh <- errNoValidEmoji
						strCh <- ""

						return
					}

					emoji = constants.CleanIdRegex.ReplaceAllString(
						constants.DiscordEmojiIdRegex.FindString(
							fmt.Sprint(emoji),
						),
						"",
					)
				} else if res := gomoji.FindAll(emoji); len(res) > 1 || len(res) == 0 {
					errCh <- errNoValidEmoji
					strCh <- ""

					return
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
			return e.Type() == discord.ComponentTypeChannelSelectMenu &&
				e.ChannelSelectMenuInteractionData().
					CustomID() ==
					"starboard:channel" &&
				e.User().ID == ctx.User().ID
		},
	)
	timeoutCtx, cancelTimeout := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancelTimeout()

	for {
		select {
		case <-timeoutCtx.Done():
			{
				cancel()

				errCh <- errTimeout
				ctxCh <- events.ComponentInteractionCreate{}
				chanCh <- discord.ResolvedChannel{}

				return
			}
		case componentEvent := <-collector:
			{
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
