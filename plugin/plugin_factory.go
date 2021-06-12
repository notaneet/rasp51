package plugin

import (
	"github.com/notaneet/rasp51/config"
	"github.com/notaneet/rasp51/plugin/masu"
)

func NewPlugin(name string, cfg config.ParserConfig) Plugin {
	switch name {
	case "МАГУ":
		return masu.GetPlugin(cfg)
	default:
		return nil
	}
}
