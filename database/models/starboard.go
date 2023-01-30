package models

import "github.com/kamva/mgm/v3"

type Starboard struct {
	mgm.DefaultModel `bson:",inline"`

	GuildId string `json:"guild_id" bson:"guild_id"`

	// Starboard name (channel name).
	Name string `json:"name" bson:"name"`

	// Starboard channel ID.
	ChannelId string `json:"channel_id" bson:"channel_id"`

	// Starboard emoji. Accepts: emoji_id or emoji_name
	Emoji string `json:"emoji" bson:"emoji"`

	// Required to appear on starboard
	Required int `json:"required" bson:"required"`
	// Required to stay on the starboard
	RequiredToStay int `json:"required_to_stay" bson:"required_to_stay"`

	// Can bots reactions count to a message appear in the starboard
	BotsReact bool `json:"bots_react" bson:"bots_react"`
	// Can bot messages be on the starboard
	BotsMessages bool `json:"bots_messages" bson:"bots_messages"`

	// Color of the embeds sent in the starboard
	EmbedColor int `json:"embed_color,omitempty" bson:"embed_color,omitempty"`

	// Levels of emojis in the starboard
	Levels Levels `json:"levels,omitempty" bson:"levels,omitempty"`

	// Works as a black/white list. NOTE: Everytime the list changes its type the array is cleaned
	// False counts as a "blacklist", NOTE: The server cant have more than one blacklist starboard.
	// - Starboard will be global and will ignore the channels on the list
	// True counts as a "whitelisDefaultModel
	// - Starboard will only detect channels on the list.
	ChannelListType bool     `json:"channel_list_type" bson:"channel_list_type"`
	ChannelList     []string `json:"channel_list" bson:"channel_list"`
}

type Levels struct {
	// Emoji used when the message meets the required reactions.
	FirstEmoji string `json:"first_emoji,omitempty" bson:"first_emoji,omitempty"`

	// NOTE: If required changes
	Second      int    `json:"second,omitempty" bson:"second,omitempty"`
	SecondEmoji string `json:"second_emoji,omitempty" bson:"second_emoji,omitempty"`

	Third      int    `json:"third,omitempty" bson:"third,omitempty"`
	ThirdEmoji string `json:"third_emoji,omitempty" bson:"third_emoji,omitempty"`
}

func (gconfig *Starboard) CollectionName() string {
	return "starboards"
}

func StarboardColl() *mgm.Collection {
	return mgm.Coll(&Starboard{})
}
