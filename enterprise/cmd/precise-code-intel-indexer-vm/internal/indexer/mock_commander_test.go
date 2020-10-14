// Code generated by github.com/efritz/go-mockgen 0.1.0; DO NOT EDIT.

package indexer

import (
	"context"
	"sync"
)

// MockCommander is a mock implementation of the Commander interface (from
// the package
// github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer-vm/internal/indexer)
// used for unit testing.
type MockCommander struct {
	// RunFunc is an instance of a mock function object controlling the
	// behavior of the method Run.
	RunFunc *CommanderRunFunc
}

// NewMockCommander creates a new mock of the Commander interface. All
// methods return zero values for all results, unless overwritten.
func NewMockCommander() *MockCommander {
	return &MockCommander{
		RunFunc: &CommanderRunFunc{
			defaultHook: func(context.Context, *CommandLogger, ...string) error {
				return nil
			},
		},
	}
}

// NewMockCommanderFrom creates a new mock of the MockCommander interface.
// All methods delegate to the given implementation, unless overwritten.
func NewMockCommanderFrom(i Commander) *MockCommander {
	return &MockCommander{
		RunFunc: &CommanderRunFunc{
			defaultHook: i.Run,
		},
	}
}

// CommanderRunFunc describes the behavior when the Run method of the parent
// MockCommander instance is invoked.
type CommanderRunFunc struct {
	defaultHook func(context.Context, *CommandLogger, ...string) error
	hooks       []func(context.Context, *CommandLogger, ...string) error
	history     []CommanderRunFuncCall
	mutex       sync.Mutex
}

// Run delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockCommander) Run(v0 context.Context, v1 *CommandLogger, v2 ...string) error {
	r0 := m.RunFunc.nextHook()(v0, v1, v2...)
	m.RunFunc.appendCall(CommanderRunFuncCall{v0, v1, v2, r0})
	return r0
}

// SetDefaultHook sets function that is called when the Run method of the
// parent MockCommander instance is invoked and the hook queue is empty.
func (f *CommanderRunFunc) SetDefaultHook(hook func(context.Context, *CommandLogger, ...string) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Run method of the parent MockCommander instance inovkes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *CommanderRunFunc) PushHook(hook func(context.Context, *CommandLogger, ...string) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *CommanderRunFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, *CommandLogger, ...string) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *CommanderRunFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *CommandLogger, ...string) error {
		return r0
	})
}

func (f *CommanderRunFunc) nextHook() func(context.Context, *CommandLogger, ...string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CommanderRunFunc) appendCall(r0 CommanderRunFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of CommanderRunFuncCall objects describing the
// invocations of this function.
func (f *CommanderRunFunc) History() []CommanderRunFuncCall {
	f.mutex.Lock()
	history := make([]CommanderRunFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CommanderRunFuncCall is an object that describes an invocation of method
// Run on an instance of MockCommander.
type CommanderRunFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 *CommandLogger
	// Arg2 is a slice containing the values of the variadic arguments
	// passed to this method invocation.
	Arg2 []string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation. The variadic slice argument is flattened in this array such
// that one positional argument and three variadic arguments would result in
// a slice of four, not two.
func (c CommanderRunFuncCall) Args() []interface{} {
	trailing := []interface{}{}
	for _, val := range c.Arg2 {
		trailing = append(trailing, val)
	}

	return append([]interface{}{c.Arg0, c.Arg1}, trailing...)
}

// Results returns an interface slice containing the results of this
// invocation.
func (c CommanderRunFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}
