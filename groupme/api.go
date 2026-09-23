package groupme

import (
	"encoding/json"
	"fmt"
	"strings"
)

const DEFAULT_MESSAGE_LIMIT = 20

// snippet returns a truncated, human-readable representation of the raw
// response body to help with debugging malformed API responses.
func snippet(data *[]byte) string {
	const maxLen = 200
	s := string(*data)
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

func unravel(data *[]byte, dest interface{}) error {
	var obj map[string]*json.RawMessage
	err := json.Unmarshal(*data, &obj)
	if err != nil {
		return err
	}

	raw, ok := obj["response"]
	if !ok || raw == nil {
		return fmt.Errorf("groupme: missing or null \"response\" field in body: %s", snippet(data))
	}

	err = json.Unmarshal(*raw, &dest)
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

	metaRaw, ok := obj["meta"]
	if !ok || metaRaw == nil {
		return fmt.Sprintf("groupme: missing or null \"meta\" field in body: %s", snippet(data))
	}

	var metaObj map[string]*json.RawMessage
	err = json.Unmarshal(*metaRaw, &metaObj)
	if err != nil {
		return err.Error()
	}

	errorsRaw, ok := metaObj["errors"]
	if !ok || errorsRaw == nil {
		return fmt.Sprintf("groupme: missing or null \"errors\" field in body: %s", snippet(data))
	}

	var errors []string
	err = json.Unmarshal(*errorsRaw, &errors)
	if err != nil {
		return err.Error()
	}
	return strings.Join(errors, "\n")
}
