package worker

import (
	"context"
	"sync"
)

type QueueWorker[T any] struct {
	queue     chan T
	processor Processor[T]
	once      sync.Once
}

func NewQueueWorker[T any](bufferSize int, processor Processor[T]) *QueueWorker[T] {
	q := &QueueWorker[T]{
		queue:     make(chan T, bufferSize),
		processor: processor,
	}
	go q.start()
	return q
}

func NewSenderWorker[T any](bufferSize int, sender Sender[T]) *QueueWorker[T] {
	return NewQueueWorker(bufferSize, &senderProcessor[T]{sender: sender})
}

type senderProcessor[T any] struct {
	sender Sender[T]
}

func (s *senderProcessor[T]) Process(data T) {
	_ = s.sender.Send(context.Background(), data)
}

func (q *QueueWorker[T]) start() {
	for task := range q.queue {
		q.processor.Process(task)
	}
}

func (q *QueueWorker[T]) Enqueue(task T) {
	select {
	case q.queue <- task:
	default:
		// Optional: log drop
	}
}

func (q *QueueWorker[T]) EnqueueAny(data any) {
	if t, ok := data.(T); ok {
		q.Enqueue(t)
	}
}

func (q *QueueWorker[T]) Stop() {
	q.once.Do(func() {
		close(q.queue)
	})
}
