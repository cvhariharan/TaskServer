package task

import (
	"os/exec"
	"io"
	"strings"
	"errors"
	"syscall"
	"github.com/rs/xid"
)

const (
	INIT = "initiated"
	RUNNING = "running"
	DONE = "done"
	PAUSE = "pause"
	FATAL = "fatal"
	KILL_ACTION = "kill"
	PAUSE_ACTION = "stop"
	RESUME_ACTION = "resume"
)

// An abstraction over linux process
type Task interface {
	Run() error
	Kill() error
	Pause() error
	Resume() error
	GetStatus() string
	GetPID() int
}

type CommandTask struct {
	status string
	cmd *exec.Cmd
}

// Returns a random string to be used a identifier
func(c *CommandTask) Init(command string) string {
	guid := xid.New()
	args := strings.Split(command, " ")
	c.cmd = exec.Command(args[0], args[1:]...)
	c.status = INIT
	return guid.String()
}

// Sets the stdout and stderr of the process to the supplied writer
func(c *CommandTask) SetOutput(w io.Writer) error {
	c.cmd.Stdout = w
	c.cmd.Stderr = w
	return nil
}

// Sets the stdin of the process to the supplied reader
func(c *CommandTask) SetInput(r io.Reader) error {
	c.cmd.Stdin = r
	return nil
}

func(c *CommandTask) Run() error {
	err := c.cmd.Start()
	if err != nil {
		return err
	}
	c.status = RUNNING
	err = c.cmd.Wait()
	if err != nil {
		return err
	}
	c.status = DONE
	return nil
}

func(c *CommandTask) Kill() error {
	if c.status == RUNNING {
		err := c.cmd.Process.Kill()
		if err != nil {
			return err
		}
		c.status = FATAL
	} else {
		return errors.New("Can't kill a stopped process")
	}
	return nil
}

func(c *CommandTask) Pause() error {
	err := c.cmd.Process.Signal(syscall.SIGSTOP)
	if err != nil {
		return err
	}
	c.status = PAUSE
	return nil
}

// Sends the continue signal to a paused process
// Platform Dependent
func(c *CommandTask) Resume() error {
	err := c.cmd.Process.Signal(syscall.SIGCONT)
	if err != nil {
		return err
	}
	c.status = RUNNING
	return nil
}

// Returns the status as the consts defined earlier
func(c *CommandTask) GetStatus() string {
	return c.status
}

func(c *CommandTask) GetPID() int {
	return c.cmd.Process.Pid
}