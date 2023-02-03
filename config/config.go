package config

import (
	"os"
	"strings"

	"github.com/disgoorg/disgo/gateway"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

var (
	MongoUri     string
	Token        string
	Intents      gateway.ConfigOpt = gateway.WithIntents(
		gateway.IntentGuilds,
		gateway.IntentGuildEmojisAndStickers,
		gateway.IntentGuildMessages,
		gateway.IntentGuildMessageReactions,
		gateway.IntentDirectMessages,
		gateway.IntentMessageContent,
	)
	DevServersId       = []string{}
	DevId        string
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Panic().Msgf("An error ocurred when reading .env: %v", err)
	}

	if os.Getenv("APP_ENV") == "development" {
		Token = os.Getenv("TEST_BOT_TOKEN")
	} else {
		Token = os.Getenv("BOT_TOKEN")
	}

	MongoUri = os.Getenv("MONGO_URI")
	DevServersId = strings.Split(os.Getenv("DEV_SERVER_ID"), ",")
	DevId = os.Getenv("DEV_ID")
}
