package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	_ "net/http/pprof"

	"gitee.com/feareless11/projmonitor/projagent/backend"
	"gitee.com/feareless11/projmonitor/projagent/conf"
	"gitee.com/feareless11/projmonitor/projagent/cron"
	"github.com/elastic/beats/libbeat/beat"
)

func prepare() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func init() {
	prepare()
	cfg := flag.String("c1", "cfg.json", "configuration file")
	version := flag.Bool("v1", false, "show version")
	help := flag.Bool("h1", false, "help")
	flag.Parse()

	handleVersion(*version)
	handleHelp(*help)
	// 解析配置文件
	handleConfig(*cfg)
	// 通过rpc的方式连接server端
	backend.InitClients(conf.Config.Web.Addrs)
	// 设置管道数量，控制检测项目时允许的并发连接数
	conf.Init()
	// 获取server端项目，初始化项目pid及检测项目pid，不存在拉起
	cron.Init()

}

func main() {
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6061", nil))
	}()
	// 检测项目的cpu、内存、out文件大小, 获取线程数、文件句柄数、tcp连接数
	go cron.StartCheck()
	// 检测jstat信息，同时录入数据库
	go cron.Gojstat()
	// 利用beat框架写入es
	err := beat.Run("javabeat", "", conf.New)
	if err != nil {
		os.Exit(1)
	}
}

func handleVersion(displayVersion bool) {
	if displayVersion {
		fmt.Println(conf.VERSION)
		os.Exit(0)
	}
}

func handleHelp(displayHelp bool) {
	if displayHelp {
		flag.Usage()
		os.Exit(0)
	}
}

func handleConfig(configFile string) {
	err := conf.Parse(configFile)
	if err != nil {
		log.Fatalln(err)
	}
}
