package util

import "sync"

type Queue struct {
	List  []interface{}
	Lock  sync.Mutex
	Begin uint64 //首节点下标
	End   uint64 //尾节点下标
	Cap   uint64 // 容量
}

func New() (q *Queue) {
	return &Queue{
		List:  make([]interface{}, 1, 1),
		Begin: 0,
		End:   0,
		Cap:   1,
		Lock:  sync.Mutex{},
	}
}

// Push 将元素添加到该队列末尾
func (q *Queue) Push(val interface{}) {
	if q == nil {
		q = New()
	}
	q.Lock.Lock()
	if q.End < q.Cap {
		//不需要扩容
		q.List[q.End] = val
	} else {
		//需要扩容
		if q.Begin > 0 {
			//首部有冗余,整体前移
			for i := uint64(0); i < q.End-q.Begin; i++ {
				q.List[i] = q.List[i+q.Begin]
			}
			q.End -= q.Begin
			q.Begin = 0
		} else {
			//冗余不足,需要扩容
			if q.Cap <= 65536 {
				//容量翻倍
				if q.Cap == 0 {
					q.Cap = 1
				}
				q.Cap *= 2
			} else {
				//容量增加2^16
				q.Cap += 2 ^ 16
			}
			//复制扩容前的元素
			tmp := make([]interface{}, q.Cap, q.Cap)
			copy(tmp, q.List)
			q.List = tmp
		}
		q.List[q.End] = val
	}
	q.End++
	q.Lock.Unlock()
}

func (q *Queue) QueueIsEmpty() (b bool) {
	if q == nil {
		q = New()
	}
	return q.Size() <= 0
}

func (q *Queue) Size() (num uint64) {
	if q == nil {
		q = New()
	}
	return q.End - q.Begin
}

func (q *Queue) Clear() {
	if q == nil {
		q = New()
	}
	q.Lock.Lock()
	q.List = make([]interface{}, 1, 1)
	q.Begin = 0
	q.End = 0
	q.Cap = 1
	q.Lock.Unlock()
}

// Pop 队首元素弹出队列
func (q *Queue) Pop() (val interface{}) {
	if q == nil {
		q = New()
		return nil
	}
	if q.QueueIsEmpty() {
		q.Clear()
		return nil
	}
	q.Lock.Lock()
	val = q.List[q.Begin]
	q.Begin++
	if q.Begin >= 1024 || q.Begin*2 > q.End {
		//首部冗余超过2^10或首部冗余超过实际使用
		q.Cap -= q.Begin
		q.End -= q.Begin
		tmp := make([]interface{}, q.Cap, q.Cap)
		copy(tmp, q.List[q.Begin:])
		q.List = tmp
		q.Begin = 0
	}
	q.Lock.Unlock()
	return val
}

// First 获取第一个元素
func (q *Queue) First() (e interface{}) {
	if q == nil {
		q = New()
		return nil
	}
	if q.QueueIsEmpty() {
		q.Clear()
		return nil
	}
	return q.List[q.Begin]
}

// Last 获取最后一个元素
func (q *Queue) Last() (e interface{}) {
	if q == nil {
		q = New()
		return nil
	}
	if q.QueueIsEmpty() {
		q.Clear()
		return nil
	}
	return q.List[q.End-1]
}
