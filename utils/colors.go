package utils

import (
	"errors"
	"strconv"
	"strings"

	"github.com/vyrekxd/kirby/constants"
)

func ToString(h int) string {
	return strconv.FormatInt(int64(h), 16)
}

func ToHex(c string) (int, error) {
	if constants.HexColorRegex.FindString(c) == "" {
		return 0, errors.New("no valid hex")
	}
	if c[0:1] == "#" {
		c = FormatHex(c)
	}
	if IsShortHex(c) {
		c = ShortToLongHex(c)
	}

	i, err := strconv.ParseInt(c, 16, 64)
	if err != nil {
		return 0, err
	}

	return int(i), nil
}

func FormatHex(h string) string {
	return strings.ToLower(strings.Replace(h, "#", "", 1))
}

func ShortToLongHex(h string) string {
	end := ""

	for _, d := range strings.Split(FormatHex(h), "") {
		end += strings.Repeat(d, 2)
	}

	return end
}

func IsShortHex(h string) bool {
	return len(h) == 3
}
