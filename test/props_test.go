package main

import (
	"fmt"
	"github.com/nhomble/groupme.go/props"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
)

func TestView(t *testing.T) {
	data := []byte("a=b")
	n := fmt.Sprintf("/tmp/.props%f", rand.Float64())
	defer os.Remove(n)
	ioutil.WriteFile(n, data, 0644)
	config, err := props.View(n)
	if err != nil {
		t.Error(err)
	}
	if val, ok := (*config)["a"]; ok {
		if "b" != val {
			t.Errorf("Should have found 'b', but config['a'] = %v", val)
		}
	} else {
		t.Error("a is not present in the map")
	}
}
