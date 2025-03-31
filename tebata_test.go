package tebata

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"syscall"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	// Test that New creates a Tebata instance with the correct signals
	s := New(syscall.SIGINT, syscall.SIGTERM)
	defer s.Close()

	if s == nil {
		t.Fatal("New returned nil")
	}

	if s.ctx == nil {
		t.Error("Context is nil")
	}

	if s.cancel == nil {
		t.Error("Cancel function is nil")
	}

	if s.signalCh == nil {
		t.Error("Signal channel is nil")
	}

	// Test that the signal channel receives signals
	go func() {
		time.Sleep(100 * time.Millisecond)
		s.signalCh <- os.Interrupt
	}()

	select {
	case <-s.signalCh:
		// Success
	case <-time.After(500 * time.Millisecond):
		t.Error("Signal channel did not receive signal")
	}
}

func TestStatus_Reserve(t *testing.T) {
	s := New(syscall.SIGINT)
	defer s.Close()

	// Test reserving a valid function
	err := s.Reserve(func() {}, nil)
	if err != nil {
		t.Errorf("Reserve returned error for valid function: %v", err)
	}

	// Test reserving a non-function
	err = s.Reserve("not a function")
	if err == nil {
		t.Error("Reserve did not return error for non-function")
	}
	if err != ErrInvalidFunction {
		t.Errorf("Expected ErrInvalidFunction, got: %v", err)
	}

	// Test that the function was added to reservedFunctions
	if len(s.reservedFunctions) != 1 {
		t.Errorf("Expected 1 reserved function, got %d", len(s.reservedFunctions))
	}
}

func TestStatus_exec(t *testing.T) {
	done := make(chan int, 1)

	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	s := New(syscall.SIGINT, syscall.SIGTERM)
	defer s.Close()

	s.Reserve(
		func(first, second int, done chan int) {
			fmt.Print(strconv.Itoa(first + second))
			done <- 1
		},
		1, 2, done,
	)

	s.signalCh <- os.Interrupt
	<-done

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = stdout

	if buf.Len() == 0 {
		t.Error("Output empty")
	}

	if buf.String() != "3" {
		t.Error("Invalid output")
	}
}

func TestStatus_exec_race_check(t *testing.T) {
	done := make(chan int, 1)

	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	s1 := New(syscall.SIGINT, syscall.SIGTERM)
	defer s1.Close()

	s1.Reserve(
		func(first, second int, done chan int) {
			fmt.Print(strconv.Itoa(first + second))
			done <- 1
		},
		1, 2, done,
	)

	s2 := New(syscall.SIGINT, syscall.SIGTERM)
	defer s2.Close()

	s2.Reserve(
		func(first, second int, done chan int) {
			fmt.Print(strconv.Itoa(first + second))
			done <- 1
		},
		1, 2, done,
	)

	s1.signalCh <- os.Interrupt
	s2.signalCh <- os.Interrupt
	<-done

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = stdout

	if buf.Len() == 0 {
		t.Error("Output empty")
	}
}

func TestClose(t *testing.T) {
	s := New(syscall.SIGINT)

	// Close the Tebata instance
	s.Close()

	// Test that the context is canceled
	select {
	case <-s.ctx.Done():
		// Success
	default:
		t.Error("Context was not canceled")
	}

	// Test that sending a signal after Close doesn't trigger exec
	// This is a bit tricky to test directly, but we can check that
	// the signal channel is closed
	_, ok := <-s.signalCh
	if ok {
		t.Error("Signal channel was not closed")
	}
}
