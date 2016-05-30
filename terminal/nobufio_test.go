package input

import (
	"fmt"
	"testing"
)

func TestGetch(t *testing.T) {
	n, err := fmt.Printf("\033[s")
	fmt.Println(n, err)
	fmt.Printf("123123")
	var x, y int
	fmt.Scanf("\033[%d,%ds", &x, &y)
	fmt.Println(x, y)
	for {
		fmt.Printf("> ")
		getch()
	}
}
