package converter

import (
	"encoding/json"
	"fmt"
	"github.com/notaneet/rasp51/model"
	"io/ioutil"
)

type JSONConverter struct {
	Pretty bool
}

func (j JSONConverter) Write(node model.TimetableNode, out string) error {
	if out == "" {
		return fmt.Errorf("-out can not be empty")
	}

	var ret []byte
	var err error
	if j.Pretty {
		ret, err = json.MarshalIndent(node, "", "  ")
	} else {
		ret, err = json.Marshal(node)
	}
	if err != nil {
		return err
	}

	return ioutil.WriteFile(out, ret, 0644)
}
