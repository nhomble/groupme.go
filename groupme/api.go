package groupme

import (
	"encoding/json"
	"strings"
)

const DEFAULT_MESSAGE_LIMIT = 20

func unravel(data *[]byte, dest interface{}) error {
	var obj map[string]*json.RawMessage
	err := json.Unmarshal(*data, &obj)
	if err != nil {
		return err
	}
	err = json.Unmarshal(*obj["response"], &dest)
	if err != nil {
		return err
	}
	return nil
}

func parseError(data *[]byte) string {
	var obj map[string]*json.RawMessage
	err := json.Unmarshal(*data, &obj)
	if err != nil {
		return err.Error()
	}
	err = json.Unmarshal(*obj["meta"], &obj)
	if err != nil {
		return err.Error()
	}

	var errors []string
	err = json.Unmarshal(*obj["errors"], &errors)
	if err != nil {
		return err.Error()
	}
	return strings.Join(errors, "\n")
}
