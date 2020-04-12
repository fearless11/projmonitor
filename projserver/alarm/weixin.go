package alarm

import (
	"log"
	"strings"

	"github.com/astaxie/beego/httplib"
)

func Weixin(alarmURL, alertname, message string) {
	req := httplib.Post(alarmURL)
	req.Header("Content-Type", "application/json")
	req.Body(message)
	resq, err := req.Response()
	if err != nil {
		log.Println("[ERROR] Send alarm fail", err)
	}
	contents, _ := req.Bytes()
	if !strings.Contains(string(contents), "success") {
		log.Println("[ERROR] Alarm Wrong return")
	}
	defer resq.Body.Close()
}
