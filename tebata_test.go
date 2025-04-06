package tebata

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"syscall"
	"testing"
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
}

func TestStatus_Reserve(t *testing.T) {
	s := New(syscall.SIGINT)
	defer s.Close()

	// Test reserving a valid function
	err := s.Reserve(func() {})
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

	// Test with correct argument types
	err = s.Reserve(func(a string, b int) {}, "test", 123)
	if err != nil {
		t.Errorf("Reserve returned error for valid function with correct args: %v", err)
	}

	// Test with too few arguments
	err = s.Reserve(func(a string, b int) {}, "test")
	if err == nil {
		t.Error("Reserve did not return error for too few arguments")
	}
	if err != ErrTooFewArgs {
		t.Errorf("Expected ErrTooFewArgs, got: %v", err)
	}

	// Test with too many arguments
	err = s.Reserve(func(a string) {}, "test", 123)
	if err == nil {
		t.Error("Reserve did not return error for too many arguments")
	}
	if err != ErrTooManyArgs {
		t.Errorf("Expected ErrTooManyArgs, got: %v", err)
	}

	// Test with incorrect argument types
	err = s.Reserve(func(a string, b int) {}, 123, "test")
	if err == nil {
		t.Error("Reserve did not return error for incorrect argument types")
	}
	if err != ErrTypeMismatch {
		t.Errorf("Expected ErrTypeMismatch, got: %v", err)
	}

	// Test with nil argument (should work for any type)
	err = s.Reserve(func(a *string) {}, nil)
	if err != nil {
		t.Errorf("Reserve returned error for nil argument: %v", err)
	}
}

func TestStatus_exec(t *testing.T) {
	done := make(chan int, 1)

	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	s := New(syscall.SIGINT, syscall.SIGTERM)
	defer s.Close()

	if err := s.Reserve(
		func(first, second int, done chan int) {
			fmt.Print(strconv.Itoa(first + second))
			done <- 1
		},
		1, 2, done,
	); err != nil {
		t.Fatalf("Failed to reserve function: %v", err)
	}

	s.signalCh <- os.Interrupt
	<-done

	if err := w.Close(); err != nil {
		t.Errorf("Failed to close pipe writer: %v", err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Errorf("Failed to copy from pipe reader: %v", err)
	}
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

	if err := s1.Reserve(
		func(first, second int, done chan int) {
			fmt.Print(strconv.Itoa(first + second))
			done <- 1
		},
		1, 2, done,
	); err != nil {
		t.Fatalf("Failed to reserve function for s1: %v", err)
	}

	s2 := New(syscall.SIGINT, syscall.SIGTERM)
	defer s2.Close()

	if err := s2.Reserve(
		func(first, second int, done chan int) {
			fmt.Print(strconv.Itoa(first + second))
			done <- 1
		},
		1, 2, done,
	); err != nil {
		t.Fatalf("Failed to reserve function for s2: %v", err)
	}

	s1.signalCh <- os.Interrupt
	s2.signalCh <- os.Interrupt
	<-done

	if err := w.Close(); err != nil {
		t.Errorf("Failed to close pipe writer: %v", err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Errorf("Failed to copy from pipe reader: %v", err)
	}
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
