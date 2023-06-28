package events

// func MessageCreate(c bot.Client) bot.EventListener {
// 	return bot.NewListenerFunc(func(e *events.MessageCreate) {
// 		if !e.Message.Author.Bot {
// 			return
// 		}

// 		guildData := models.GuildConfig{
// 			Lang: "es-MX",
// 		}
// 		err := models.GuildConfigColl().
// 			FindByID(e.GuildID.String(), &guildData)
// 		if err != nil {
// 			return
// 		}

// 		cmdPack := langs.Pack(guildData.Lang).
// 			Command("starboard").
// 			SubCommand("interactivo")
// 		data := &models.TempStarboard{}
// 		err = models.TempStarboardColl().First(models.TempStarboard{
// 			GuildId: e.GuildID.String(),
// 			UserId:  e.Message.Author.ID.String(),
// 		}, data)
// 		if err != nil {
// 			log.Error().Err(err).Msg("Error ocurred when trying to find tempstarboard
// document in \"starboard:interactivo:message_create\"")

// 			return
// 		}

// 		content := e.Message.Content

// 		if res := constants.DiscordEmojiRegex.FindString(fmt.Sprint(content)); res
// != "" {
// 			content = constants.CleanIdRegex.ReplaceAllString(
// 				constants.DiscordEmojiIdRegex.FindString(
// 					fmt.Sprint(content),
// 				),
// 				"",
// 			)
// 		} else if res := gomoji.FindAll(content); len(res) > 1 || len(res) == 0 {
// 			starboard.DeleteTempStarboard(data)
// 			_, err = c.Rest().UpdateMessage(
// 				snowflake.MustParse(data.MsgChannelId),
// 				snowflake.MustParse(data.MessageId),
// 				discord.MessageUpdate{
// 					Content:    cmdPack.Get("errNoValidEmoji"),
// 					Embeds:     json.Ptr([]discord.Embed{}),
// 					Components: json.Ptr([]discord.ContainerComponent{}),
// 				},
// 			)
// 			if err != nil {
// 				log.Error().Err(err).Msg("Error ocurred when trying to update main
// message in \"starboard:interactivo:message_create\"")
// 			}

// 			return
// 		}

// 		data.Emoji = content
// 		err = models.TempStarboardColl().Update(data)
// 		if err != nil {
// 			starboard.DeleteTempStarboard(data)
// 			c.Rest().UpdateMessage(snowflake.MustParse(data.MsgChannelId),
// e.Message.ID, discord.MessageUpdate{
// 				Content:    cmdPack.Getf("errCantUpdate", err),
// 				Embeds:     json.Ptr([]discord.Embed{}),
// 				Components: json.Ptr([]discord.ContainerComponent{}),
// 			})

// 			return
// 		}

// 		_, err = c.Rest().UpdateMessage(snowflake.MustParse(data.MsgChannelId),
// e.Message.ID, discord.MessageUpdate{
// 			Embeds: json.Ptr([]discord.Embed{
// 				{
// 					Author: json.Ptr(discord.EmbedAuthor{
// 						Name:    e.Message.Author.Username,
// 						IconURL: *e.Message.Author.AvatarURL(),
// 					}),
// 					Title: *cmdPack.Get("starboardCreating"),
// 					Color: constants.Colors.Main,
// 					Description: *cmdPack.Get("selectName") +
// 						*cmdPack.Get("optionalParam"),
// 					Timestamp: json.Ptr(time.Now()),
// 					Fields: []discord.EmbedField{
// 						{
// 							Name: "\u0020",
// 							Value: *cmdPack.Get("starboardData") +
// 								*cmdPack.Getf("starboardDataChannel", data.ChannelId) +
// 								*cmdPack.Getf("starboardDataEmoji", content),
// 						},
// 					},
// 				},
// 			}),
// 			Components: json.Ptr([]discord.ContainerComponent{
// 				discord.NewActionRow(
// 					discord.NewSecondaryButton(
// 						*cmdPack.Get("skip"),
// 						"starboard:skip",
// 					),
// 				),
// 			}),
// 		})
// 		if err != nil {
// 			starboard.DeleteTempStarboard(data)
// 			log.Error().Err(err).Msg("Error ocurred when trying to respond in
// \"starboard:interactivo:message_create\"")

// 			return
// 		}

// 		utils.WaitDo(time.Second*50, func() {
// 			find := &models.TempStarboard{}
// 			models.TempStarboardColl().First(data, find)

// 			if find.Name == "" {
// 				err := models.TempStarboardColl().Delete(find)
// 				if err != nil {
// 					log.Error().Err(err).Msg("Error ocurred when trying to delete document
// in \"starboard:interactivo\"")
// 				}

// 				_, errM := c.Rest().UpdateMessage(
// 					snowflake.MustParse(data.MsgChannelId),
// 					snowflake.MustParse(data.MessageId),
// 					discord.MessageUpdate{
// 						Content:    cmdPack.Get("errTimeout"),
// 						Embeds:     json.Ptr([]discord.Embed{}),
// 						Components: json.Ptr([]discord.ContainerComponent{}),
// 					},
// 				)
// 				if errM != nil {
// 					log.Error().Err(err).Msg("Error ocurred when trying to edit message in
// \"starboard:interactivo:message_create\"")
// 				}
// 			}
// 		})

// 		// if len(content) > 25 {
// 		// 	starboard.DeleteTempStarboard(data)
// 		// 	_, err = c.Rest().UpdateMessage(
// 		// 		snowflake.MustParse(data.MsgChannelId),
// 		// 		snowflake.MustParse(data.MessageId),
// 		// 		discord.MessageUpdate{
// 		// 			Content:    cmdPack.Get("errNoValidName"),
// 		// 			Embeds:     json.Ptr([]discord.Embed{}),
// 		// 			Components: json.Ptr([]discord.ContainerComponent{}),
// 		// 		},
// 		// 	)
// 		// 	if err != nil {
// 		// 		log.Error().Err(err).Msg("Error ocurred when trying to update main
// message in \"starboard:interactivo:message_create\"")
// 		// 	}

// 		// 	return
// 		// }

// 		// data.Name = content
// 		// err = models.TempStarboardColl().Update(data)
// 		// if err != nil {
// 		// 	starboard.DeleteTempStarboard(data)
// 		// 	c.Rest().UpdateMessage(snowflake.MustParse(data.MsgChannelId),
// e.Message.ID, discord.MessageUpdate{
// 		// 		Content:    cmdPack.Getf("errCantUpdate", err),
// 		// 		Embeds:     json.Ptr([]discord.Embed{}),
// 		// 		Components: json.Ptr([]discord.ContainerComponent{}),
// 		// 	})

// 		// 	return
// 		// }
// 	})
// }
