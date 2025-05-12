package util

import "sync"

type Stack struct {
	Top  int //栈顶元素的索引
	List []string
	Lock sync.Mutex
}

// 入栈
func (s *Stack) push(val string) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	s.Top++
	s.List = append(s.List, val)
}

// 出栈
func (s *Stack) pop() (string, bool) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	if s.Top == 0 {
		return "", false
	}

	val := s.List[s.Top-1]
	s.List = s.List[:s.Top-1]
	s.Top--
	return val, true
}

//判断栈是否为空
func (s *Stack) stackIsEmpty() bool {
	return s.Top == 0
}
