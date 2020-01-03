package props

import (
	"bufio"
	"os"
	"strings"
)

type GroupmeProps map[string]string

var defaultName string = ".groupme.go.properties"

// Get name to configuration file - assumes linux
// TODO consider other compatibilities with filesystem
func DefaultName() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return home + "/" + defaultName
}

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
		text := scanner.Text()
		parts := strings.Split(text, "=")
		props[parts[0]] = strings.Join(parts[1:], "=")
	}
	return &props, nil
}
