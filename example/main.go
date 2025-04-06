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
	if err := t.Reserve(sum, 1, 2); err != nil {
		fmt.Printf("Failed to reserve sum function: %v\n", err)
		return
	}
	if err := t.Reserve(hello); err != nil {
		fmt.Printf("Failed to reserve hello function: %v\n", err)
		return
	}
	if err := t.Reserve(os.Exit, 0); err != nil {
		fmt.Printf("Failed to reserve exit function: %v\n", err)
		return
	}

	fmt.Println("Signal handler registered. Press Ctrl+C to trigger.")
	select {}
}

func sum(firstArg, secondArg int) {
	fmt.Println(strconv.Itoa(firstArg + secondArg))
}

func hello() {
	fmt.Println("Hello")
}
