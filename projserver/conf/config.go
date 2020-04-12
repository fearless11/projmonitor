package conf

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/toolkits/file"
)

type AlarmConf struct {
	Enable   bool   `json:"enable"`
	Alarmurl string `json:"alarmurl"`
}

type PullConf struct {
	Enable  bool   `json:"enable"`
	Pullurl string `json:"pullurl"`
}

type MysqlConfig struct {
	Addr     string `json:"addr"`
	Interval int    `json:"interval"`
	Idle     int    `json:"idle"`
	Max      int    `json:"max"`
}

type RpcConfig struct {
	Listen string `json:"listen"`
}

type GlobalConfig struct {
	Debug bool         `json:"debug"`
	Rpc   *RpcConfig   `json:"rpc"`
	Mysql *MysqlConfig `json:"mysql"`
	Alarm *AlarmConf   `json:"alarm"`
	Pull  *PullConf    `json:"pull"`
	Web   string       `json:"web"`
}

var (
	Config     *GlobalConfig
	configLock = new(sync.RWMutex)
)

func Parse(cfg string) error {
	if cfg == "" {
		return fmt.Errorf("use -c to specify configuration file")
	}

	if !file.IsExist(cfg) {
		return fmt.Errorf("configuration file %s is nonexistent", cfg)
	}

	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		return fmt.Errorf("read configuration file %s fail %s", cfg, err.Error())
	}

	var c GlobalConfig
	err = json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		return fmt.Errorf("parse configuration file %s fail %s", cfg, err.Error())
	}

	configLock.Lock()
	defer configLock.Unlock()
	Config = &c

	log.Println("load configuration file", cfg, "successfully")
	return nil
}
