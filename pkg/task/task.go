package task

import (
	"os/exec"
	"io"
	"strings"
	"errors"
	"syscall"
)

const (
	INIT = "initiated"
	RUNNING = "running"
	DONE = "done"
	PAUSE = "pause"
	FATAL = "fatal"
)

// Models a single task
type Task interface {
	Init(string) error
	GetOutput() (io.ReadCloser, error)
	GetInput() (io.WriteCloser, error)
	Run() error
	Kill() error
	Pause() error
	Resume() error
	GetStatus() string
}

type CommandTask struct {
	status string
	cmd *exec.Cmd
}

func(c *CommandTask) Init(command string) error {
	args := strings.Split(command, " ")
	c.cmd = exec.Command(args[0], args[1:]...)
	c.status = INIT
	return nil
}

func(c *CommandTask) SetOutput(w io.Writer) error {
	c.cmd.Stdout = w
	c.cmd.Stderr = w
	return nil
}

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

func(c *CommandTask) Resume() error {
	err := c.cmd.Process.Signal(syscall.SIGCONT)
	if err != nil {
		return err
	}
	c.status = RUNNING
	return nil
}


func(c *CommandTask) GetStatus() string {
	return c.status
}