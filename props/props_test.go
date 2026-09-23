package props

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempProps(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.prop")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp props file: %v", err)
	}
	return path
}

func TestView_StripsDoubleQuotes(t *testing.T) {
	path := writeTempProps(t, `token="abc123"`)
	props, err := View(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := props["token"]; got != "abc123" {
		t.Errorf("expected token=abc123, got %q", got)
	}
}

func TestView_TrimsWhitespace(t *testing.T) {
	path := writeTempProps(t, `token = abc123`)
	props, err := View(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := props["token"]; got != "abc123" {
		t.Errorf("expected token=abc123, got %q", got)
	}
}

func TestView_SkipsBlankAndCommentLines(t *testing.T) {
	path := writeTempProps(t, "\n# this is a comment\ntoken=abc123\n\n# another comment\n")
	props, err := View(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(props) != 1 {
		t.Fatalf("expected 1 prop, got %d: %v", len(props), props)
	}
	if got := props["token"]; got != "abc123" {
		t.Errorf("expected token=abc123, got %q", got)
	}
}

func TestView_UnquotedKeyValue(t *testing.T) {
	path := writeTempProps(t, `token=abc123`)
	props, err := View(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := props["token"]; got != "abc123" {
		t.Errorf("expected token=abc123, got %q", got)
	}
}
