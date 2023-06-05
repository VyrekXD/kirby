package cmd_util

import "errors"

var (
	errResolved         = errors.New("resolved")
	errNoChannel        = errors.New("noChannel")
	errNoValidEmoji     = errors.New("noValidEmoji")
	errTimeout          = errors.New("timeout")
	errNoValidNumber    = errors.New("noValidNumber")
	errNoValidName      = errors.New("noValidName")
	errNoValidHex       = errors.New("noValidHex")
	errNoValidParsedHex = errors.New("noValidParsedHex")
)

var (
	collBaseId    = "starboard:"
	collChannelId = collBaseId + "channel"
	collSkipId    = collBaseId + "skip"
	collYesId     = collBaseId + "yes"
	collNoId      = collBaseId + "no"
)
