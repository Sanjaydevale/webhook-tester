package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerCmdArgs(t *testing.T) {
	t.Run("server port is configurable", func(t *testing.T) {
		argsStub := []string{"-p", "8888"}
		got, _ := handleCmdArgs(argsStub)
		assert.Equal(t, 8888, got.port)
	})

	t.Run("domain is configurable", func(t *testing.T) {
		argsStub := []string{"-d", "test"}
		got, _ := handleCmdArgs(argsStub)
		assert.Equal(t, "test", got.domain)
	})
}
