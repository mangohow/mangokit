package parallel

import (
	"context"
	"fmt"
	"sync"
)

type DoWorkPieceFunc[T any] func(piece int, val T) error

type DoWorkPieceResultFucnc[T, R any] func(piece int, val T) (R, error)

type options struct {
	chunkSize int
	stopOnErr bool
}

type Options func(*options)

// WithChunkSize allows to set chunks of work items to the workers, rather than
// processing one by one.
// It is recommended to use this option if the number of pieces significantly
// higher than the number of workers and the work done for each item is small.
func WithChunkSize(c int) func(*options) {
	return func(o *options) {
		o.chunkSize = c
	}
}

func WithStopOnError(stop bool) func(*options) {
	return func(o *options) {
		o.stopOnErr = stop
	}
}

// Parallelize is a framework that allows for parallelizing N
// independent pieces of work until done or the context is canceled.
func Parallelize[T any](ctx context.Context, workers int, tasks []T, doWorkPiece DoWorkPieceFunc[T], opts ...Options) error {
	pieces := len(tasks)
	if pieces == 0 {
		return nil
	}
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}
	chunkSize := o.chunkSize
	if chunkSize < 1 {
		chunkSize = 1
	}

	chunks := ceilDiv(pieces, chunkSize)
	toProcess := make(chan int, chunks)
	for i := 0; i < chunks; i++ {
		toProcess <- i
	}
	close(toProcess)

	if ctx == nil {
		ctx = context.Background()
	}
	stopCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	if chunks < workers {
		workers = chunks
	}
	errCh := make(chan error, 1)
	wg := sync.WaitGroup{}
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {

				}
			}()
			defer wg.Done()
			for chunk := range toProcess {
				start := chunk * chunkSize
				end := start + chunkSize
				if end > pieces {
					end = pieces
				}
				for p := start; p < end; p++ {
					select {
					case <-stopCtx.Done():
						return
					default:
						err := doWorkPiece(p, tasks[p])
						if err != nil && o.stopOnErr {
							select {
							case errCh <- err:
							default:
							}
							cancel()
							return
						}
					}
				}
			}
		}()
	}
	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
	}

	return nil
}

func ParallelizeResult[T, R any](ctx context.Context, workers int, tasks []T, doWorkPiece DoWorkPieceResultFucnc[T, R], opts ...Options) ([]R, error) {
	pieces := len(tasks)
	if pieces == 0 {
		return nil, nil
	}
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}
	chunkSize := o.chunkSize
	if chunkSize < 1 {
		chunkSize = 1
	}

	chunks := ceilDiv(pieces, chunkSize)
	toProcess := make(chan int, chunks)
	for i := 0; i < chunks; i++ {
		toProcess <- i
	}
	close(toProcess)

	if ctx == nil {
		ctx = context.Background()
	}
	stopCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	if chunks < workers {
		workers = chunks
	}
	wg := sync.WaitGroup{}
	result := make([]R, 0, pieces)
	chSize := pieces
	if chSize > 1024 {
		chSize = 1024
	}
	errCh := make(chan error, 1)
	resultCh := make(chan R, chSize)
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-stopCtx.Done():
				return
			case r := <-resultCh:
				result = append(result, r)
				if len(result) == pieces {
					return
				}
			}
		}
	}()

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					if !o.stopOnErr {
						return
					}
					var (
						err error
						ok  bool
					)
					if err, ok = r.(error); !ok {
						err = fmt.Errorf("recovered panic: %v", r)
					}

					select {
					case errCh <- err:
						cancel()
					default:
					}
				}
			}()
			defer wg.Done()
			for chunk := range toProcess {
				start := chunk * chunkSize
				end := start + chunkSize
				if end > pieces {
					end = pieces
				}
				for p := start; p < end; p++ {
					select {
					case <-stopCtx.Done():
						return
					default:
						v, err := doWorkPiece(p, tasks[p])
						if err != nil && o.stopOnErr {
							select {
							case errCh <- err:
							default:
							}
							cancel()
							return
						}
						resultCh <- v
					}
				}
			}
		}()
	}
	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
	}

	return result, nil
}

func ceilDiv(a, b int) int {
	return (a + b - 1) / b
}
