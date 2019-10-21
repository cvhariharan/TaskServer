package task

import (
	"testing"
	"os"
	"time"
	"fmt"
)


func TestInit(t *testing.T) {
	c := new(CommandTask)
	err := c.Init("ls -al")
	if err != nil {
		t.Error(err)
	}
	if c.GetStatus() != INIT {
		t.Error("status not set to initiated")
	}
}

func TestRun(t *testing.T) {
	c := new(CommandTask)
	err := c.Init("ls -al")
	if err != nil {
		t.Error(err)
	}
	err = c.Run()
	if err != nil {
		t.Error(err)
	}
}

func TestGetOutput(t *testing.T) {
	c := new(CommandTask)
	err := c.Init("ls -alh")
	if err != nil {
		t.Error(err)
	}
	c.SetOutput(os.Stdout)
	err = c.Run()
	if err != nil {
		t.Error(err)
	}
}

func TestSetInput(t *testing.T) {
	c := new(CommandTask)
	err := c.Init("cp /dev/stdin test.txt")
	if err != nil {
		t.Error(err)
	}
	f, err := os.Open("commandtask_test.go")
	if err != nil {
		t.Error(err)
	}
	c.SetInput(f)
	err = c.Run()
	if err != nil {
		t.Error(err)
	}
}

func TestPause(t *testing.T) {
	c := new(CommandTask)
	err := c.Init("python test/loop.py")
	if err != nil {
		t.Error(err)
	}
	go func() {
		err = c.Run()
		if err != nil {
			fmt.Println(err)
			// t.Error(err)
		}
	}()
	time.Sleep(100 * time.Millisecond)
	err = c.Pause()
	if err != nil {
		fmt.Println(err)
	}
	if c.status != PAUSE {
		t.Error("status not set to pause")
	}
}

func TestResume(t *testing.T) {
	c := new(CommandTask)
	err := c.Init("python test/loop.py")
	if err != nil {
		t.Error(err)
	}
	go func() {
		err = c.Run()
		if err != nil {
			fmt.Println(err)
			// t.Error(err)
		}
	}()
	time.Sleep(100 * time.Millisecond)
	err = c.Pause()
	if err != nil {
		fmt.Println(err)
	}
	if c.status != PAUSE {
		t.Error("status not set to pause")
	}
	c.Resume()
	if c.status != RUNNING {
		t.Error("status not set to running")
	}
}

func TestKill(t *testing.T) {
	c := new(CommandTask)
	err := c.Init("python test/loop.py")
	if err != nil {
		t.Error(err)
	}
	go func() {
		err = c.Run()
		if err != nil {
			fmt.Println(err)
			// t.Error(err)
		}
	}()
	time.Sleep(100 * time.Millisecond)
	err = c.Kill()
	if err != nil {
		fmt.Println(err)
	}
	if c.status != FATAL {
		t.Error("status not set to fatal")
	}
}

