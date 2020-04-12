package alarm

import (
	"fmt"
	"testing"
)

func Test_alarmofRestart(t *testing.T) {
	alarmURL := "http://alarm.we.com/api/v1/alerts"
	appName := "test"
	module := "abc"
	host := "1.1.1.1"
	env := "test"
	alertname := "服务" + appName + "的" + module + "模块pid不存在restart"
	message := fmt.Sprintf("[{   \"labels\":{ \"proj\":\"%v\",\"alertname\":\"%v\",\"host\":\"%v\",\"env\":\"%v\"}} ]", appName, alertname, host, env)
	weixin(alarmURL, alertname, message)
}
