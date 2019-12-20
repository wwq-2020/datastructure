package queue

import (
	"errors"
	"sync/atomic"
	"time"
	"unsafe"
)

var (
	// ErrQueueClosed 队列已关闭
	ErrQueueClosed = errors.New("queue closed")
	// ErrQueueEmpty 队列空
	ErrQueueEmpty = errors.New("queue empty")
	// ErrBPopTimeout bpop超时
	ErrBPopTimeout = errors.New("bpop timeout")
)

type item struct {
	val  interface{}
	next unsafe.Pointer
}

// Option 选项
type Option func(*option)

type option struct {
	blockCheckInterval time.Duration
}

// Queue 队列
type Queue struct {
	head     unsafe.Pointer
	tail     unsafe.Pointer
	readyCnt int64
	closing  uint32
	option   *option
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

// New 初始化队列
func New(opts ...Option) *Queue {
	option := defaultOption()
	for _, opt := range opts {
		opt(option)
	}
	item := &item{}
	pointer := unsafe.Pointer(item)
	q := &Queue{
		head:   pointer,
		tail:   pointer,
		option: option,
	}
	return q
}

// Push 添加元素
func (q *Queue) Push(val interface{}) {
	newItem := &item{val: val}
	newItemPointer := unsafe.Pointer(newItem)
	for {
		tailPointer := atomic.LoadPointer(&q.tail)
		oldTailPointer := tailPointer
		tail := (*item)(tailPointer)

		for !atomic.CompareAndSwapPointer(&tail.next, nil, newItemPointer) {
			for tail.next != nil {
				tail = (*item)(unsafe.Pointer(tail.next))
			}
		}

		if atomic.CompareAndSwapPointer(&q.tail, oldTailPointer, newItemPointer) {
			atomic.AddInt64(&q.readyCnt, 1)
			break
		}
	}

}

// Pop 获取元素
func (q *Queue) Pop() (interface{}, error) {
	return q.pop(false, nil)
}

// BPop 获取元素或者阻塞等待
func (q *Queue) BPop(timeout time.Duration) (interface{}, error) {
	var timerCh <-chan time.Time
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		timerCh = timer.C
	}
	return q.pop(true, timerCh)
}

func (q *Queue) pop(block bool, timerCh <-chan time.Time) (interface{}, error) {
	for {

		headPointer := atomic.LoadPointer(&q.head)
		head := (*item)(headPointer)

		next := atomic.LoadPointer(&head.next)
		if next == nil {
			if !block {
				return nil, ErrQueueEmpty
			}
			for {
				if atomic.LoadUint32(&q.closing) == 1 {
					return nil, ErrQueueClosed
				}
				select {
				case <-timerCh:
					return nil, ErrBPopTimeout
				default:
				}
				oldReadyCnt := atomic.LoadInt64(&q.readyCnt)
				if oldReadyCnt < 1 {
					time.Sleep(q.option.blockCheckInterval)
					continue
				}
				break
			}
			continue
		}
		nextPointer := unsafe.Pointer(next)

		if atomic.CompareAndSwapPointer(&q.head, headPointer, nextPointer) {
			nextItem := (*item)(unsafe.Pointer(next))
			atomic.AddInt64(&q.readyCnt, -1)
			return nextItem.val, nil
		}
	}
}

// Iter 迭代
func (q *Queue) Iter(fn func(interface{})) {
	for {
		if atomic.LoadUint32(&q.closing) == 1 {
			return
		}
		val, err := q.pop(true, nil)
		if err == nil {
			fn(val)
		}
	}
}

// Close 关闭
func (q *Queue) Close() {
	atomic.StoreUint32(&q.closing, 1)
}
