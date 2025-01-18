package time

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

type slot struct {
	timers list.List
	sync.Mutex
}

func (s *slot) add(t *tw) {
	s.Lock()
	s.addLocked(t)
	s.Unlock()
}

func (s *slot) addLocked(t *tw) {
	e := s.timers.PushBack(t)
	t.self = e
}

func (s *slot) removeLocked(t *tw) {
	s.timers.Remove(t.self)
}

func (s *slot) remove(t *tw) {
	s.Lock()
	s.timers.Remove(t.self)
	s.Unlock()
}

// 时间轮定时器容器
type TimerWheel struct {
	sc     int           // 有多少个槽(slot count)
	si     time.Duration // 每个槽的时间刻度(slot interval)
	sp     int           // 指向某个槽的指针(slot pointer)
	circle time.Duration // 转一圈所需时间
	slots  []*slot       // 所有的槽

	triggerFn func(fn func())

	// debug信息, 记录最长用时和最长队列长度
	longestUseTime atomic.Int64
	totalUseTime   atomic.Int64
	tickTimes      atomic.Int64
	longestListLen atomic.Int64
}

// NewExecutePoolTimerWheel 创建一个时间轮定时容器
// inputs:
//
//	sc: slot count 槽数量
//	si: slot interval 每个槽的时间刻度
//
// 槽的数量越多，刻度越小，则越精确，但是消耗的资源越多
func NewExecutePoolTimerWheel(sc int, si time.Duration, poolSize int) TimerContainer {
	if sc <= 0 || si <= 0 {
		panic("sc and si must be positive")
	}
	w := &TimerWheel{
		sc:     sc,
		si:     si,
		circle: time.Duration(sc) * si,
		slots:  make([]*slot, sc),
	}
	for i := range w.slots {
		w.slots[i] = &slot{}
		w.slots[i].timers.Init()
	}
	once := sync.Once{}
	ch := make(chan func(), 1024)
	w.triggerFn = func(fn func()) {
		once.Do(func() {
			for i := 0; i < poolSize; i++ {
				go func() {
					for {
						fn := <-ch
						if fn != nil {
							fn()
						}
					}
				}()
			}
		})

		select {
		case ch <- fn:
		default:
			go func() {
				ch <- fn
			}()
		}
	}

	go w.start()

	return w
}

// NewAsyncTimerWheel 创建一个时间轮定时容器
// inputs:
//
//	sc: slot count 槽数量
//	si: slot interval 每个槽的时间刻度
//
// 槽的数量越多，刻度越小，则越精确，但是消耗的资源越多
func NewAsyncTimerWheel(sc int, si time.Duration) TimerContainer {
	if sc <= 0 || si <= 0 {
		panic("sc and si must be positive")
	}
	w := &TimerWheel{
		sc:     sc,
		si:     si,
		circle: time.Duration(sc) * si,
		slots:  make([]*slot, sc),
	}
	for i := range w.slots {
		w.slots[i] = &slot{}
		w.slots[i].timers.Init()
	}

	w.triggerFn = func(fn func()) {
		if fn != nil {
			go fn()
		}
	}

	go w.start()

	return w
}

// NewSyncTimerWheel 创建一个时间轮定时容器
// inputs:
//
//	sc: slot count 槽数量
//	si: slot interval 每个槽的时间刻度
//
// 槽的数量越多，刻度越小，则越精确，但是消耗的资源越多
func NewSyncTimerWheel(sc int, si time.Duration) TimerContainer {
	if sc <= 0 || si <= 0 {
		panic("sc and si must be positive")
	}
	w := &TimerWheel{
		sc:     sc,
		si:     si,
		circle: time.Duration(sc) * si,
		slots:  make([]*slot, sc),
	}
	for i := range w.slots {
		w.slots[i] = &slot{}
		w.slots[i].timers.Init()
	}

	w.triggerFn = func(fn func()) {
		if fn != nil {
			fn()
		}
	}

	go w.start()

	return w
}

func (t *TimerWheel) SetTimer(d time.Duration, fn func()) Timer {
	return t.add(nil, d, fn, false, false)
}

func (t *TimerWheel) SetTicker(d time.Duration, fn func()) Ticker {
	return t.add(nil, d, fn, true, false)
}

func (t *TimerWheel) add(tm *tw, d time.Duration, fn func(), ticker, locked bool) *tw {
	if tm == nil {
		tm = &tw{
			w:        t,
			fn:       fn,
			duration: d,
			ticker:   ticker,
		}
	}
	// 一槽10ms 600圈 一圈6000ms  当前位于20槽位
	// 插入 15秒触发 = 15000
	// 需要经过 15000 / 10 = 1500 个槽
	// 需要 1500 / 600 = 2 圈
	// 槽位 1500 % 600 + 20 = 320

	// 计算需要多少圈
	n := int(d / t.circle)
	// 计算槽的位置
	pos := (int(d/t.si) + t.sp) % t.sc
	tm.n = n
	tm.pos = pos

	if locked {
		t.slots[pos].addLocked(tm)
		return tm
	}
	t.slots[pos].add(tm)

	return tm
}

func (t *TimerWheel) start() {
	i := 0
	for {
		time.Sleep(t.si)
		i++
		t.tick()
	}
}

func (t *TimerWheel) tick() {
	now := time.Now()
	defer func() {
		t.sp++
		t.sp %= t.sc
		du := time.Since(now)
		t.totalUseTime.Add(int64(du))
		t.tickTimes.Add(1)
		t.longestUseTime.Store(max(t.longestUseTime.Load(), int64(du)))
	}()

	s := t.slots[t.sp]
	s.Lock()
	defer s.Unlock()
	if s.timers.Len() == 0 {
		return
	}
	t.longestListLen.Store(max(t.longestListLen.Load(), int64(s.timers.Len())))

	for e := s.timers.Front(); e != nil; {
		tm := e.Value.(*tw)
		tm.n--
		if tm.n >= 0 {
			return
		}

		e = e.Next()
		t.triggerFn(tm.fn)
		s.removeLocked(tm)
		if tm.ticker {
			t.add(tm, tm.duration, tm.fn, tm.ticker, true)
		}
	}

}

func (t *TimerWheel) remove(tm *tw) {
	t.slots[tm.pos].remove(tm)
}

func (t *TimerWheel) DebugLongestTickTime() time.Duration {
	return time.Duration(t.longestUseTime.Load())
}

func (t *TimerWheel) DebugLongestSlotLen() int64 {
	return t.longestListLen.Load()
}

func (t *TimerWheel) DebugTotalTickTime() time.Duration {
	return time.Duration(t.totalUseTime.Load())
}

func (t *TimerWheel) DebugAvgTickTime() time.Duration {
	if t.tickTimes.Load() == 0 {
		return 0
	}
	return time.Duration(t.totalUseTime.Load() / t.tickTimes.Load())
}

type tw struct {
	w        *TimerWheel
	self     *list.Element
	n        int // 需要转多少圈触发
	pos      int // 在哪个槽
	fn       func()
	duration time.Duration
	ticker   bool
}

func (t *tw) Stop() {
	t.w.remove(t)
}

func (t *tw) Reset(d time.Duration) {
	t.w.remove(t)
	t.w.add(t, d, t.fn, t.ticker, false)
}
