package models

import (
	"github.com/kamva/mgm/v3"
)

type StarboardMessage struct {
	DefaultModel `bson:",inline"`

	// Guild ID
	GuildId string `json:"guild_id" bson:"guild_id"`

	// Channel ID
	ChannelId string `json:"channel_id" bson:"channel_id"`

	// The ID of the starboard channel
	StarboardChId string `json:"starboard_ch_id" bson:"starboard_ch_id"`

	// The ID of the message on the starboard channel
	StarboardMsgId string `json:"starboard_msg_id" bson:"starboard_msg_id"`
}

func (smsg *StarboardMessage) CollectionName() string {
	return "starboard_messages"
}

func StarboardMessageColl() *mgm.Collection {
	return mgm.Coll(&StarboardMessage{})
}
