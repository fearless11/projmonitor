package cron

import (
	"log"
	"time"

	"gitee.com/feareless11/projmonitor/projagent/conf"
)

// StartCheck Regular inspection of java project
func StartCheck() {
	t1 := time.NewTicker(time.Duration(conf.Config.CheckInterval) * time.Second)
	defer t1.Stop()
	for {
		for _, item := range conf.DetectedItemMap.GetAll() {
			// control concurrent threads through buffer channels
			conf.WorkerChan <- 1
			if conf.Config.Debug {
				log.Println("[DEBUG] read pid cache", item)
			}
			go checkTargetStatus(item)
		}
		<-t1.C
	}
}

// CheckTargetStatus  check pid,memory,outFile of java process
func checkTargetStatus(item *conf.CheckResult) {
	defer func() {
		<-conf.WorkerChan
	}()

	//goroutine timeout mechanism : default 1000ms
	c1 := make(chan int)
	go item.CheckAll(c1)
	select {
	case <-c1:
	case <-time.After(time.Duration(conf.Config.GOInterval) * time.Second):
		log.Println("[WARN] item.CheckAll timeout:", item)
	}
}

//Gojstat check java information for jstat
func Gojstat() {
	t1 := time.NewTicker(time.Duration(conf.Config.Jvm.Interval) * time.Second)
	defer t1.Stop()

	for {
		for _, item := range conf.DetectedItemMap.GetAll() {
			if item.Toggle == "0" {
				continue
			}
			c2 := make(chan int)
			// 检测项目jstat的信息，同时录入数据库，influxdb或者es
			go item.CheckJvm(c2)
			select {
			case <-c2:
			case <-time.After(time.Duration(conf.Config.GOInterval) * time.Second):
				log.Println("[WARN] checkJvm", item, "timeout")
			}
		}
		<-t1.C
	}
}
