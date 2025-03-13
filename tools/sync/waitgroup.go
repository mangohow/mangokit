package sync

import "sync"

type WaitGroup struct {
	wg sync.WaitGroup
}

func (wg *WaitGroup) Go(fn func()) {
	wg.wg.Add(1)
	go func() {
		defer wg.wg.Done()
		fn()
	}()
}

func (wg *WaitGroup) Wait() {
	wg.wg.Wait()
}
