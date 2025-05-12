package util

import "sync"

var argumentsLock = sync.Map{}

// ArgumentsNewLock 创建变量锁
func ArgumentsNewLock(executeId string) {
	argumentsLock.Store(executeId, sync.RWMutex{})
}

// ArgumentsGetLock 获取读写锁
func ArgumentsGetLock(executeId string) sync.RWMutex {
	m, _ := argumentsLock.Load(executeId)
	return m.(sync.RWMutex)
}

// ArgumentsRLock 读上锁
func ArgumentsRLock(executeId string) {
	mutex := ArgumentsGetLock(executeId)
	mutex.RLock()
	argumentsLock.Store(executeId, mutex)
}

// ArgumentsRUnLock 读释放锁
func ArgumentsRUnLock(executeId string) {
	mutex := ArgumentsGetLock(executeId)
	mutex.RUnlock()
	argumentsLock.Store(executeId, mutex)
}

// ArgumentsLock 写上锁
func ArgumentsLock(executeId string) {
	mutex := ArgumentsGetLock(executeId)
	mutex.Lock()
	argumentsLock.Store(executeId, mutex)
}

// ArgumentsUnLock 写释放锁
func ArgumentsUnLock(executeId string) {
	mutex := ArgumentsGetLock(executeId)
	mutex.Unlock()
	argumentsLock.Store(executeId, mutex)
}

// ArgumentsGet 获取arguments值
func ArgumentsGet(arguments *sync.Map, key string) interface{} {
	v, _ := arguments.Load(key)
	return v
}

// ArgumentsSet 设置arguments值
func ArgumentsSet(arguments *sync.Map, key string, val interface{}) {
	arguments.Store(key, val)
}
