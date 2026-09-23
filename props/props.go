package props

import (
	"bufio"
	"os"
	"strings"
)

type GroupmeProps map[string]string

// View properties in map
func View(propLocation string) (*GroupmeProps, error) {
	f, err := os.Open(propLocation)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	props := make(GroupmeProps)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		parts := strings.SplitN(text, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := unquote(strings.TrimSpace(parts[1]))
		props[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &props, nil
}

// unquote strips a single layer of surrounding matching quotes (' or ") from
// a value, if present.
func unquote(value string) string {
	if len(value) >= 2 {
		first, last := value[0], value[len(value)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}
