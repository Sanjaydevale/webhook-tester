package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleCmdArgs(t *testing.T) {
	t.Run("port number is configurable", func(t *testing.T) {
		got, _ := handleCmdArgs([]string{"-p", "8080"})
		want := 8080
		if got.port != want {
			t.Errorf("invalid port config, got %d want %d", got.port, want)
		}
	})

	t.Run("fields are configurable", func(t *testing.T) {
		got, _ := handleCmdArgs([]string{"Method", "Body"})
		want := []string{"Method", "Body"}
		assert.ElementsMatch(t, got.fields, want)
	})
}
