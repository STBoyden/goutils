package fsutils

import "testing"

func TestPathExists(t *testing.T) {
	if !PathExists("fsutils.go") {
		t.Error("fsutils.go should exist")
	}
}
