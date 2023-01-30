package langs

import (
	"fmt"
	"os"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/rs/zerolog/log"
)

var _packs = map[string]*gabs.Container{}

type CommandPack struct {
	langPack  LangPack
	commmands []string
}

func (cp *CommandPack) SubCommand(c string) *CommandPack {
	cp.commmands = append(cp.commmands, c)

	return cp
}

func (cp *CommandPack) Getf(k string, a ...any) *string {
	t := cp.Get(k)
	p := fmt.Sprintf(*t, a...)

	return &p
}

func (cp *CommandPack) Get(k string) *string {
	d := []string{}
	d = append(d, cp.commmands...)
	d = append(d, k)

	t, ok := cp.langPack._container.S(d...).Data().(string)
	if !ok {
		return &cp.langPack.NotFoundText
	}

	return &t
}

type LangPack struct {
	_container   *gabs.Container
	NotFoundText string
}

func (p *LangPack) Command(c string) *CommandPack {
	return &CommandPack{langPack: *p, commmands: []string{c}}
}

func Pack(c string) *LangPack {
	pack := _packs[c]
	if pack == nil {
		log.Panic().Msgf(`Cannot find "%v" lang pack`, c)
	}

	notFound, ok := pack.S("notFound").Data().(string)
	if !ok {
		log.Panic().Msgf(`Cannot find "notFound" text in lang pack "%v"`, c)
	}

	return &LangPack{_container: pack, NotFoundText: notFound}

}

func Load() error {
	files, err := os.ReadDir("./langs/packs")
	if err != nil {
		log.Panic().Err(err).Msg("Error loading lang packs: ")
	}

	for _, file := range files {
		data, err := os.ReadFile("./langs/packs/" + file.Name())
		if err != nil {
			return fmt.Errorf(`error loading lang pack "%v": %v`, file.Name(), err)
		}

		parsed, err := gabs.ParseJSON(data)
		if err != nil {
			return fmt.Errorf(`error parsing lang pack "%v" json: %v`, file.Name(), err)
		}

		_packs[strings.Replace(file.Name(), ".json", "", 1)] = parsed
	}

	return nil
}
