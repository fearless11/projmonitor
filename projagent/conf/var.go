package conf

import "sync"

// var CheckResultQueue *list.SafeLinkedList
var WorkerChan chan int


func Init() {
	WorkerChan = make(chan int, Config.Worker)
}

// key:appname-module ,value: CheckResult
type DetectedItemSafeMap struct {
	sync.RWMutex
	M map[string]*CheckResult
}

var (
	DetectedItemMap = &DetectedItemSafeMap{M: make(map[string]*CheckResult)}
)

func (D *DetectedItemSafeMap) Set(key string, CheckResult *CheckResult) {
	D.Lock()
	D.M[key] = CheckResult
	D.Unlock()
}

func (D *DetectedItemSafeMap) Get(key string) (*CheckResult, bool) {
	D.RLock()
	item, exists := D.M[key]
	D.RUnlock()
	return item, exists
}

func (D *DetectedItemSafeMap) Del(key string) {
	D.Lock()
	delete(D.M, key)
	D.Unlock()
}

func (D *DetectedItemSafeMap) GetAll() []*CheckResult {
	D.RLock()
	vals := make([]*CheckResult, 0)
	for _, val := range D.M {
		vals = append(vals, val)
	}
	D.RUnlock()
	return vals
}
