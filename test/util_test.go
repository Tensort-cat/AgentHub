package test

import (
	"path/filepath"
	"testing"
)

func TestFilepath(t *testing.T) {
	ext := filepath.Ext("fuck.md")
	t.Log(ext)
}
