package converter

import "github.com/notaneet/rasp51/model"

type IConverter interface {
	Write(node model.TimetableNode, out string) error
}
