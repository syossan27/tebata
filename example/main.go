package main

import (
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/syossan27/tebata"
)

// Expect output when press the Ctrl + C:
//
//	3
//	Hello
func main() {
	t := tebata.New(syscall.SIGINT, syscall.SIGTERM)

	// Do function when catch signal.
	t.Reserve(sum, 1, 2)
	t.Reserve(hello)
	t.Reserve(os.Exit, 0)

	for {
		// Do something
	}
}

func sum(firstArg, secondArg int) {
	fmt.Println(strconv.Itoa(firstArg + secondArg))
}

func hello() {
	fmt.Println("Hello")
}
