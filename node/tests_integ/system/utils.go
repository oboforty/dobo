package tests_system

import (
	"bytes"
	"os"
	"testing"
)

func CapturePrint(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Close the writer and restore os.Stdout
	t.Cleanup(func() {
		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)

		t.Log(buf.String())
	})
}
