package models

import (
	"github.com/kamva/mgm/v3"
)

type GuildConfig struct {
	DefaultModel `bson:",inline"`

	Lang string `json:"lang" bson:"lang"`
}

func (gconfig *GuildConfig) CollectionName() string {
	return "guilds_config"
}

func GuildConfigColl() *mgm.Collection {
	return mgm.Coll(&GuildConfig{})
}
