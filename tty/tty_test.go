package tty

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"
)

// createScript creates a temporary script file with the given content.
func createScript(content string) (string, func()) {
	file, err := os.CreateTemp("", "script")
	if err != nil {
		panic(err)
	}

	if _, err := file.WriteString(content); err != nil {
		panic(err)
	}

	if err := file.Close(); err != nil {
		panic(err)
	}

	if err := os.Chmod(file.Name(), 0755); err != nil {
		panic(err)
	}

	return file.Name(), func() {
		_ = os.Remove(file.Name())
	}
}

// processExists checks if a process with the given PID exists.
func processExists(pid int) (bool, error) {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, err
	}

	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return false, nil
	}

	return true, nil
}
func TestCommand_Read(t *testing.T) {
	t.Run("test Read()", func(t *testing.T) {
		cmd, err := New(`echo -n hello world`)
		assert.NoError(t, err)

		buff := make([]byte, 1024)
		n, err := cmd.Read(buff)
		assert.NoError(t, err)
		assert.Equal(t, 11, n)
		assert.Equal(t, "hello world", string(buff[:n]))
	})
}

func TestCommand_Write(t *testing.T) {
	t.Run("test Write()", func(t *testing.T) {
		cmd, err := New("cat")
		assert.NoError(t, err)

		n, err := cmd.Write([]byte("hello world"))
		assert.NoError(t, err)
		assert.Equal(t, 11, n)

		buff := make([]byte, 1024)
		n, err = cmd.Read(buff)
		assert.NoError(t, err)
		assert.Equal(t, 11, n)
		assert.Equal(t, "hello world", string(buff[:n]))
	})
}

func TestCommand_Close(t *testing.T) {
	t.Run("test Close()", func(t *testing.T) {
		cmd, err := New(`cat`)
		assert.NoError(t, err)

		pid := cmd.GetPID()

		exists, err := processExists(pid)
		assert.NoError(t, err)
		assert.True(t, exists)
		err = cmd.Close()
		assert.NoError(t, err)

		// Check if the process is closed
		exists, err = processExists(pid)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestCommand_ResizeWindow(t *testing.T) {
	t.Run("test ResizeWindow()/GetWindowSize()", func(t *testing.T) {
		cmd, err := New(`cat`)
		assert.NoError(t, err)

		err = cmd.ResizeWindow(80, 24)
		assert.NoError(t, err)

		width, height, err := cmd.GetWindowSize()
		assert.NoError(t, err)
		assert.Equal(t, 80, width)
		assert.Equal(t, 24, height)
	})
}

func TestCommand_WithCtx(t *testing.T) {
	t.Run("test WithCtx()", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cmd, err := New(`cat`, WithContext(ctx))
		assert.NoError(t, err)

		cancel()

		time.Sleep(100 * time.Millisecond)
		exist, err := processExists(cmd.GetPID())
		assert.NoError(t, err)
		assert.False(t, exist)
	})
}

func TestCommand_WithWorkdir(t *testing.T) {
	t.Run("test WithWorkDir()", func(t *testing.T) {
		cmd, err := New(`pwd`, WithWorkdir("/"))
		assert.NoError(t, err)

		buff := make([]byte, 1024)
		n, err := cmd.Read(buff)
		res := string(buff[:n])
		res = strings.Trim(res, "\n\r ")
		assert.NoError(t, err)
		assert.Equal(t, "/", res)
	})
}

func TestCommand_WithExtraEnv(t *testing.T) {
	t.Run("test WithExtraEnv()", func(t *testing.T) {
		cmd, err := New(`env`, WithExtraEnv("HELLO=WORLD", "FOO=BAR"))
		assert.NoError(t, err)

		buff, err := io.ReadAll(cmd)
		assert.NoError(t, err)
		assert.Contains(t, string(buff), "HELLO=WORLD")
		assert.Contains(t, string(buff), "FOO=BAR")
	})
}

func TestCommand_WithEmptyEnv(t *testing.T) {
	t.Run("test WithEmptyEnv()", func(t *testing.T) {
		os.Setenv("HELLO", "WORLD")

		cmd, err := New(`env`, WithEmptyEnv())
		assert.NoError(t, err)

		buff, err := io.ReadAll(cmd)
		assert.NoError(t, err)
		assert.NotContains(t, string(buff), "HELLO=WORLD")
	})
}
