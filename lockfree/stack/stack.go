package stack

import (
	"sync/atomic"
	"unsafe"
)

type stackItem struct {
	prev unsafe.Pointer
	val  interface{}
}

// Stack 栈
type Stack struct {
	top unsafe.Pointer
}

// New 初始化
func New() *Stack {
	return &Stack{}
}

// Push 添加
func (s *Stack) Push(item interface{}) {
	newItem := &stackItem{val: item}
	for {
		oldPointer := atomic.LoadPointer(&s.top)
		oldItem := (*stackItem)(oldPointer)
		if oldItem != nil {
			newItem.prev = oldPointer
		}
		newPointer := unsafe.Pointer(newItem)
		if atomic.CompareAndSwapPointer(&s.top, oldPointer, newPointer) {
			return
		}
	}
}

// Pop 取
func (s *Stack) Pop() (interface{}, bool) {
	for {
		oldTopPointer := atomic.LoadPointer(&s.top)
		if oldTopPointer == nil {
			return nil, false
		}
		oldTopItem := (*stackItem)(oldTopPointer)
		newTopPointer := atomic.LoadPointer(&oldTopItem.prev)
		if atomic.CompareAndSwapPointer(&s.top, oldTopPointer, newTopPointer) {
			return oldTopItem.val, true
		}
	}
}
