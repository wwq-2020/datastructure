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

type queueItem struct {
	val  interface{}
	next unsafe.Pointer
}

// Queue 队列
type Queue struct {
	head               unsafe.Pointer
	tail               unsafe.Pointer
	readyCnt           int64
	closing            uint32
	blockCheckInterval time.Duration
}

// New 初始化队列
func New() *Queue {
	item := &queueItem{}
	pointer := unsafe.Pointer(item)
	q := &Queue{
		head:               pointer,
		tail:               pointer,
		blockCheckInterval: 100 * time.Millisecond,
	}
	return q
}

// Push 添加元素
func (q *Queue) Push(val interface{}) {
	newItem := &queueItem{val: val}
	newItemPointer := unsafe.Pointer(newItem)
	tailPointer := atomic.LoadPointer(&q.tail)
	tail := (*queueItem)(tailPointer)

	for !atomic.CompareAndSwapPointer(&tail.next, nil, newItemPointer) {
		nextPointer := atomic.LoadPointer(&tail.next)
		for nextPointer != nil {
			tail = (*queueItem)(nextPointer)
			tailPointer = nextPointer
			nextPointer = atomic.LoadPointer(&tail.next)
		}
	}

	for {
		oldTailPointer := atomic.LoadPointer(&q.tail)
		nextPointer := atomic.LoadPointer(&tail.next)
		for nextPointer != nil {
			tail = (*queueItem)(nextPointer)
			tailPointer = nextPointer
			nextPointer = atomic.LoadPointer(&tail.next)
		}
		for atomic.CompareAndSwapPointer(&q.tail, oldTailPointer, tailPointer) {
			atomic.AddInt64(&q.readyCnt, 1)
			return
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
		head := (*queueItem)(headPointer)

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
					time.Sleep(q.blockCheckInterval)
					continue
				}
				break
			}
			continue
		}
		nextPointer := unsafe.Pointer(next)

		if atomic.CompareAndSwapPointer(&q.head, headPointer, nextPointer) {
			nextItem := (*queueItem)(unsafe.Pointer(next))
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
