package conf

import "testing"

func Test_toAlarm(t *testing.T){
	item :=  &CheckResult{
		AppService:      "dw-server",
		AppName:         "dw-server",
		AppType:         "v1",
		Module:          "server",
		ModType:         "v1",
		ModVersion:      "v1",
		Host:            "127.0.0.1",
		Env: 			 "v1",
		Toggle:          "1",
		PrecentOfCPU:    "0",
		Memory:          "0",
		Pid:             "-1",
		OffsetOfOutFile: 0,
		FD:              " ",
		Thread:          " ",
		TCP:             " ",
		Jvm:             &CheckJvm{},
	}

	item.toAlarm("pid")
}
