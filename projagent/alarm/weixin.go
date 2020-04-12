package alarm

import (
	"fmt"
	"log"
	"strings"

	"github.com/astaxie/beego/httplib"
)

func Weixin(alarmURL, message string) {

	req := httplib.Post(alarmURL)
	req.Header("Content-Type", "application/json")
	req.Body(message)

	resq, err := req.Response()
	if err != nil {
		log.Println("[ERROR] Send alarm fail", err)
		return
	}
	contents, _ := req.Bytes()
	defer resq.Body.Close()

	fmt.Println(string(contents))
	if !strings.Contains(string(contents), "success") {
		log.Println("[ERROR] Alarm Wrong return")
		return
	}
}
