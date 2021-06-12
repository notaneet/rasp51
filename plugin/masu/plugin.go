package masu

import (
	"github.com/notaneet/rasp51/config"
	"github.com/notaneet/rasp51/model"
)

type _MASUMurmanskPlugin struct {
	timetable model.TimetableNode
	config    config.ParserConfig
}

func GetPlugin(cfg config.ParserConfig) *_MASUMurmanskPlugin {
	t := model.TimetableNode{}

	return &_MASUMurmanskPlugin{timetable: t, config: cfg}
}

func (p _MASUMurmanskPlugin) GetInstitution() string {
	return "МАГУ"
}

func (p *_MASUMurmanskPlugin) GetTimetable() (error, model.TimetableNode) {
	if err := p.scrap(); err != nil {
		return err, nil
	}

	return nil, p.timetable
}
