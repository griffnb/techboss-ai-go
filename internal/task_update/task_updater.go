package task_update

import (
	"context"
	"sync"
)

type contextKey struct{}

type TaskUpdate[T any] struct {
	mu      sync.Mutex
	updates chan T
	closed  chan struct{}
	wg      sync.WaitGroup
	done    bool
}

func NewTaskUpdate[T any]() *TaskUpdate[T] {
	return &TaskUpdate[T]{
		updates: make(chan T, 10),
		closed:  make(chan struct{}),
	}
}

func (tu *TaskUpdate[T]) Send(update T) {
	tu.mu.Lock()
	if tu.done {
		tu.mu.Unlock()
		return
	}
	tu.wg.Add(1)
	tu.mu.Unlock()

	tu.updates <- update
}

func (tu *TaskUpdate[T]) Finish() {
	tu.mu.Lock()
	if tu.done {
		tu.mu.Unlock()
		return
	}
	tu.done = true
	tu.mu.Unlock()

	tu.wg.Wait()
	close(tu.updates)
	close(tu.closed)
}

func (tu *TaskUpdate[T]) Listen(handler func(T)) {
	go func() {
		for {
			select {
			case update, ok := <-tu.updates:
				if !ok {
					return
				}
				handler(update)
				tu.wg.Done()
			case <-tu.closed:
				return
			}
		}
	}()
}

func (tu *TaskUpdate[T]) Done() <-chan struct{} {
	return tu.closed
}

func WithTaskUpdater[T any](ctx context.Context, updater *TaskUpdate[T]) context.Context {
	return context.WithValue(ctx, contextKey{}, updater)
}

func GetTaskUpdater[T any](ctx context.Context) (*TaskUpdate[T], bool) {
	updater, ok := ctx.Value(contextKey{}).(*TaskUpdate[T])
	return updater, ok
}
