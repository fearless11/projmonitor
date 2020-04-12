package conf

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"github.com/toolkits/file"
)

type WebConf struct {
	Addrs    []string `json:"addrs"`
	Interval int      `json:"interval"`
	Timeout  int      `json:"timeout"`
}

type AlarmConf struct {
	Enable        bool    `json:"enable"`
	PercentofCPU  float64 `json:"cpu"`
	FreeMem       int     `json:"mem"`
	SizeOutFile   int64   `json:"outfilesize"`
	ChangeOutFile int64   `json:"outfilechange"`
	Alarmurl      string  `json:"alarmurl"`
}

type Jvm struct {
	Enable        bool    `json:"enable"`
	Interval      int     `json:"interval"`
	YGC           float64 `json:"ygc"`
	YGCT          float64 `json:"ygct"`
	FGC           float64 `json:"fgc"`
	FGCT          float64 `json:"fgct"`
	RatioGCT      float64 `json:"ratiogct"`
	SizeYGCOC     float64 `json:"sizeygcoc"`
	AvgSizeYGCOC  float64 `json:"avgsizeygcoc"`
	SizeOldOU     float64 `json:"sizeoldou"`
	RatioOldOU    float64 `json:"ratiooldou"`
	SizeAllHeap   float64 `json:"sizeallheap"`
	SizeUseHeap   float64 `json:"sizeuseheap"`
}

type Influxdb struct {
	Enable   bool   `josn:"enable"`
	Addr     string `json:"addr"`
	DataBase string `json:"database"`
}

type Elastic struct {
	Enable   bool   `josn:"enable"`
}

type GlobalConfig struct {
	Debug         bool       `json:"debug"`
	Hostname      string     `json:"hostname"`
	V1DIR         string     `json:"v1dir"`
	V4DIR         string     `json:"v4dir"`
	Worker        int        `json:"worker"`
	GOInterval    int        `json:"gointerval"`
	CheckInterval int        `json:"checkinterval"`
	AllocateMem   int        `json:"allocatemem"`
	ExecRestart   bool       `json:"execrestart"`
	Jvm           *Jvm       `json:"jvm"`
	Web           *WebConf   `json:"web"`
	Alarm         *AlarmConf `json:"alarm"`
	Influxdb      *Influxdb  `json:"influxdb"`
	Elastic       *Elastic    `json:"elastic"`
}

var (
	Config *GlobalConfig
)

func Hostname() (string, error) {
	hostname := Config.Hostname
	if hostname != "" {
		return hostname, nil
	}

	return os.Hostname()
}

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

	Config = &c

	log.Println("load configuration file", cfg, "successfully")
	return nil
}
