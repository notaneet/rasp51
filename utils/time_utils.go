package utils

import (
	"strings"
	"time"
)

func AddTimeToDate(date time.Time, stringTime string) time.Time {
	timeRaw := strings.Split(strings.TrimSpace(stringTime), ".")
	return time.Date(date.Year(), date.Month(), date.Day(), MustAtoi(timeRaw[0]), MustAtoi(timeRaw[1]), date.Second(), 0, time.Local)
}

func GetInterval(str string) (f *time.Time, s *time.Time) {
	if str != "" {
		spl := strings.Split(str, "-")
		if spl[0] != "" {
			fS, _ := time.Parse("02.01.2006", spl[0])
			f = &fS
		}
		if len(spl) > 1 && spl[1] != "" {
			sS, _ := time.Parse("02.01.2006", spl[1])
			s = &sS
		}
	}
	return
}
