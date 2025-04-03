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

// ErrTypeMismatch is returned when the function arguments don't match the provided arguments.
var ErrTypeMismatch = errors.New("type mismatch: function parameter types don't match provided argument types")

// ErrTooFewArgs is returned when too few arguments are provided for the function.
var ErrTooFewArgs = errors.New("too few arguments: not enough arguments provided for function")

// ErrTooManyArgs is returned when too many arguments are provided for the function.
var ErrTooManyArgs = errors.New("too many arguments: too many arguments provided for function")

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
// It also validates that the provided arguments match the function's parameter types.
func (s *Tebata) Reserve(function any, args ...any) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	funcValue := reflect.ValueOf(function)
	if funcValue.Kind() != reflect.Func {
		return ErrInvalidFunction
	}

	// Get the function type to check parameter types
	funcType := funcValue.Type()
	numParams := funcType.NumIn()
	numArgs := len(args)

	// Check if we have too few arguments
	if numArgs < numParams {
		return ErrTooFewArgs
	}

	// Check if we have too many arguments
	if numArgs > numParams {
		return ErrTooManyArgs
	}

	// Check if argument types match parameter types
	for i := 0; i < numParams; i++ {
		paramType := funcType.In(i)

		// Skip nil arguments as they can be assigned to any type
		if args[i] == nil {
			continue
		}

		argType := reflect.TypeOf(args[i])

		// Check if the argument can be assigned to the parameter
		if !argType.AssignableTo(paramType) {
			return ErrTypeMismatch
		}
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
