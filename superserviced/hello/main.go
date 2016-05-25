package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	i := 0
	arg := os.Args
	fmt.Println(arg)
	for {
		fmt.Println(arg[1], i)
		i++
		time.Sleep(1 * time.Second)
	}
}
