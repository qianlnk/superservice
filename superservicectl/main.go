package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"qianno.xie/superservice/superservicectl/command"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("> ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.Trim(cmd, "\n")

		if cmd == "" {
			continue
		}
		command.DealCommand(cmd)
	}
}
