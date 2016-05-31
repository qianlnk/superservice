package main

import (
	"fmt"

	"qianno.xie/superservice/superservicectl/command"
	"qianno.xie/superservice/terminal"
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
