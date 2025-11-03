package worker

import (
	"fmt"
	"sync"
)

type GenericEnqueuer interface {
	EnqueueAny(data any)
}

type jobDispatcher struct {
	queues map[string]GenericEnqueuer
	mu     sync.RWMutex
}

var (
	dispatcher = &jobDispatcher{
		queues: make(map[string]GenericEnqueuer),
	}

	stopMu    sync.Mutex
	stopFuncs []func()
)

// RegisterEnqueuer registers a queue with optional stop function
func RegisterEnqueuer(name string, queue GenericEnqueuer, stopFn ...func()) {
	dispatcher.mu.Lock()
	defer dispatcher.mu.Unlock()
	dispatcher.queues[name] = queue

	if len(stopFn) > 0 {
		stopMu.Lock()
		stopFuncs = append(stopFuncs, stopFn[0])
		stopMu.Unlock()
	}
}

// Enqueue sends data into a registered queue
func Enqueue(name string, data any) {
	dispatcher.mu.RLock()
	defer dispatcher.mu.RUnlock()

	if q, ok := dispatcher.queues[name]; ok {
		q.EnqueueAny(data)
	} else {
		fmt.Printf("❌ No queue registered with name: %s\n", name)
	}
}

// StopAllWorkers stops all registered queues gracefully
func StopAllWorkers() {
	stopMu.Lock()
	defer stopMu.Unlock()
	for _, stop := range stopFuncs {
		stop()
	}
}
