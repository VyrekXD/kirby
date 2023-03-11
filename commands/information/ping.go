package information

import (
	"context"
	"fmt"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"
	"github.com/kamva/mgm/v3"
	"github.com/rs/zerolog/log"
	"github.com/vyrekxd/kirby/constants"
)

var Ping = discord.SlashCommandCreate{
	Name:        "ping",
	Description: "Obten la latencia del bot",
	DescriptionLocalizations: map[discord.Locale]string{
		discord.LocaleEnglishUS: "Obtain the latency of the bot",
		discord.LocaleEnglishGB: "Obtain the latency of the bot",
	},
}

func emojiPing(p int64) string {
	if p <= 90 {
		return "🟢"
	} else if p > 90 && p <= 150 {
		return "🟠"
	} else if p > 150 && p < 200 {
		return "🔴"
	} else {
		return "⚫"
	}
}

func PingHandler(ctx *handler.CommandEvent) error {
	msgTime := time.Now()
	err := ctx.DeferCreateMessage(false)
	if err != nil {
		log.Error().
			Err(err).
			Msgf(`Error when trying to defer message in command "%v": `, Ping.Name)
		return err
	}
	msgPing := time.Since(msgTime)

	_, client, _, err := mgm.DefaultConfigs()
	if err != nil {
		ctx.UpdateInteractionResponse(discord.MessageUpdate{
			Content: json.Ptr(
				fmt.Sprintf(
					"Un error ocurrio tratando de obtener el cliente de la db: %v",
					err.Error(),
				),
			),
		})

		log.Error().
			Err(err).
			Msgf(`An error ocurred trying to obtain mgm defaults in command "%v": `, Ping.Name)

		return err
	}

	clientTime := time.Now()
	err = client.Ping(context.Background(), nil)
	if err != nil {
		ctx.UpdateInteractionResponse(discord.MessageUpdate{
			Content: json.Ptr(
				fmt.Sprintf(
					"Un error ocurrio tratando de obtener el ping de la db: %v",
					err.Error(),
				),
			),
		})

		log.Error().
			Err(err).
			Msgf(`An error ocurred trying to obtain ping from db in command "%v": `, Ping.Name)

		return err
	}
	dbPing := time.Since(clientTime)

	gatewayPing := ctx.Client().Gateway().Latency()

	restTime := time.Now()
	_, err = ctx.Client().Rest().GetCurrentUser("")
	if err != nil {
		ctx.UpdateInteractionResponse(discord.MessageUpdate{
			Content: json.Ptr(
				fmt.Sprintf(
					"Un error ocurrio tratando de obtener el usuario de bot (@me): %v",
					err.Error(),
				),
			),
		})

		log.Error().
			Err(err).
			Msgf(`An error ocurred trying to obtain current user (@me) in command "%v": `, Ping.Name)

		return err
	}
	restPing := time.Since(restTime)

	_, err = ctx.UpdateInteractionResponse(discord.MessageUpdate{
		Embeds: json.Ptr([]discord.Embed{
			{
				Author: json.Ptr(discord.EmbedAuthor{
					Name:    ctx.User().Username,
					IconURL: *ctx.User().AvatarURL(),
				}),
				Title:       "Latencia de Kirby",
				Color:       constants.Colors.Main,
				Description: "Estos numeros son un aproximado de la latencia de las conexiones del bot.",
				Fields: []discord.EmbedField{
					{
						Name: "📤 Gateway",
						Value: fmt.Sprintf(
							"%v ms (`%v`) ms",
							gatewayPing.Milliseconds(),
							emojiPing(gatewayPing.Milliseconds()),
						),
						Inline: json.Ptr(true),
					},
					{
						Name:   "\u0020",
						Value:  "\u0020",
						Inline: json.Ptr(true),
					},
					{
						Name: "📡 Discord API",
						Value: fmt.Sprintf(
							"%v ms (`%v`)",
							restPing.Milliseconds(),
							emojiPing(restPing.Milliseconds()),
						),
						Inline: json.Ptr(true),
					},
					{
						Name: "📨 Delay de Mensajes",
						Value: fmt.Sprintf(
							"%v ms (`%v`)",
							msgPing.Milliseconds(),
							emojiPing(msgPing.Milliseconds()),
						),
						Inline: json.Ptr(true),
					},
					{
						Name:   "\u0020",
						Value:  "\u0020",
						Inline: json.Ptr(true),
					},
					{
						Name: "📦 Base de Datos",
						Value: fmt.Sprintf(
							"%v ms (`%v`)",
							dbPing.Milliseconds(),
							emojiPing(dbPing.Milliseconds()),
						),
						Inline: json.Ptr(true),
					},
				},
			},
		},
		),
	})
	if err != nil {
		log.Error().
			Err(err).
			Msgf(`Error when trying to respond succesfull in command "%v" `, Ping.Name)

		return err
	}

	return nil

}
