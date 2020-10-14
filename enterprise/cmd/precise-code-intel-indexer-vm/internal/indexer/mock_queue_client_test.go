// Code generated by github.com/efritz/go-mockgen 0.1.0; DO NOT EDIT.

package indexer

import (
	"context"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"sync"
)

// MockQueueClient is a mock implementation of the queueClient interface
// (from the package
// github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer-vm/internal/indexer)
// used for unit testing.
type MockQueueClient struct {
	// CompleteFunc is an instance of a mock function object controlling the
	// behavior of the method Complete.
	CompleteFunc *QueueClientCompleteFunc
	// DequeueFunc is an instance of a mock function object controlling the
	// behavior of the method Dequeue.
	DequeueFunc *QueueClientDequeueFunc
	// HeartbeatFunc is an instance of a mock function object controlling
	// the behavior of the method Heartbeat.
	HeartbeatFunc *QueueClientHeartbeatFunc
	// SetLogContentsFunc is an instance of a mock function object
	// controlling the behavior of the method SetLogContents.
	SetLogContentsFunc *QueueClientSetLogContentsFunc
}

// NewMockQueueClient creates a new mock of the queueClient interface. All
// methods return zero values for all results, unless overwritten.
func NewMockQueueClient() *MockQueueClient {
	return &MockQueueClient{
		CompleteFunc: &QueueClientCompleteFunc{
			defaultHook: func(context.Context, int, error) error {
				return nil
			},
		},
		DequeueFunc: &QueueClientDequeueFunc{
			defaultHook: func(context.Context) (store.Index, bool, error) {
				return store.Index{}, false, nil
			},
		},
		HeartbeatFunc: &QueueClientHeartbeatFunc{
			defaultHook: func(context.Context, []int) error {
				return nil
			},
		},
		SetLogContentsFunc: &QueueClientSetLogContentsFunc{
			defaultHook: func(context.Context, int, string) error {
				return nil
			},
		},
	}
}

// surrogateMockQueueClient is a copy of the queueClient interface (from the
// package
// github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer-vm/internal/indexer).
// It is redefined here as it is unexported in the source packge.
type surrogateMockQueueClient interface {
	Complete(context.Context, int, error) error
	Dequeue(context.Context) (store.Index, bool, error)
	Heartbeat(context.Context, []int) error
	SetLogContents(context.Context, int, string) error
}

// NewMockQueueClientFrom creates a new mock of the MockQueueClient
// interface. All methods delegate to the given implementation, unless
// overwritten.
func NewMockQueueClientFrom(i surrogateMockQueueClient) *MockQueueClient {
	return &MockQueueClient{
		CompleteFunc: &QueueClientCompleteFunc{
			defaultHook: i.Complete,
		},
		DequeueFunc: &QueueClientDequeueFunc{
			defaultHook: i.Dequeue,
		},
		HeartbeatFunc: &QueueClientHeartbeatFunc{
			defaultHook: i.Heartbeat,
		},
		SetLogContentsFunc: &QueueClientSetLogContentsFunc{
			defaultHook: i.SetLogContents,
		},
	}
}

// QueueClientCompleteFunc describes the behavior when the Complete method
// of the parent MockQueueClient instance is invoked.
type QueueClientCompleteFunc struct {
	defaultHook func(context.Context, int, error) error
	hooks       []func(context.Context, int, error) error
	history     []QueueClientCompleteFuncCall
	mutex       sync.Mutex
}

// Complete delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockQueueClient) Complete(v0 context.Context, v1 int, v2 error) error {
	r0 := m.CompleteFunc.nextHook()(v0, v1, v2)
	m.CompleteFunc.appendCall(QueueClientCompleteFuncCall{v0, v1, v2, r0})
	return r0
}

// SetDefaultHook sets function that is called when the Complete method of
// the parent MockQueueClient instance is invoked and the hook queue is
// empty.
func (f *QueueClientCompleteFunc) SetDefaultHook(hook func(context.Context, int, error) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Complete method of the parent MockQueueClient instance inovkes the hook
// at the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *QueueClientCompleteFunc) PushHook(hook func(context.Context, int, error) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *QueueClientCompleteFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, int, error) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *QueueClientCompleteFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, error) error {
		return r0
	})
}

func (f *QueueClientCompleteFunc) nextHook() func(context.Context, int, error) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *QueueClientCompleteFunc) appendCall(r0 QueueClientCompleteFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of QueueClientCompleteFuncCall objects
// describing the invocations of this function.
func (f *QueueClientCompleteFunc) History() []QueueClientCompleteFuncCall {
	f.mutex.Lock()
	history := make([]QueueClientCompleteFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// QueueClientCompleteFuncCall is an object that describes an invocation of
// method Complete on an instance of MockQueueClient.
type QueueClientCompleteFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 error
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c QueueClientCompleteFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c QueueClientCompleteFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// QueueClientDequeueFunc describes the behavior when the Dequeue method of
// the parent MockQueueClient instance is invoked.
type QueueClientDequeueFunc struct {
	defaultHook func(context.Context) (store.Index, bool, error)
	hooks       []func(context.Context) (store.Index, bool, error)
	history     []QueueClientDequeueFuncCall
	mutex       sync.Mutex
}

// Dequeue delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockQueueClient) Dequeue(v0 context.Context) (store.Index, bool, error) {
	r0, r1, r2 := m.DequeueFunc.nextHook()(v0)
	m.DequeueFunc.appendCall(QueueClientDequeueFuncCall{v0, r0, r1, r2})
	return r0, r1, r2
}

// SetDefaultHook sets function that is called when the Dequeue method of
// the parent MockQueueClient instance is invoked and the hook queue is
// empty.
func (f *QueueClientDequeueFunc) SetDefaultHook(hook func(context.Context) (store.Index, bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Dequeue method of the parent MockQueueClient instance inovkes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *QueueClientDequeueFunc) PushHook(hook func(context.Context) (store.Index, bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *QueueClientDequeueFunc) SetDefaultReturn(r0 store.Index, r1 bool, r2 error) {
	f.SetDefaultHook(func(context.Context) (store.Index, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *QueueClientDequeueFunc) PushReturn(r0 store.Index, r1 bool, r2 error) {
	f.PushHook(func(context.Context) (store.Index, bool, error) {
		return r0, r1, r2
	})
}

func (f *QueueClientDequeueFunc) nextHook() func(context.Context) (store.Index, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *QueueClientDequeueFunc) appendCall(r0 QueueClientDequeueFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of QueueClientDequeueFuncCall objects
// describing the invocations of this function.
func (f *QueueClientDequeueFunc) History() []QueueClientDequeueFuncCall {
	f.mutex.Lock()
	history := make([]QueueClientDequeueFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// QueueClientDequeueFuncCall is an object that describes an invocation of
// method Dequeue on an instance of MockQueueClient.
type QueueClientDequeueFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 store.Index
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 bool
	// Result2 is the value of the 3rd result returned from this method
	// invocation.
	Result2 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c QueueClientDequeueFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c QueueClientDequeueFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1, c.Result2}
}

// QueueClientHeartbeatFunc describes the behavior when the Heartbeat method
// of the parent MockQueueClient instance is invoked.
type QueueClientHeartbeatFunc struct {
	defaultHook func(context.Context, []int) error
	hooks       []func(context.Context, []int) error
	history     []QueueClientHeartbeatFuncCall
	mutex       sync.Mutex
}

// Heartbeat delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockQueueClient) Heartbeat(v0 context.Context, v1 []int) error {
	r0 := m.HeartbeatFunc.nextHook()(v0, v1)
	m.HeartbeatFunc.appendCall(QueueClientHeartbeatFuncCall{v0, v1, r0})
	return r0
}

// SetDefaultHook sets function that is called when the Heartbeat method of
// the parent MockQueueClient instance is invoked and the hook queue is
// empty.
func (f *QueueClientHeartbeatFunc) SetDefaultHook(hook func(context.Context, []int) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Heartbeat method of the parent MockQueueClient instance inovkes the hook
// at the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *QueueClientHeartbeatFunc) PushHook(hook func(context.Context, []int) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *QueueClientHeartbeatFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, []int) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *QueueClientHeartbeatFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, []int) error {
		return r0
	})
}

func (f *QueueClientHeartbeatFunc) nextHook() func(context.Context, []int) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *QueueClientHeartbeatFunc) appendCall(r0 QueueClientHeartbeatFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of QueueClientHeartbeatFuncCall objects
// describing the invocations of this function.
func (f *QueueClientHeartbeatFunc) History() []QueueClientHeartbeatFuncCall {
	f.mutex.Lock()
	history := make([]QueueClientHeartbeatFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// QueueClientHeartbeatFuncCall is an object that describes an invocation of
// method Heartbeat on an instance of MockQueueClient.
type QueueClientHeartbeatFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 []int
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c QueueClientHeartbeatFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c QueueClientHeartbeatFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// QueueClientSetLogContentsFunc describes the behavior when the
// SetLogContents method of the parent MockQueueClient instance is invoked.
type QueueClientSetLogContentsFunc struct {
	defaultHook func(context.Context, int, string) error
	hooks       []func(context.Context, int, string) error
	history     []QueueClientSetLogContentsFuncCall
	mutex       sync.Mutex
}

// SetLogContents delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockQueueClient) SetLogContents(v0 context.Context, v1 int, v2 string) error {
	r0 := m.SetLogContentsFunc.nextHook()(v0, v1, v2)
	m.SetLogContentsFunc.appendCall(QueueClientSetLogContentsFuncCall{v0, v1, v2, r0})
	return r0
}

// SetDefaultHook sets function that is called when the SetLogContents
// method of the parent MockQueueClient instance is invoked and the hook
// queue is empty.
func (f *QueueClientSetLogContentsFunc) SetDefaultHook(hook func(context.Context, int, string) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// SetLogContents method of the parent MockQueueClient instance inovkes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *QueueClientSetLogContentsFunc) PushHook(hook func(context.Context, int, string) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *QueueClientSetLogContentsFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, int, string) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *QueueClientSetLogContentsFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, string) error {
		return r0
	})
}

func (f *QueueClientSetLogContentsFunc) nextHook() func(context.Context, int, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *QueueClientSetLogContentsFunc) appendCall(r0 QueueClientSetLogContentsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of QueueClientSetLogContentsFuncCall objects
// describing the invocations of this function.
func (f *QueueClientSetLogContentsFunc) History() []QueueClientSetLogContentsFuncCall {
	f.mutex.Lock()
	history := make([]QueueClientSetLogContentsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// QueueClientSetLogContentsFuncCall is an object that describes an
// invocation of method SetLogContents on an instance of MockQueueClient.
type QueueClientSetLogContentsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c QueueClientSetLogContentsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c QueueClientSetLogContentsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}
