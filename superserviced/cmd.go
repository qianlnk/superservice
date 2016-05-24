package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Cmd struct {
	Name    string
	Command string
	User    string
}

func (c *Cmd) start() error {
	if ServiceList[c.Name].Status == RUNNING {
		return errors.New(fmt.Sprintf("%s already running.", c.Name))
	}
	cmd := exec.Command(c.Command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return err
	}
	ServiceList[c.Name].Running()
	err = cmd.Wait()
	if err != nil {
		return err
	}
	ServiceList[c.Name].Stop()
	return nil
}

func (c *Cmd) Stop() error {
	cmd := exec.Command("kill", "-9", fmt.Sprintf("%d", ServiceList[c.Name].Pid))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil && strings.Contains(err.Error(), "killed") == false {
		return err
	}
	ServiceList[c.Name].Stop()
	return nil
}

func (c *Cmd) Restart() error {
	err := c.Stop()
	if err != nil {
		return err
	}
	return c.start()
}
