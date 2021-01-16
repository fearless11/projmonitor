package cron

import (
	"log"
	"time"

	"projmonitor/projagent/backend"
	"projmonitor/projserver/api"

	"projmonitor/projagent/conf"
	webg "projmonitor/projserver/conf"
)

//GetProjInfo  Get project information every 5 minutes from server
func InitProjInfo() {
	t1 := time.NewTicker(time.Duration(conf.Config.Web.Interval) * time.Second)
	defer t1.Stop()

	for {
		// 从server端获取项目信息
		items, _ := GetItem()
		for _, item := range items {
			log.Println("[INFO] receive Item ", *item)		
			// 通过toggle开关判断项目是否需要监控 0:不监控 1:监控
			if item.Toggle == "0" {
				key := item.AppName + "-" + item.Module
				if _, ok := conf.DetectedItemMap.Get(key); ok {
					log.Println("[INFO] del pid cache", item)
					conf.DetectedItemMap.Del(key)
				}
				continue
			}
			conf.WorkerChan <- 1
			go initProjInfo(item)
		}
		<-t1.C
	}
}

func GetItem() ([]*webg.DetectedItem, error) {
	hostname, _ := conf.Hostname()

	var resp api.GetItemResponse
	err := backend.CallRpc("Web.GetItem", hostname, &resp)
	if err != nil {
		log.Println(err)
	}
	if resp.Message != "" {
		log.Println(resp.Message)
	}
	return resp.Data, err
}

//CheckTargetStatus  check pid,memory,outFile of java process
func initProjInfo(item *webg.DetectedItem) {
	defer func() {
		<-conf.WorkerChan
	}()

	key := item.AppName + "-" + item.Module

    // 项目信息处理到map,首次将全量新增，后续增量新增
	if _, ok := conf.DetectedItemMap.Get(key); !ok {
		checkProj := conf.NewCheckProject(item)
		checkProj.Flag = false
		checkProj.Inflx = false
		checkProj.Elast = false

		//goroutine timeout mechanism
		c1 := make(chan int)
		go checkProj.InitPidCache(c1)
		select {
		case <-c1:
		case <-time.After(time.Duration(conf.Config.GOInterval) * time.Second):
			if conf.Config.Debug {
				log.Println("[DEBUG] initcache", *checkProj, "timeout")
			}
		}
	}
}
