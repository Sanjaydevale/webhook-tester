package main

import (
	"fmt"
	"io"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerCmdArgs(t *testing.T) {
	t.Run("server port and domain is configurable", func(t *testing.T) {
		argsStub := []string{"-p", "8888", "-d", "test"}
		got, _ := handleCmdArgs(argsStub)
		assert.Equal(t, 8888, got.port)
		assert.Equal(t, "test", got.domain)
	})
}

func TestServer(t *testing.T) {
	t.Run("test graceful shutdown", func(t *testing.T) {
		binName := makeBinary(t)
		runCmd := exec.Command(binName, "-p", "8080", "-d", "localhost:8080")
		out, _ := runCmd.StdoutPipe()
		runCmd.Start()
		time.Sleep(time.Second)
		runCmd.Process.Signal(syscall.SIGINT)
		got, _ := io.ReadAll(out)
		runCmd.Wait()

		assert.Contains(t, string(got), "Shutting Down Server")
	})
}

func makeBinary(tb testing.TB) string {
	tb.Helper()
	binName := fmt.Sprintf("/tmp/server%d", time.Now().Unix())
	buildCmd := exec.Command("go", "build", "-o", binName, "./main.go")
	err := buildCmd.Run()
	require.NoError(tb, err)
	return binName
}
