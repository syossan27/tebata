// Package tebata provides a way to handle OS signals gracefully.
package tebata

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"reflect"
	"sync"
)

// ErrInvalidFunction is returned when a non-function is passed to Reserve.
var ErrInvalidFunction = errors.New("invalid function argument: expected a function")

// ErrInvalidArgs is returned when invalid arguments are passed to Reserve.
var ErrInvalidArgs = errors.New("invalid args argument: expected a slice")

// Tebata handles signal-triggered function execution.
type Tebata struct {
	mutex             sync.Mutex
	ctx               context.Context
	cancel            context.CancelFunc
	signalCh          chan os.Signal
	reservedFunctions []functionData
}

// functionData stores a function and its arguments to be executed when a signal is received.
type functionData struct {
	function any
	args     []any
}

// New creates a new Tebata instance and starts listening for the specified signals.
// It uses context for better cancellation support.
func New(signals ...os.Signal) *Tebata {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Tebata{
		ctx:      ctx,
		cancel:   cancel,
		signalCh: make(chan os.Signal, 1),
	}
	signal.Notify(s.signalCh, signals...)
	go s.listen()
	return s
}

// listen waits for signals and executes reserved functions when signals are received.
func (s *Tebata) listen() {
	for {
		select {
		case <-s.signalCh:
			s.exec()
		case <-s.ctx.Done():
			return
		}
	}
}

// exec executes all reserved functions.
func (s *Tebata) exec() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, fd := range s.reservedFunctions {
		// Use reflection to call the function with its arguments
		function := reflect.ValueOf(fd.function)
		var args []reflect.Value
		for _, arg := range fd.args {
			args = append(args, reflect.ValueOf(arg))
		}
		function.Call(args)
	}
}

// Reserve registers a function to be executed when a signal is received.
// It returns an error if the function or arguments are invalid.
func (s *Tebata) Reserve(function any, args ...any) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if reflect.ValueOf(function).Kind() != reflect.Func {
		return ErrInvalidFunction
	}

	s.reservedFunctions = append(
		s.reservedFunctions,
		functionData{
			function: function,
			args:     args,
		},
	)

	return nil
}

// Close stops the signal handling and cleans up resources.
func (s *Tebata) Close() {
	s.cancel()
	signal.Stop(s.signalCh)
	close(s.signalCh)
}
