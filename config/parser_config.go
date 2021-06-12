package config

import (
	"github.com/notaneet/rasp51/utils"
	"time"
)

type ParserConfig struct {
	GroupMatcher    Matcher
	FacultyMatcher  Matcher
	LecturerMatcher Matcher
	ClassMatcher    Matcher
	CampusMatcher   Matcher

	Interval string

	//Ленивая загрузка
	StartTime *time.Time
	EndTime   *time.Time
}

func (cfg *ParserConfig) Init() {
	if cfg.Interval != "" && cfg.StartTime == nil && cfg.EndTime == nil {
		cfg.StartTime, cfg.EndTime = utils.GetInterval(cfg.Interval)
	}
}
