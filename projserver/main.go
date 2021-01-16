package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"

	"projmonitor/projserver/api"
	"projmonitor/projserver/conf"
	"projmonitor/projserver/cron"
	"projmonitor/projserver/model"
	"projmonitor/projserver/phttp"
)

func prepare() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func init() {
	prepare()
	cfg := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")
	help := flag.Bool("h", false, "help")
	flag.Parse()
	handleVersion(*version)
	handleHelp(*help)
	handleConfig(*cfg)
	model.InitMysql()
	cron.Init()
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()
	go api.Start()
	go phttp.Start()
	select {}
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
