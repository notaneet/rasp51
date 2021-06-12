package converter

import "github.com/notaneet/rasp51/model"

type DummyConverter struct {
}

func (d DummyConverter) Write(model.TimetableNode, string) error {
	return nil
}
