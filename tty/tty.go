package tty

import (
	"context"
	"fmt"
	cpty "github.com/creack/pty"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	defaultCloseSignal  = syscall.SIGKILL
	defaultCloseTimeout = 10 * time.Second
)

type window struct {
	row uint16
	col uint16
	x   uint16
	y   uint16
}

type TTY struct {
	bin  string
	argv []string

	cmd *exec.Cmd
	pty *os.File

	cancelCtx  context.Context
	cancelFunc context.CancelFunc
}

func New(command string, optfs ...OptionFunc) (*TTY, error) {
	opt := newOption(optfs...)

	ctx, cancel := context.WithCancel(context.Background())
	c := &TTY{
		cancelCtx:  ctx,
		cancelFunc: cancel,
	}

	if args := strings.Split(command, " "); len(args) > 1 {
		c.bin = args[0]
		c.argv = args[1:]
	} else {
		c.bin = args[0]
	}

	c.cmd = exec.CommandContext(opt.ctx, c.bin, c.argv...)

	env := []string{"TERM=xterm", "LANG=en_US.UTF-8", "LC_ALL=en_US.UTF-8", "LANGUAGE=en_US.UTF-8"}

	if opt.useCurrentEnv {
		env = append(env, os.Environ()...)
	}

	env = append(env, opt.extraEnv...)
	c.cmd.Env = env

	if opt.workdir != nil {
		c.cmd.Dir = *opt.workdir
	}

	pty, err := cpty.Start(c.cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	c.pty = pty
	c.waitProcess()

	return c, nil
}

func (c *TTY) waitProcess() {
	go func() {
		defer func() {
			c.pty.Close()
			c.cancelFunc()
		}()

		c.cmd.Wait()
	}()
}

// GetPID returns the process ID of the tty.
func (c *TTY) GetPID() int {
	return c.cmd.Process.Pid
}

// Read reads data from the tty's stdout.
func (c *TTY) Read(p []byte) (n int, err error) {
	return c.pty.Read(p)
}

// Write writes data to the tty's stdin.
func (c *TTY) Write(p []byte) (n int, err error) {
	return c.pty.Write(p)
}

// Close sends the close signal to the tty and waits for it to close.
func (c *TTY) Close() error {
	if c.cmd.Process != nil {
		c.cmd.Process.Signal(syscall.SIGKILL)
	}

	select {
	case <-c.cancelCtx.Done():
		return nil
	}

	return nil
}

func (c *TTY) ResizeWindow(width int, height int) error {
	w := window{
		uint16(height),
		uint16(width),
		0,
		0,
	}

	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, c.pty.Fd(), syscall.TIOCSWINSZ, uintptr(unsafe.Pointer(&w))); err != 0 {
		return err
	}

	return nil
}

// GetWindowSize returns the size of the tty's window.
func (c *TTY) GetWindowSize() (int, int, error) {
	w := window{}

	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, c.pty.Fd(), syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&w))); err != 0 {
		return 0, 0, err
	}

	return int(w.col), int(w.row), nil
}

func (c *TTY) Done() <-chan struct{} {
	return c.cancelCtx.Done()
}
