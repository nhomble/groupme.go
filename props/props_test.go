package props

import (
	"io/ioutil"
	"path"
	"testing"
)

func TestView(t *testing.T) {
	data := []byte("a=b")
	n := path.Join(t.TempDir(), "props")
	ioutil.WriteFile(n, data, 0644)
	config, err := View(n)
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
