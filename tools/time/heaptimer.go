package time

import (
	"container/heap"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

type HeapTimer struct {
	timers    []*timer
	addCh     chan *timer
	modCh     chan modTimer
	removeCh  chan *timer
	triggerFn func(t *timer)
}

type modTimer struct {
	t *timer
	d time.Duration
}

func NewExecutePoolHeapTimer(poolSize int) TimerContainer {
	t := &HeapTimer{
		addCh:    make(chan *timer, 1024),
		modCh:    make(chan modTimer, 1024),
		removeCh: make(chan *timer, 1024),
	}

	once := sync.Once{}
	ch := make(chan *timer, 1024)
	// 触发回调函数时，通过协程池去处理
	t.triggerFn = func(t *timer) {
		once.Do(func() {
			for i := 0; i < poolSize; i++ {
				go func() {
					for {
						t := <-ch
						if t.fn != nil {
							t.fn()
						}
					}
				}()
			}
		})

		select {
		case ch <- t:
		default:
			go func() {
				ch <- t
			}()
		}
	}

	go t.tick()

	return t
}

func NewAsyncHeapTimer() TimerContainer {
	t := &HeapTimer{
		addCh:    make(chan *timer, 1024),
		modCh:    make(chan modTimer, 1024),
		removeCh: make(chan *timer, 1024),
	}

	// 触发回调函数时，直接启动一个Goroutine去处理
	t.triggerFn = func(t *timer) {
		if t.fn != nil {
			go t.fn()
		}
	}

	go t.tick()

	return t
}

func NewSyncHeapTimer() TimerContainer {
	t := &HeapTimer{
		addCh:    make(chan *timer, 1024),
		modCh:    make(chan modTimer, 1024),
		removeCh: make(chan *timer, 1024),
	}

	// 触发回调函数时，直接同步处理
	t.triggerFn = func(t *timer) {
		if t.fn != nil {
			t.fn()
		}
	}

	go t.tick()

	return t
}

func (h *HeapTimer) Len() int {
	return len(h.timers)
}

func (h *HeapTimer) Less(i, j int) bool {
	return h.timers[i].trigger.Before(h.timers[j].trigger)
}

func (h *HeapTimer) Swap(i, j int) {
	h.timers[i], h.timers[j] = h.timers[j], h.timers[i]
}

func (h *HeapTimer) Push(x any) {
	h.timers = append(h.timers, x.(*timer))
}

func (h *HeapTimer) Pop() any {
	n := len(h.timers) - 1
	t := h.timers[n]
	// 置为nil让垃圾回收器可以回收
	h.timers[n] = nil
	h.timers = h.timers[:n]
	return t
}

func (h *HeapTimer) top() *timer {
	return h.timers[0]
}

// 通过一个goroutine来处理所有定时器的添加、重置、删除以及定时器的触发
// 采取无锁化的方法
func (h *HeapTimer) tick() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "tick panic: %v\n%s", r, string(debug.Stack()))
			go h.tick()
		}
	}()

	minTrigger := time.Hour
	if len(h.timers) > 0 {
		minTrigger = h.top().trigger.Sub(time.Now())
		if minTrigger < 0 {
			minTrigger = 0
		}
	}
	sysTm := time.NewTimer(minTrigger)
	defer sysTm.Stop()
	// 使用非阻塞的方式添加，防止阻塞该goroutine
	addFn := func(t *timer) {
		select {
		case h.addCh <- t:
		default:
			go func() {
				h.addCh <- t
			}()
		}
	}

	for {
		select {
		// 向堆中添加一个定时器
		case t := <-h.addCh:
			// 1. 如果添加的是第一个定时器，那么它肯定是最早触发的，需要重置定时器
			// 2. 和堆顶的最小触发定时器进行比较，如果刚添加的更早触发，则需要重置系统定时器
			minTrigger = time.Duration(-1)
			if h.Len() == 0 || (h.Len() > 0 && h.top().trigger.After(t.trigger)) {
				minTrigger = t.trigger.Sub(time.Now())
			}
			if minTrigger != -1 {
				sysTm.Reset(minTrigger)
			}
			heap.Push(h, t)
		// 删除原来的定时器，重新添加一个
		case t := <-h.modCh:
			tt := *t.t
			tt.duration = t.d
			tt.trigger = time.Now().Add(t.d)
			t.t.removed = true
			addFn(&tt)
		// 删除时，只设置删除标志
		case t := <-h.removeCh:
			t.removed = true
		case <-sysTm.C:
		}
		now := time.Now()
		for h.Len() > 0 {
			m := h.top()
			// 如果堆顶的定时器没有触发, 设置最小触发时间，等下次触发
			if m.trigger.After(now) {
				sysTm.Reset(m.trigger.Sub(now))
				break
			}

			// 定时器触发，检查是否被删除了
			// 如果被删除了，从堆中移除，并且再次检查堆顶
			if m.removed {
				heap.Pop(h)
				continue
			}

			// 定时器触发了，从堆中删除，并执行回调函数
			heap.Pop(h)
			h.triggerFn(m)

			// 如果是trigger，则重新添加到定时器集合中
			if m.ticker {
				m.trigger = now.Add(m.duration)
				addFn(m)
			}
		}

		// 如果堆中没有元素，则休眠一小时
		if h.Len() == 0 {
			sysTm.Reset(time.Hour)
			continue
		}
	}

}

func (h *HeapTimer) add(d time.Duration, fn func(), ticker bool) *timer {
	if fn == nil {
		panic("nil function")
	}
	if d < 0 {
		panic("negative duration")
	}

	tm := &timer{
		h:        h,
		fn:       fn,
		ticker:   ticker,
		trigger:  time.Now().Add(d),
		duration: d,
	}

	h.addCh <- tm

	return tm
}

func (h *HeapTimer) remove(t *timer) {
	h.removeCh <- t
}

func (h *HeapTimer) reset(tm *timer, duration time.Duration) {
	if duration < 0 {
		panic("negative duration")
	}
	h.modCh <- modTimer{
		t: tm,
		d: duration,
	}
}

func (h *HeapTimer) SetTimer(duration time.Duration, fn func()) Timer {
	return h.add(duration, fn, false)
}

func (h *HeapTimer) SetTicker(duration time.Duration, fn func()) Ticker {
	return h.add(duration, fn, true)
}

type timer struct {
	h        *HeapTimer
	fn       func()
	ticker   bool
	trigger  time.Time
	duration time.Duration
	removed  bool
}

func (t *timer) Stop() {
	t.h.remove(t)
}

func (t *timer) Reset(duration time.Duration) {
	t.h.reset(t, duration)
}
