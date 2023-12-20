package parallel

import (
	"context"
	"sync"
)

func Parallel[T any](ctx context.Context, parallel int, tasks []T, taskFn func(task T) error) error {
	_, err := ParallelResult[T, int](ctx, parallel, tasks, nil, taskFn)
	return err
}

// ParallelResult 启动parallel个goroutine并行执行tasks，如果遇到错误则终止执行
func ParallelResult[T1, T2 any](ctx context.Context, parallel int, tasks []T1, resultCh chan T2, taskFn func(task T1) error) ([]T2, error) {
	var (
		wg       = &sync.WaitGroup{}
		taskCh   = make(chan T1, 128)
		counter  = make(chan struct{}, 64)
		errCh    = make(chan error, 1)
		finished = 0
		result   = make([]T2, 0)
		err      error
	)
	c, cancel := context.WithCancel(ctx)

	// 启动一个goroutine来分发任务
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < len(tasks); i++ {
			// 防止发生错误时，其它worker退出，导致taskCh被填满，
			// 从而导致当前goroutine阻塞在taskCh上，造成死锁
			select {
			case <-c.Done():
				return
			case taskCh <- tasks[i]:
			}
		}
	}()

	// 启动parallel个goroutine来处理任务
	wg.Add(parallel)
	for i := 0; i < parallel; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				// 退出信号
				case <-c.Done():
					return
				// 从taskCh中获取任务
				case task := <-taskCh:
					// 如果发生错误，则投递到errCh中
					if err := taskFn(task); err != nil {
						// 防止多个goroutine都阻塞在errCh上
						select {
						case errCh <- err:
						default:
						}
						cancel()
						return
					}
					// 通知主goroutine任务完成一个
					counter <- struct{}{}
				}
			}
		}()
	}

loop:
	for {
		select {
		// 用户退出信号
		case <-ctx.Done():
			err = ctx.Err()
			break loop
		// 记录任务完成数量，如果任务全部完成，则通知worker退出
		case <-counter:
			finished++
			if finished == len(tasks) {
				break loop
			}
		// 获取发生的错误，并且结束其它worker的后续的执行
		case err = <-errCh:
			break loop
		// 获取结果，保存到切片中
		case res := <-resultCh:
			result = append(result, res)
		}
	}

	// 通知所有其它goroutine退出
	cancel()

	// 由于select是随机的，因此可能出现下面的情况：
	// 在上面for中的select中随机到了counter，结束了循环
	// 但是结果可能还没有收集完毕，因此在下面再次收集
sloop:
	for {
		select {
		case res := <-resultCh:
			result = append(result, res)
		default:
			break sloop
		}
	}

	// 等待其它goroutine全部退出
	wg.Wait()

	return result, err
}
