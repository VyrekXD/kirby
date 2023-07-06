package components

import (
	"strconv"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"
	"github.com/disgoorg/snowflake/v2"
	"github.com/rs/zerolog/log"
	"github.com/vyrekxd/kirby/commands/starboard"
	"github.com/vyrekxd/kirby/constants"
	"github.com/vyrekxd/kirby/database/models"
	"github.com/vyrekxd/kirby/langs"
	"github.com/vyrekxd/kirby/utils"
)

func Modal(e *handler.ModalEvent) error {
	guildData := models.GuildConfig{
		Lang: "es-MX",
	}
	err := models.GuildConfigColl().
		FindByID(e.GuildID().String(), &guildData)
	if err != nil {
		return nil
	}

	cmdPack := langs.Pack(guildData.Lang).
		Command("starboard").
		SubCommand("interactivo")
	data := &models.TempStarboard{}
	err = models.TempStarboardColl().FindByID(e.Variables["id"], data)
	if err != nil {
		starboard.DeleteTempStarboard(data)
		e.UpdateMessage(discord.MessageUpdate{
			Content:    cmdPack.Getf("errUnexpected", err),
			Embeds:     json.Ptr([]discord.Embed{}),
			Components: json.Ptr([]discord.ContainerComponent{}),
		})

		return nil

	}
	if data.UserId != e.User().ID.String() {
		return nil
	}

	name := e.Data.Text(starboard.NameInputId)
	if name == "" {
		ch, err := e.Client().
			Rest().
			GetChannel(snowflake.MustParse(data.ChannelId))
		if err != nil {
			starboard.DeleteTempStarboard(data)
			e.UpdateMessage(discord.MessageUpdate{
				Content:    cmdPack.Getf("errUnexpected", err),
				Embeds:     json.Ptr([]discord.Embed{}),
				Components: json.Ptr([]discord.ContainerComponent{}),
			})

			return nil
		}

		name = ch.Name()
	}

	requiredStr := e.Data.Text(starboard.RequiredInputId)
	required, err := strconv.Atoi(requiredStr)
	if err != nil {
		starboard.DeleteTempStarboard(data)
		e.UpdateMessage(discord.MessageUpdate{
			Content:    cmdPack.Get("errNoValidNumber"),
			Embeds:     json.Ptr([]discord.Embed{}),
			Components: json.Ptr([]discord.ContainerComponent{}),
		})

		return nil
	}

	if required < 1 || required > 100 {
		starboard.DeleteTempStarboard(data)
		e.UpdateMessage(discord.MessageUpdate{
			Content:    cmdPack.Get("errNoValidRequired"),
			Embeds:     json.Ptr([]discord.Embed{}),
			Components: json.Ptr([]discord.ContainerComponent{}),
		})

		return nil
	}

	data.Name = name
	data.Required = required
	data.Phase = models.PhaseModal
	err = models.TempStarboardColl().Update(data)
	if err != nil {
		starboard.DeleteTempStarboard(data)
		e.UpdateMessage(discord.MessageUpdate{
			Content:    cmdPack.Getf("errCantUpdate", err),
			Embeds:     json.Ptr([]discord.Embed{}),
			Components: json.Ptr([]discord.ContainerComponent{}),
		})

		return nil
	}

	err = e.UpdateMessage(discord.MessageUpdate{
		Embeds: json.Ptr([]discord.Embed{
			{
				Author: json.Ptr(discord.EmbedAuthor{
					Name:    e.User().Username,
					IconURL: *e.User().AvatarURL(),
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
							*cmdPack.Getf("starboardDataChannel", data.ChannelId) +
							*cmdPack.Getf("starboardDataName", name),
						Inline: json.Ptr(true),
					},
					{
						Name: "\u0020",
						Value: *cmdPack.Get("starboardRequisites") +
							*cmdPack.Getf("starboardRequisitesRequired", required),
						Inline: json.Ptr(true),
					},
				},
			},
		}),
		Components: json.Ptr([]discord.ContainerComponent{
			discord.NewActionRow(
				discord.NewPrimaryButton(
					*langs.Pack(guildData.Lang).GetGlobal("yes"),
					starboard.YesButtonId+"/"+data.ID.Hex(),
				),
				discord.NewPrimaryButton(
					*langs.Pack(guildData.Lang).GetGlobal("no"),
					starboard.NoButtonId+"/"+data.ID.Hex(),
				),
				discord.NewSecondaryButton(
					*cmdPack.Get("skip"),
					starboard.SkipButtonId+"/"+data.ID.Hex(),
				),
			),
		}),
	})
	if err != nil {
		starboard.DeleteTempStarboard(data)
		log.Error().
			Err(err).
			Msg("Error ocurred when trying to respond in \"starboard:interactivo:channel\"")

		return nil
	}

	utils.WaitDo(time.Second*50, func() {
		find := &models.TempStarboard{}
		err := models.TempStarboardColl().First(data, find)

		if err == nil && find.Phase != models.PhaseBotsMessages {
			err := models.TempStarboardColl().Delete(find)
			if err != nil {
				log.Error().
					Err(err).
					Msg("Error ocurred when trying to delete document in \"starboard:interactivo\"")
			}

			_, errM := e.Client().Rest().UpdateMessage(
				snowflake.MustParse(find.MsgChannelId),
				snowflake.MustParse(find.MessageId),
				discord.MessageUpdate{
					Content:    cmdPack.Get("errTimeout"),
					Embeds:     json.Ptr([]discord.Embed{}),
					Components: json.Ptr([]discord.ContainerComponent{}),
				},
			)
			if errM != nil {
				log.Error().
					Err(err).
					Msg("Error ocurred when trying to edit message in \"starboard:interactivo\"")
			}
		}
	})

	return nil
}
