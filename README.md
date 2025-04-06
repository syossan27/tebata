# Tebata（手旗）

[![GoDoc](https://godoc.org/github.com/syossan27/tebata?status.svg)](https://godoc.org/github.com/syossan27/tebata)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](https://github.com/syossan27/tebata/blob/master/LICENSE)

![logo](https://github.com/syossan27/tebata/blob/master/resources/logo.png)

## Overview

Simple linux signal handler for Go

## Installation

```
go get -u github.com/syossan27/tebata
```

## Usage

```go
package main
 
import (
	"fmt"
	"os"
	"strconv"
	"syscall"
	
	"github.com/syossan27/tebata"
)
 
func main() {
	t := tebata.New(syscall.SIGINT, syscall.SIGTERM)
	
	// Do function when catch signal.
	// Always check for errors when reserving functions
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
	// Use select{} for a non-CPU-consuming infinite wait
	select {}
}
 
func sum(firstArg, secondArg int) {
	fmt.Println(strconv.Itoa(firstArg + secondArg))	
}
 
func hello() {
	fmt.Println("Hello")	
}
 
// Expect output when type Ctrl + C:
//   3
//   Hello
```

### Error Handling

The `Reserve` method returns an error if:
- The provided function is not a valid function
- The argument types don't match the function parameter types
- Too few or too many arguments are provided

Always check the error return value when calling `Reserve` to ensure your signal handlers are properly registered.


## Why?

Look this article.

[How to implement a signal handler in Go](https://dev.to/syossan27/how-to-implement-a-signal-handler-in-go--582c)

## Artwork credits
The Go gopher was designed by [Renée French](http://reneefrench.blogspot.com/).  
The gopher side portrait is designed by [Takuya Ueda](https://twitter.com/tenntenn).  
Licensed under the "Creative CommonsAttribution 3.0" license.
