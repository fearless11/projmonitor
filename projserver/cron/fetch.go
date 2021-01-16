package cron

import (
	"log"
	"time"

	"projmonitor/projserver/conf"
	"projmonitor/projserver/model"
)

//GetProjectItem Get project information every 5 minutes from database
// pull the project after 5s
func GetProjectItem() {
	t1 := time.NewTicker(time.Duration(conf.Config.Mysql.Interval) * time.Second)
	defer t1.Stop()

	for {
		getProjectItem()
		if conf.Config.Pull.Enable {
			time.Sleep(time.Duration(120) * time.Second)
			pullProject()
		}
		<-t1.C
	}
}

// getProjectItem Get project information from database
func getProjectItem() {
	detectedItemMap := make(map[string][]*conf.DetectedItem)
	projs, err := model.GetAllProjectByCron()
	if err != nil {
		log.Println("[ERROR] get all project by cron:", err)
		return
	}

	for _, p := range projs {
		key := p.Host
		detectedItem := newDetectedItem(p)
		if _, exists := detectedItemMap[key]; exists {
			detectedItemMap[key] = append(detectedItemMap[key], detectedItem)
		} else {
			detectedItemMap[key] = []*conf.DetectedItem{detectedItem}
		}
		if conf.Config.Debug {
			log.Println("[DEBUG] detected project", *detectedItem)
		}
	}
	conf.DetectedItemMap.Set(detectedItemMap)
}

//newDetectedItem key==agent host ip
func newDetectedItem(p *model.Project) *conf.DetectedItem {
	detectedItem := &conf.DetectedItem{
		AppService: p.AppService,
		AppName:    p.AppName,
		AppType:    p.AppType,
		Module:     p.Module,
		ModType:    p.ModType,
		ModVersion: p.ModVersion,
		Host:       p.Host,
		Instance:   p.Instance,
		Env:        p.Env,
		Toggle:     p.Toggle,
	}
	return detectedItem
}
