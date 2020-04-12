package cron

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"gitee.com/feareless11/projmonitor/projserver/alarm"
	"gitee.com/feareless11/projmonitor/projserver/conf"
	"gitee.com/feareless11/projmonitor/projserver/model"
)

type PullProject struct {
	AppService  string   `json:"appservice"`
	AppName     string   `josn:"appname"`
	AppType     string   `json:"apptype"`
	Module      string   `json:"module"`
	ModType     string   `json:"moduletype"`
	ModVersion  string   `json:"moduleversion"`
	Description string   `json:"description"`
	Owner       string   `json:"owner"`
	Instance    string   `json:"instance"`
	Env         string   `josn:"env"`
	Deploy      []string `json:"deploy"`
	Running     []string `json:"running"`
}

func pullProject() {
	client := &http.Client{}
	// req, err := http.NewRequest("GET", "http://192.168.3.73:1992/", nil)
	req, err := http.NewRequest("GET", conf.Config.Pull.Pullurl, nil)
	if err != nil {
		log.Print("[ERROR] httpNewRequest", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println("[ERROR]", err)
		return
	}
	Receive(resp)
}

func Receive(w *http.Response) (err error) {
	dec := json.NewDecoder(w.Body)
	defer w.Body.Close()

	_, err = dec.Token()
	if err != nil {
		log.Println("[ERROR] decoder token fail", err)
		return
	}
	for dec.More() {
		item := &PullProject{}
		if err = dec.Decode(&item); err != nil {
			log.Println(err)
			return
		}
		if conf.Config.Debug {
			log.Println("[DEBUG] pull project ", item)
		}
		compareProject(item)
	}
	return nil
}

func compareProject(item *PullProject) {
	recItem := &model.Project{
		AppService: item.AppService,
		AppName:    item.AppName,
		AppType:    item.AppType,
		Module:     item.Module,
		ModType:    item.ModType,
		ModVersion: item.ModVersion,
		Instance:   item.Instance,
		Env:        item.Env,
	}

	if item.AppName != "" && item.AppType != "" && item.Module != "" {

		//判断项目是否存在
		for _, deployHost := range item.Deploy {
			//部署信息为空
			if deployHost == "" {
				toAlarm(item, "NoDeploy")
				return
			}

			recItem.Host = deployHost
			recItem.Toggle = "1"

			//判断项目是否存在,默认不存在

			flag := true
			if _, ok := conf.DetectedItemMap.M[deployHost]; ok {
				for _, proj := range conf.DetectedItemMap.M[deployHost] {
					if recItem.AppName == proj.AppName && recItem.Module == proj.Module {
						flag = false
						break
					}
				}
			}

			//不存在
			if flag {
				if conf.Config.Debug {
					log.Println("[DEBUG] insert into project", recItem)
				}
				recItem.Add()
			}
		}

		//判断部署信息与运行信息是否一致
		for _, runningHost := range item.Running {
			flag := false
			for _, deployHost := range item.Deploy {
				if runningHost == deployHost {
					flag = true
					break
				}
			}
			//不一致
			if !flag {
				if conf.Config.Debug {
					log.Println("[WARN] Start alarm", item)
				}
				toAlarm(item, "inconsistent")
			}
		}
	}
}

func toAlarm(item *PullProject, flag string) {
	appName := item.AppName
	module := item.Module
	env := item.Env

	var alertname, message string
	switch flag {
	case "inconsistent":
		alertname = "服务" + appName + "模块" + module + "部署信息与运行信息不一致"
		message = fmt.Sprintf("[{   \"labels\":{ \"proj\":\"%v\",\"alertname\":\"%v\",\"env\":\"%v\"},\"annotations\":{\"DeployHost\":\"%v\",\"RunningHost\":\"%v\"}} ]", appName, alertname, env, item.Deploy, item.Running)
	case "NoDeploy":
		alertname = "服务" + appName + "模块" + module + "没有部署主机"
	}

	alarmConf := conf.Config.Alarm
	if alarmConf.Enable {
		alarmURL := alarmConf.Alarmurl
		log.Println("[WARN] Start alarm", alarmURL, alertname, message)
		go alarm.Weixin(alarmURL, alertname, message)
	}
}
