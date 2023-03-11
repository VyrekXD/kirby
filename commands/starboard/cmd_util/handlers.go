package cmd_util

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"
	"github.com/forPelevin/gomoji"
	"github.com/rs/zerolog/log"
	"github.com/vyrekxd/kirby/constants"
	"github.com/vyrekxd/kirby/langs"
)

func getChannelCh(
	ctx *handler.CommandEvent,
	errCh chan error,
	ctxCh chan events.ComponentInteractionCreate,
	chanCh chan discord.ResolvedChannel,
) {
	defer close(errCh)
	defer close(ctxCh)
	defer close(chanCh)

	log.Print("in getchannelch")

	collector, cancel := bot.NewEventCollector(
		ctx.Client(),
		func(e *events.ComponentInteractionCreate) bool {
			return e.Type() == discord.ComponentTypeChannelSelectMenu &&
				e.ChannelSelectMenuInteractionData().
					CustomID() ==
					collChannelId &&
				e.User().ID == ctx.User().ID
		},
	)

	timeoutCtx, cancelTimeout := defaultContext()
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
				log.Print("got coll")

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

func GetChannel(
	ctx *handler.CommandEvent,
	msg *discord.Message,
	cmdPack *langs.CommandPack,
) (events.ComponentInteractionCreate, discord.ResolvedChannel, error) {
	errCh := make(chan error)
	ctxCh := make(chan events.ComponentInteractionCreate)
	chanCh := make(chan discord.ResolvedChannel)
	log.Print("before getchannelch")
	go getChannelCh(ctx, errCh, ctxCh, chanCh)
	err, newCtx, channel := <-errCh, <-ctxCh, <-chanCh
	log.Print("got values")
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

			return events.ComponentInteractionCreate{}, discord.ResolvedChannel{}, errResolved
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

			return events.ComponentInteractionCreate{}, discord.ResolvedChannel{}, errResolved
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

				return events.ComponentInteractionCreate{}, discord.ResolvedChannel{}, errResolved
			}
		}
	}

	return newCtx, channel, nil
} 

func getEmojiCh(
	ctx events.ComponentInteractionCreate,
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

	timeoutCtx, cancelTimeout := defaultContext()
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

func GetEmoji(
	ctx events.ComponentInteractionCreate,
	msg *discord.Message,
	cmdPack *langs.CommandPack,
) (string, error) {
	errCh := make(chan error)
	strCh := make(chan string)
	go getEmojiCh(ctx, errCh, strCh)
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

			return "", errResolved
		}
	case errNoValidEmoji:
		{
			ctx.DeferUpdateMessage()

			ctx.Client().Rest().UpdateMessage(
				msg.ChannelID,
				msg.ID,
				discord.MessageUpdate{
					Content:    cmdPack.Get("errNoValidEmoji"),
					Embeds:     json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				},
			)

			return "", errResolved
		}
	default:
		{
			if err != nil {
				ctx.DeferUpdateMessage()

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

				return "", errResolved
			}
		}
	}

	return emoji, nil
}

func getNameCh(
	ctx events.ComponentInteractionCreate,
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

	timeoutCtx, cancelTimeout := defaultContext()
	defer cancelTimeout()

	skipCollector, skipCancel := skipButton(ctx)

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

func GetName(
	ctx events.ComponentInteractionCreate,
	msg *discord.Message,
	cmdPack *langs.CommandPack,
) (string, error) {
	errCh := make(chan error)
	strCh := make(chan string)
	go getNameCh(ctx, errCh, strCh)
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

			return "", errResolved
		}
	case errNoValidName:
		{
			ctx.DeferUpdateMessage()

			ctx.Client().Rest().UpdateMessage(
				msg.ChannelID,
				msg.ID,
				discord.MessageUpdate{
					Content:    cmdPack.Get("errNoValidName"),
					Embeds:     json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				},
			)

			return "", errResolved
		}
	default:
		{
			if err != nil {
				ctx.DeferUpdateMessage()

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

				return "", errResolved
			}
		}
	}

	return name, nil
}

func getIntCh(
	ctx events.ComponentInteractionCreate,
	errCh chan error,
	intCh chan int,
) {
	defer close(errCh)
	defer close(intCh)

	collector, cancel := bot.NewEventCollector(
		ctx.Client(),
		func(e *events.MessageCreate) bool {
			return ctx.User().ID == e.Message.Author.ID &&
				ctx.ChannelID() == e.ChannelID
		},
	)

	timeoutCtx, cancelTimeout := defaultContext()
	defer cancelTimeout()

	for {
		select {
		case <-timeoutCtx.Done():
			{
				cancel()

				errCh <- errTimeout
				intCh <- 0

				return
			}
		case msgEvent := <-collector:
			{
				cancel()

				num, err := strconv.Atoi(msgEvent.Message.Content)
				if err != nil {
					errCh <- errNoValidNumber
					intCh <- 0

					return
				}

				errCh <- nil
				intCh <- num

				return
			}
		}
	}
}

func GetInt(
	ctx events.ComponentInteractionCreate,
	msg *discord.Message,
	cmdPack *langs.CommandPack,
) (int, error) {
	errCh := make(chan error)
	intCh := make(chan int)
	go getIntCh(ctx, errCh, intCh)
	err, val := <-errCh, <-intCh
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

			return 0, errResolved
		}
	case errNoValidNumber:
		{
			ctx.DeferUpdateMessage()

			ctx.Client().Rest().UpdateMessage(
				msg.ChannelID,
				msg.ID,
				discord.MessageUpdate{
					Content:    cmdPack.Get("errNoValidEmoji"),
					Embeds:     json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				},
			)

			return 0, errResolved
		}
	default:
		{
			if err != nil {
				ctx.DeferUpdateMessage()

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

				return 0, errResolved
			}
		}
	}

	return val, nil
}

func getBoolCh(
	ctx events.ComponentInteractionCreate,
	errCh chan error,
	boolCh chan bool,
) {
	defer close(errCh)
	defer close(boolCh)

	collector, cancel := bot.NewEventCollector(
		ctx.Client(),
		func(e *events.ComponentInteractionCreate) bool {
			return e.Type() == discord.InteractionTypeComponent &&
				(e.ButtonInteractionData().CustomID() == collYesId ||
					e.ButtonInteractionData().CustomID() == collNoId ||
					e.ButtonInteractionData().CustomID() == collSkipId) &&
				ctx.User().ID == e.Message.Author.ID &&
				ctx.ChannelID() == e.ChannelID()
		},
	)
	timeoutCtx, cancelTimeout := defaultContext()
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
		case buttonEvent := <-collector:
			{
				cancel()

				button := buttonEvent.ButtonInteractionData()

				switch {
				case button.CustomID() == collYesId:
					{

						errCh <- nil
						boolCh <- true

						return
					}
				case button.CustomID() == collSkipId:
				case button.CustomID() == collNoId:
					{
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

func GetBool(
	ctx events.ComponentInteractionCreate,
	msg *discord.Message,
	cmdPack *langs.CommandPack,
) (bool, error) {
	errCh := make(chan error)
	boolCh := make(chan bool)
	go getBoolCh(ctx, errCh, boolCh)
	err, val := <-errCh, <-boolCh
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

			return false, errResolved
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

				return false, errResolved
			}
		}
	}

	return val, nil
}

func getColorCh(
	ctx events.ComponentInteractionCreate,
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
	timeoutCtx, cancelTimeout := defaultContext()
	defer cancelTimeout()

	skipCollector, skipCancel := skipButton(ctx)

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

				var parsedEmbedColor int

				if( len(msgEvent.Message.Content) != 4 && len(msgEvent.Message.Content) != 7) ||
				constants.HexColorRegex.FindString(msgEvent.Message.Content) == "" {
					errCh <- errNoValidHex
					numCh <- 0
				} else if len(msgEvent.Message.Content) == 4 {
					fembedColor := ""

					for _, d := range strings.Split(strings.ToLower(strings.Replace(msgEvent.Message.Content, "#", "", 1)), "") {
						fembedColor += strings.Repeat(d, 2)
					}
			
					parsedInt, err := strconv.ParseInt(fembedColor, 16, 64)
					if err != nil {
						errCh <- errNoValidParsedHex
						numCh <- 0
			
						return
					}
			
					parsedEmbedColor = int(parsedInt)
				} else {
					parsedInt, err := strconv.ParseInt(strings.ToLower(strings.Replace(msgEvent.Message.Content, "#", "", 1)), 16, 64)
					if err != nil {
						errCh <- errNoValidParsedHex
						numCh <- 0
			
						return 
					}

					parsedEmbedColor = int(parsedInt)
				}

				errCh <- nil
				numCh <- parsedEmbedColor

				return
			}
		case <-skipCollector:
			{
				cancel()
				skipCancel()

				errCh <- nil
				numCh <- 0

				return
			}
		}
	}
}

func GetColor(
	ctx events.ComponentInteractionCreate,
	msg *discord.Message,
	cmdPack *langs.CommandPack,
	guildLang string,
) (int, error) {
	errCh := make(chan error)
	intCh := make(chan int)
	go getColorCh(ctx, errCh, intCh)
	err, color := <-errCh, <-intCh
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

			return 0, errResolved
		}
	case errNoValidHex: {
		ctx.Client().Rest().UpdateMessage(
			msg.ChannelID,
			msg.ID,
			discord.MessageUpdate{
				Content:    langs.Pack(guildLang).Command("starboard").SubCommand("manual").Get("noValidHex"),
				Embeds:     json.Ptr([]discord.Embed{}),
				Components: json.Ptr([]discord.ContainerComponent{}),
			},
		)

		return 0, errResolved
	}
	case errNoValidHex: {
		ctx.Client().Rest().UpdateMessage(
			msg.ChannelID,
			msg.ID,
			discord.MessageUpdate{
				Content:    langs.Pack(guildLang).Command("starboard").SubCommand("manual").Get("noValidParsedHex"),
				Embeds:     json.Ptr([]discord.Embed{}),
				Components: json.Ptr([]discord.ContainerComponent{}),
			},
		)

		return 0, errResolved
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

				return 0, errResolved
			}
		}
	}

	return color, nil
}