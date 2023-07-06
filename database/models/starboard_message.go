package models

import (
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StarboardMessage struct {
	DefaultModel `bson:",inline"`

	// Guild ID
	GuildId string `json:"guild_id" bson:"guild_id"`

	// Channel ID
	ChannelId string `json:"channel_id" bson:"channel_id"`

	// The ID of the starboard
	StarboardId primitive.ObjectID `json:"starboard_id" bson:"starboard_id"`

	// The ID of the message on the starboard channel
	StarboardMsgId string `json:"starboard_msg_id" bson:"starboard_msg_id"`

	// Users who reacted to the message
	Users []string `json:"users" bson:"users"`
}

func (smsg *StarboardMessage) CollectionName() string {
	return "starboard_messages"
}

func StarboardMessageColl() *mgm.Collection {
	return mgm.Coll(&StarboardMessage{})
}
