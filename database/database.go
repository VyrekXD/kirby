package database

import (
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/vyrekxd/kirby/config"
)

var (
	db *mgm.Config
)

func Connect() error {
	err := mgm.SetDefaultConfig(nil, "kirby", options.Client().ApplyURI(config.MongoUri))

	return err
}
