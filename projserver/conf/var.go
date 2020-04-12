package conf

import "sync"

//DetectedItem Data structures distributed to agent
type DetectedItem struct {
	AppService string `json:"appservice"`
	AppName    string `json:"appname"`
	AppType    string `json:"apptype"`
	Module     string `json:"module"`
	ModType    string `json:"modtype"`
	ModVersion string `json:"modversion"`
	Host       string `json:"host"`
	Instance   string `json:"instance"`
	Env        string `json:"env"`
	Toggle     string `json:"toggle"`
}

// key: the ip of agent
// value: projects that require agent monitoring
type DetectedItemSafeMap struct {
	sync.RWMutex
	M map[string][]*DetectedItem
}

var (
	DetectedItemMap = &DetectedItemSafeMap{M: make(map[string][]*DetectedItem)}
)

func (this *DetectedItemSafeMap) Get(key string) ([]*DetectedItem, bool) {
	this.RLock()
	defer this.RUnlock()
	ipItem, exists := this.M[key]
	return ipItem, exists
}

func (this *DetectedItemSafeMap) Set(detectedItemMap map[string][]*DetectedItem) {
	this.Lock()
	defer this.Unlock()
	this.M = detectedItemMap
}

func (this *DetectedItemSafeMap) GetALL() map[string][]*DetectedItem {
	return this.M
}
