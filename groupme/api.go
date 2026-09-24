package groupme

import (
	"encoding/json"
	"fmt"
	"strings"
)

const DefaultMessageLimit = 20

// validID rejects id values that would turn a path-segment substitution into
// a different endpoint: empty, ".", and ".." all pass through
// url.PathEscape unchanged but resolve to a different URL path.
func validID(id string) error {
	if id == "" || id == "." || id == ".." {
		return fmt.Errorf("groupme: invalid id %q", id)
	}
	return nil
}

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

// APIError is returned for any non-2xx HTTP response. It carries the HTTP
// status code and, when the response body could be parsed, the list of
// error messages reported by the GroupMe API.
type APIError struct {
	StatusCode int
	Errors     []string
	Method     string
	URL        string
}

func (e *APIError) Error() string {
	if len(e.Errors) > 0 {
		return fmt.Sprintf("groupme: %s %s failed with status=%d: %s", e.Method, e.URL, e.StatusCode, strings.Join(e.Errors, "; "))
	}
	return fmt.Sprintf("groupme: %s %s failed with status=%d", e.Method, e.URL, e.StatusCode)
}

// parseErrorList extracts the meta.errors list from a GroupMe API error
// response body. It reports ok=false if the body could not be parsed or
// did not contain the expected fields.
func parseErrorList(data *[]byte) ([]string, bool) {
	var obj map[string]*json.RawMessage
	if err := json.Unmarshal(*data, &obj); err != nil {
		return nil, false
	}

	metaRaw, ok := obj["meta"]
	if !ok || metaRaw == nil {
		return nil, false
	}

	var metaObj map[string]*json.RawMessage
	if err := json.Unmarshal(*metaRaw, &metaObj); err != nil {
		return nil, false
	}

	errorsRaw, ok := metaObj["errors"]
	if !ok || errorsRaw == nil {
		return nil, false
	}

	var errs []string
	if err := json.Unmarshal(*errorsRaw, &errs); err != nil {
		return nil, false
	}
	return errs, true
}

// newAPIError builds an APIError from a non-2xx HTTP response, populating
// the parsed error messages when the body can be parsed, and falling back
// to a generic message (with StatusCode still set) otherwise.
func newAPIError(method, url string, statusCode int, data []byte) *APIError {
	errs, ok := parseErrorList(&data)
	if !ok {
		return &APIError{StatusCode: statusCode, Method: method, URL: url}
	}
	return &APIError{StatusCode: statusCode, Errors: errs, Method: method, URL: url}
}
