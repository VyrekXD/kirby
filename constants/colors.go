package constants

import (
	"math/rand"
	"regexp"
)

var HexColorRegex = regexp.MustCompile("^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$")

type colors struct {
	Main  int
	Error int
	Good  int
	Info  int
}

var Colors = colors{
	Main:  0xfe9cb1,
	Error: 0xf42c2c,
	Good:  0x2cd649,
	Info:  0xefe92d,
}

func Random() int {
	return rand.Intn(0xffffff + 1)
}
