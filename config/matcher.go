package config

import (
	"github.com/notaneet/rasp51/utils"
	"regexp"
	"strings"
)

type Matcher struct {
	MatchRaw utils.StringEnum
	Regexp   *regexp.Regexp
}

func (m *Matcher) Match(text string) bool {
	if len(m.MatchRaw) == 0 {
		return true
	}

	for _, s := range m.MatchRaw {
		if s == text {
			return true
		} else if strings.HasPrefix(s, "~") && regexp.MustCompile(s[1:]).MatchString(text) {
			return true
		}
	}

	return false
}
