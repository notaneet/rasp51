package plugin

import "github.com/notaneet/rasp51/model"

type Plugin interface {
	GetInstitution() string
	GetTimetable() (error, model.TimetableNode)
}
