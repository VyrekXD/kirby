package utils

import (
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/snowflake/v2"
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

func ParseEmoji(emoji string, gId snowflake.ID, c bot.Client) (string, error) {
	id, err := snowflake.Parse(emoji)
	if err != nil {
		return emoji, nil
	} else {
		data, err := c.Rest().GetEmoji(id, id)
		if err != nil {
			return "", err
		}

		return data.Mention(), nil
	}
}

type MapCallback[T any] func(v T) T

func Map[T any](arr []T, callback MapCallback[T]) []T {
	arr = make([]T, len(arr))

	for i, v := range arr {
		arr[i] = callback(v)
	}

	return arr
}
