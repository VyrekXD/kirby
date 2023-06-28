package utils

import (
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/snowflake/v2"
	"github.com/forPelevin/gomoji"
)

func WaitDo(timeout time.Duration, do func()) {
	go func() {
		time.Sleep(timeout)

		do()
	}()
}

func ReadableResult(
	r *interface{},
	nonNilValue string,
	nilValue string,
) string {
	if *r == nil {
		return nilValue
	} else {
		return nonNilValue
	}
}

func ReadableBool(b *bool, trueValue string, falseValue string) string {
	if *b {
		return trueValue
	} else {
		return falseValue
	}
}

func ParseEmoji(
	s bot.Client,
	guildId snowflake.ID,
	emoji string,
) (string, error) {
	if res := gomoji.FindAll(emoji); res != nil {
		return emoji, nil
	} else {
		id, err := snowflake.Parse(emoji)
		if err != nil {
			return "", err
		}

		e, err := s.Rest().GetEmoji(guildId, id)
		if err != nil {
			return "", err
		}

		return e.String(), err
	}
}
