package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleCmdArgs(t *testing.T) {
	t.Run("port number is configurable", func(t *testing.T) {
		got, _ := handleCmdArgs([]string{"-p", "8888"})
		var gotPorts []int
		for _, port := range got.ports {
			gotPorts = append(gotPorts, port)
		}
		want := []int{8888}
		assert.Equal(t, gotPorts, want, "invalid port config")
	})

	t.Run("fields are configurable", func(t *testing.T) {
		got, err := handleCmdArgs([]string{"-p", "8080", "Method", "Body"})
		require.NoError(t, err, "handling cmd args")
		want := []string{"Method", "Body"}
		assert.ElementsMatch(t, got.fields, want)
	})
}

func ExampleFields() {
	handleFieldArgs([]string{"test"})
	// output:
	// does not contain filed  test
	// available fields :
	// Body
	// Close
	// ContentLength
	// Header
	// Host
	// Method
	// Proto
	// ProtoMajor
	// ProtoMinor
	// RemoteAddr
	// RequestURI
	// TransferEncoding
	// URL
}
