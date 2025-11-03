package worker

import "context"

type Sender[T any] interface {
	Send(ctx context.Context, data T) error
}

type Processor[T any] interface {
	Process(data T)
}
