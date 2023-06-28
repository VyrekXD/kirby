package models

import "github.com/kamva/mgm/v3"

var (
	PhaseSelectChannel = 0
	PhaseModal         = 1
	PhaseBotsReact     = 2
	PhaseBotsMessages  = 3
	PhaseEmoji         = 4
	PhaseEmbedColor    = 5
)

type TempStarboard struct {
	mgm.DefaultModel `bson:",inline"`

	Phase        int    `json:"phase"          bson:"phase"`
	GuildId      string `json:"guild_id"       bson:"guild_id"`
	UserId       string `json:"user_id"        bson:"user_id"`
	MessageId    string `json:"message_id"     bson:"message_id"`
	MsgChannelId string `json:"msg_channel_id" bson:"msg_channel_id"`

	Name string `json:"name" bson:"name"`

	ChannelId string `json:"channel_id" bson:"channel_id"`

	Emoji string `json:"emoji" bson:"emoji"`

	Required       int `json:"required"         bson:"required"`
	RequiredToStay int `json:"required_to_stay" bson:"required_to_stay"`

	BotsReact    bool `json:"bots_react"    bson:"bots_react"`
	BotsMessages bool `json:"bots_messages" bson:"bots_messages"`

	EmbedColor int `json:"embed_color,omitempty" bson:"embed_color,omitempty"`

	Levels Levels `json:"levels,omitempty" bson:"levels,omitempty"`

	ChannelListType bool     `json:"channel_list_type" bson:"channel_list_type"`
	ChannelList     []string `json:"channel_list"      bson:"channel_list"`
}

func (gconfig *TempStarboard) CollectionName() string {
	return "temp_starboards"
}

func TempStarboardColl() *mgm.Collection {
	return mgm.Coll(&TempStarboard{})
}
