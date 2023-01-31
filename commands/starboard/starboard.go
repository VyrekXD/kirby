package starboard

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/rs/zerolog/log"
	"github.com/vyrekxd/kirby/database/models"
	"github.com/vyrekxd/kirby/langs"
	"go.mongodb.org/mongo-driver/mongo"
)

var Starboard = discord.SlashCommandCreate{
	Name:        "starboard",
	Description: "Crea, edita o elimina una starboard, de forma interactiva o manual.",
	DescriptionLocalizations: map[discord.Locale]string{
		discord.LocaleEnglishUS: "Create a starboard, interactively or manually.",
		discord.LocaleEnglishGB: "Create a starboard, interactively or manually.",
	},
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionSubCommand{
			Name: "interactivo",
			NameLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS: "interactive",
				discord.LocaleEnglishGB: "interactive",
			},
			Description: "Crea una starboard de forma interactiva.",
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS: "Create a starboard interactively.",
				discord.LocaleEnglishGB: "Create a starboard interactively.",
			},
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "manual",
			Description: "Crea una starboard de forma manual.",
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleEnglishUS: "Create a starboard manually.",
				discord.LocaleEnglishGB: "Create a starboard manually.",
			},
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionChannel{
					Name: "canal",
					NameLocalizations: map[discord.Locale]string{
						discord.LocaleEnglishUS: "canal",
						discord.LocaleEnglishGB: "canal",
					},
					Description: "El canal de la starboard.",
					DescriptionLocalizations: map[discord.Locale]string{
						discord.LocaleEnglishUS: "The channel of the starboard.",
						discord.LocaleEnglishGB: "The channel of the starboard.",
					},
					Required:     true,
					ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildText, discord.ChannelTypeGuildNews},
				},
				discord.ApplicationCommandOptionString{
					Name:        "emoji",
					Description: `El emoji de la starboard. Formato: "<:nombre:id>", ":emoji_default:" o "🐶"`,
					DescriptionLocalizations: map[discord.Locale]string{
						discord.LocaleEnglishUS: `The emoji of the starboard. Formats: "<:name:id>", ":default_emoji:" or "🐶".`,
						discord.LocaleEnglishGB: `The emoji of the starboard. Formats: "<:name:id>", ":default_emoji:" or "🐶".`,
					},
					Required: true,
				},
				discord.ApplicationCommandOptionInt{
					Name: "requeridos",
					NameLocalizations: map[discord.Locale]string{
						discord.LocaleEnglishUS: "required",
						discord.LocaleEnglishGB: "required",
					},
					Description: "Las reacciones requeridas para salir en la starboard. Debe ser mas de 0.",
					DescriptionLocalizations: map[discord.Locale]string{
						discord.LocaleEnglishUS: "The reactions required to be on the starboard. Needs to be more than 0.",
						discord.LocaleEnglishGB: "The reactions required to be on the starboard. Needs to be more than 0.",
					},
					Required: true,
				},
				discord.ApplicationCommandOptionInt{
					Name: "requeridos-para-quedarse",
					NameLocalizations: map[discord.Locale]string{
						discord.LocaleEnglishUS: "required-to-stay",
						discord.LocaleEnglishGB: "required-to-stay",
					},
					Description: "Las reacciones req. para no salir en la starboard. Debe ser mas que las reacciones \"requeridas\".",
					DescriptionLocalizations: map[discord.Locale]string{
						discord.LocaleEnglishUS: "The reactions required to stay on the starboard. Needs to be more than \"required\".",
						discord.LocaleEnglishGB: "The reactions required to stay on the starboard. Needs to be more than \"required\".",
					},
					Required: true,
				},
				discord.ApplicationCommandOptionString{
					Name: "nombre",
					NameLocalizations: map[discord.Locale]string{
						discord.LocaleEnglishUS: "name",
						discord.LocaleEnglishGB: "name",
					},
					Description: "El nombre de la starboard. Si no es ingresada se usara el nombre del canal.",
					DescriptionLocalizations: map[discord.Locale]string{
						discord.LocaleEnglishUS: "The name of the starboard. If you dont put one the name would be the name of the channel.",
						discord.LocaleEnglishGB: "The name of the starboard. If you dont put one the name would be the name of the channel.",
					},
					Required: false,
				},
				discord.ApplicationCommandOptionBool{
					Name: "bots-reacciones",
					NameLocalizations: map[discord.Locale]string{
						discord.LocaleEnglishUS: "bots-reactions",
						discord.LocaleEnglishGB: "bots-reactions",
					},
					Description: "Si las reacciones de bots deberian ser contadas.",
					DescriptionLocalizations: map[discord.Locale]string{
						discord.LocaleEnglishUS: "If bots reactions count.",
						discord.LocaleEnglishGB: "If bots reactions count.",
					},
					Required: false,
				},
				discord.ApplicationCommandOptionBool{
					Name: "bots-mensajes",
					NameLocalizations: map[discord.Locale]string{
						discord.LocaleEnglishUS: "bots-messages",
						discord.LocaleEnglishGB: "bots-messages",
					},
					Description: "Si los mensajes de bots pueden salir en la starboard.",
					DescriptionLocalizations: map[discord.Locale]string{
						discord.LocaleEnglishUS: "If bots messages can be on the starboard.",
						discord.LocaleEnglishGB: "If bots messages can be on the starboard.",
					},
					Required: false,
				},
				discord.ApplicationCommandOptionString{
					Name:        "embed-color",
					Description: "El color del embed de la starboard.",
					DescriptionLocalizations: map[discord.Locale]string{
						discord.LocaleEnglishUS: "The embed color of the starboard.",
						discord.LocaleEnglishGB: "The embed color of the starboard.",
					},
					Required: false,
				},
				discord.ApplicationCommandOptionBool{
					Name: "lista-tipo",
					NameLocalizations: map[discord.Locale]string{
						discord.LocaleEnglishUS: "list-type",
						discord.LocaleEnglishGB: "list-type",
					},
					// True (Blanco): Solo los mensajes en los canales de la lista podran estar en esta starboard. False (Negro): Solo mensajes que no esten en estos canales podran estar en esta starboard. Solo UN canal puede estar en la lista tipo false (negro).
					Description: "El tipo de lista. Usa el comando \"/infolista\" para mas informacion de las listas.",
					DescriptionLocalizations: map[discord.Locale]string{
						// True (White): Only messages from the channels in the list can be on the starboard. False (Black): Only messages that are NOT in the channel list can be on the starboard. Only ONE channel can be on list type false (black).
						discord.LocaleEnglishUS: "The list type. Use the command \"/listinfo\" for more information about lists.",
						discord.LocaleEnglishGB: "The list type. Use the command \"/listinfo\" for more information about lists.",
					},
					Required: false,
				},
				discord.ApplicationCommandOptionString{
					Name: "canales-lista",
					NameLocalizations: map[discord.Locale]string{
						discord.LocaleEnglishUS: "channels-list",
						discord.LocaleEnglishGB: "channels-list",
					},
					Description: `Los canales en la lista; Separalos por "," o " " (espacios). Usa "/lista-tipo" para mas informacion.`,
					DescriptionLocalizations: map[discord.Locale]string{
						discord.LocaleEnglishUS: "The channels on the list; Separe them by \",\" o \" \" (space). Use \"list-type\" for more information.",
						discord.LocaleEnglishGB: "The channels on the list; Separe them by \",\" o \" \" (space). Use \"list-type\" for more information.",
					},
					Required: false,
				},
			},
		},
	},
}

func StarboardHandler(ctx *handler.CommandEvent) error {
	guildData := models.GuildConfig{}
	err := models.GuildConfigColl().FindByID(ctx.GuildID().String(), &guildData)
	if err == mongo.ErrNoDocuments {
		guildData = models.GuildConfig{DefaultModel: models.DefaultModel{ID: ctx.GuildID().String()}, Lang: "es-MX"}
		err := models.GuildConfigColl().Create(&guildData)
		if err != nil {
			ctx.CreateMessage(discord.NewMessageCreateBuilder().
				SetContent(*langs.Pack(guildData.Lang).Command("starboard").Getf("errCreateGuild", err)).
				Build(),
			)

			return nil
		}
	}

	err = ctx.DeferCreateMessage(false)
	if err != nil {
		log.Error().Err(err).Msgf(`Error when trying to defer message in command "%v": `, Starboard.Name)
		return err
	} else if ctx.SlashCommandInteractionData().SubCommandName == nil {
		ctx.UpdateInteractionResponse(discord.MessageUpdate{
			Content: langs.Pack(guildData.Lang).Command("starboard").Get("errNoSubCommand"),
		})

		return nil
	}

	data := ctx.SlashCommandInteractionData()
	subcommand := data.SubCommandName

	switch *subcommand {
	case "interactivo":
		{
			return starboardInteractivo(ctx)
		}
	case "manual":
		{
			return starboardManual(ctx)
		}
	}

	return nil
}
