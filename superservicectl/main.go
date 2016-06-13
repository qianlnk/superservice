package main

import (
	"fmt"

	"github.com/qianlnk/superservice/superservicectl/command"
	"github.com/qianlnk/terminal"
)

func main() {
	superterm := terminal.NewTerminal("> ")
	superterm.SetSystemCommand(command.CommandList)
	for {
		fmt.Printf("> ")
		cmd := superterm.GetCommand()
		fmt.Println()
		if cmd == "" {
			continue
		}
		if cmd == "exit" {
			break
		}
		command.DealCommand(cmd)
	}
}
