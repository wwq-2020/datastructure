package ring

import (
	"errors"
	"sync/atomic"
	"time"
	"unsafe"
)

var (
	// ErrRingClosed 队列已关闭
	ErrRingClosed = errors.New("ring closed")
	// ErrRingFull 队列满
	ErrRingFull = errors.New("ring full")
	// ErrRingEmpty 队列空
	ErrRingEmpty = errors.New("ring empty")
	// ErrBPopTimeout bpop超时
	ErrBPopTimeout = errors.New("bpop timeout")
)

// Option 选项
type Option func(*option)

type option struct {
	blockCheckInterval time.Duration
}

type ringItem struct {
	val       interface{}
	valid     bool
	timestamp int64
}

// Ring 环形列表
type Ring struct {
	list      []unsafe.Pointer
	rPos      int32
	wPos      int32
	cap       int32
	option    *option
	closing   uint32
	rReadyCnt int64
	wReadyCnt int64
	first     int32
}

// WithBlockCheckInterval 添加
func WithBlockCheckInterval(blockCheckInterval time.Duration) Option {
	return func(option *option) {
		option.blockCheckInterval = blockCheckInterval
	}
}

func defaultOption() *option {
	return &option{
		blockCheckInterval: time.Millisecond * 100,
	}
}

// New 初始化Ring
func New(size int, opts ...Option) *Ring {
	option := defaultOption()
	for _, opt := range opts {
		opt(option)
	}
	list := make([]unsafe.Pointer, size)
	for i := range list {
		list[i] = unsafe.Pointer(&ringItem{})
	}
	return &Ring{
		list:   list,
		rPos:   -1,
		wPos:   -1,
		cap:    int32(size),
		option: option,
		first:  1,
	}
}

// Push 添加元素
func (r *Ring) Push(val interface{}) error {
	return r.push(val, false, nil)
}

// BPush 添加元素或者阻塞
func (r *Ring) BPush(val interface{}, timeout time.Duration) error {
	var timerCh <-chan time.Time
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		timerCh = timer.C
	}
	return r.push(val, true, timerCh)
}

func (r *Ring) push(val interface{}, block bool, timerCh <-chan time.Time) error {
	newRingItem := &ringItem{val: val, valid: true, timestamp: time.Now().Unix()}
	newRingItemPointer := unsafe.Pointer(newRingItem)
	var oldRPos, oldWPos, curCount, newWPos int32
	for {
		oldRPos = atomic.LoadInt32(&r.rPos)
		oldWPos = atomic.LoadInt32(&r.wPos)
		first := atomic.LoadInt32(&r.first)
		if oldWPos >= oldRPos {
			curCount = oldWPos - oldRPos
		} else {
			curCount = oldWPos + r.cap - oldRPos
		}

		if curCount >= r.cap && first == 0 {
			return ErrRingFull
		}
		newWPos = oldWPos + 1
		if newWPos >= r.cap-1 {
			atomic.StoreInt32(&r.first, 0)
		}
		if newWPos >= r.cap {
			newWPos = newWPos - r.cap
		}

		if atomic.CompareAndSwapInt32(&r.wPos, oldWPos, newWPos) {
			atomic.AddInt64(&r.wReadyCnt, -1)
			break
		}
		if !block {
			continue
		}
		for {
			if atomic.LoadUint32(&r.closing) == 1 {
				return ErrRingClosed
			}
			select {
			case <-timerCh:
				return ErrBPopTimeout
			default:
			}
			oldWReadyCnt := atomic.LoadInt64(&r.wReadyCnt)
			if oldWReadyCnt < 1 {
				time.Sleep(r.option.blockCheckInterval)
				continue
			}
			break
		}
	}

	for {
		oldPointer := atomic.LoadPointer(&r.list[newWPos])
		ringItem := (*ringItem)(oldPointer)
		if ringItem.timestamp > newRingItem.timestamp {
			return nil
		}
		if ringItem.valid {
			continue
		}
		if atomic.CompareAndSwapPointer(&r.list[newWPos], oldPointer, newRingItemPointer) {
			return nil
		}
	}
}

// Pop 获取元素
func (r *Ring) Pop() (interface{}, error) {
	return r.pop(false, nil)
}

// BPop 获取元素或者阻塞
func (r *Ring) BPop(timeout time.Duration) (interface{}, error) {
	var timerCh <-chan time.Time
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		timerCh = timer.C
	}
	return r.pop(true, timerCh)
}

// Iter 迭代
func (r *Ring) Iter(fn func(interface{})) {
	for {
		if atomic.LoadUint32(&r.closing) == 1 {
			return
		}
		val, err := r.pop(true, nil)
		if err == nil {
			fn(val)
		}
	}
}

func (r *Ring) pop(block bool, timerCh <-chan time.Time) (interface{}, error) {
	var oldRPos, oldWPos, newRPos, curCount int32
	for {
		oldRPos = atomic.LoadInt32(&r.rPos)
		oldWPos = atomic.LoadInt32(&r.wPos)
		if oldWPos >= oldRPos {
			curCount = oldWPos - oldRPos
		} else {
			curCount = oldWPos + r.cap - oldRPos
		}

		if curCount < 1 {
			return nil, ErrRingEmpty
		}
		newRPos = oldRPos + 1
		if newRPos >= r.cap {
			newRPos = newRPos - r.cap
		}

		if atomic.CompareAndSwapInt32(&r.rPos, oldRPos, newRPos) {
			atomic.AddInt64(&r.wReadyCnt, 1)
			break
		}
		if !block {
			continue
		}
		for {
			if atomic.LoadUint32(&r.closing) == 1 {
				return nil, ErrRingClosed
			}
			select {
			case <-timerCh:
				return nil, ErrBPopTimeout
			default:
			}
			oldRReadyCnt := atomic.LoadInt64(&r.rReadyCnt)
			if oldRReadyCnt < 1 {
				time.Sleep(r.option.blockCheckInterval)
				continue
			}
			break
		}
	}

	newRingItem := &ringItem{}
	newRingItemPointer := unsafe.Pointer(newRingItem)
	for {
		oldPointer := atomic.LoadPointer(&r.list[newRPos])
		ringItemx := (*ringItem)(oldPointer)

		if !ringItemx.valid {
			continue
		}
		if atomic.CompareAndSwapPointer(&r.list[newRPos], oldPointer, newRingItemPointer) {
			return ringItemx.val, nil
		}
	}
}

// Close 关闭
func (r *Ring) Close() {
	atomic.StoreUint32(&r.closing, 1)
}
