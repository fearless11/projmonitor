package conf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"gitee.com/feareless11/projmonitor/projagent/alarm"
	webg "gitee.com/feareless11/projmonitor/projserver/conf"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/publisher"
	"github.com/shirou/gopsutil/process"
)

const (
	ScModule = "sc"
)

//CheckResultItem  Data structures distributed to alarm
type CheckResult struct {
	AppService      string `json:"appservice"`
	AppName         string `json:"appname"`
	AppType         string `json:"apptype"`
	Module          string `json:"module"`
	ModType         string `json:"modtype"`
	ModVersion      string `json:"modversion"`
	Host            string `json:"host"`
	Env             string `json:"env"`
	Pid             string `json:"pid"`
	PrecentOfCPU    string `json:"percentofcpu"`
	Memory          string `json:"memory"`
	OffsetOfOutFile int64  `json:"offsetofoutFile"`
	ErrMsgOutFile   string `json:"errmsgoutFile"`
	Toggle          string `json:"toggle"`
	FD              string `json:"fd"`
	Thread          string `json:"thread"`
	TCP             string `json:"tcp"`
	Jvm             *CheckJvm
	Flag            bool `json:"flag"`
	Inflx           bool `json:"inflx"`
	Elast           bool `json:"elast"`
	client          publisher.Client
	done            chan struct{}
}

type CheckJvm struct {
	//compare the value of the last change
	SizeYGC      float64 `json:"SizeYGC"`
	SizeYGCT     float64 `json:"SizeYGCT"`
	SizeFGC      float64 `json:"SizeFGC"`
	SizeFGCT     float64 `json:"SizeFGCT"`
	SizeYGCOC    float64 `json:"SizeYGCOC"`
	AvgSizeYGCOC float64 `json:"AvgSizeYGCOC"`
	RatioGC      float64 `json:"RatioGC"`
	SizeOldOU    float64 `json:"SizeOldOU"`
	RatioOldOU   float64 `json:"RatioOldOU"`
	SizeAllHeap  float64 `json:"SizeHeap"`
	SizeUseHeap  float64 `json:"SizeUseHeap"`

	//the value of the this test
	S0C  float64 `json:"S0C"`
	S1C  float64 `json:"S1C"`
	S0U  float64 `json:"S0U"`
	S1U  float64 `json:"S1U"`
	EC   float64 `json:"EC"`
	EU   float64 `json:"EU"`
	OC   float64 `json:"OC"`
	OU   float64 `json:"OU"`
	YGC  float64 `json:"YGC"`
	YGCT float64 `json:"YGCT"`
	FGC  float64 `json:"FGC"`
	FGCT float64 `json:"FGCT"`
}

var (
	CheckToggle chan int
	CheckMsg    chan *CheckResult
)

func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	CheckToggle = make(chan int)
	CheckMsg = make(chan *CheckResult)
	bt := &CheckResult{
		done: make(chan struct{}),
	}
	return bt, nil
}

func (bt *CheckResult) Run(b *beat.Beat) error {
	// logp.Info("parsebeat is running! Hit CTRL-C to stop it.")
	bt.client = b.Publisher.Connect()
	for {
		select {
		case <-bt.done:
			return nil
		case <-CheckToggle:
		}
		proj := <-CheckMsg
		if len(proj.AppName) != 0 {
			bt.publishItem(proj)
		}
	}
}

func (bt *CheckResult) publishItem(proj *CheckResult) {
	event := common.MapStr{
		"@timestamp": common.Time(time.Now().UTC()),
		"type":       "javabeat",
		"appservice": proj.AppService,
		"appname":    proj.AppName,
		"module":     proj.Module,
		"host":       proj.Host,
		"env":        proj.Env,
		"cpu":        StringTofloat(proj.PrecentOfCPU),
		"mem":        StringTofloat(proj.Memory),
		"fd":         StringTofloat(proj.FD),
		"thread":     StringTofloat(proj.Thread),
		"tcp":        StringTofloat(proj.TCP),
		"ygc":        proj.Jvm.SizeYGC,
		"ygct":       proj.Jvm.SizeYGCT,
		"fgc":        proj.Jvm.SizeFGC,
		"fgct":       proj.Jvm.SizeFGCT,
		"sizeygc":    proj.Jvm.SizeYGCOC,
		"avgsizeygc": proj.Jvm.AvgSizeYGCOC,
		"gcratio":    proj.Jvm.RatioGC,
		"ou":         proj.Jvm.SizeOldOU,
		"ouratio":    proj.Jvm.RatioOldOU,
		"heaptotal":  proj.Jvm.SizeAllHeap,
		"heapuse":    proj.Jvm.SizeUseHeap,
	}
	bt.client.PublishEvent(event)
}

func (item *CheckResult) connectbeat() {
	// 通过通道传递数据完成项目信息录入
	CheckToggle <- 1
	CheckMsg <- item
}

func (bt *CheckResult) Stop() {
	close(bt.done)
	bt.client.Close()
}

func NewCheckProject(item *webg.DetectedItem) *CheckResult {
	itemCheckResult := &CheckResult{
		AppService:      item.AppService,
		AppName:         item.AppName,
		AppType:         item.AppType,
		Module:          item.Module,
		ModType:         item.ModType,
		ModVersion:      item.ModVersion,
		Host:            item.Host,
		Env:             item.Env,
		Toggle:          item.Toggle,
		PrecentOfCPU:    "0",
		Memory:          "0",
		Pid:             "-1",
		OffsetOfOutFile: 0,
		FD:              " ",
		Thread:          " ",
		TCP:             " ",
		Jvm:             &CheckJvm{},
	}
	return itemCheckResult
}

func (item *CheckResult) InitPidCache(c1 chan int) {
	item.checkPid()
	c1 <- 1
}

func (item *CheckResult) CheckAll(c1 chan int) {
	// 检测项目cpu、内存信息，首次检测出错，先校验pid，
	// 若pid存在，再次检测cpu、内存。返回err不在后续检测
	err := item.memCPUByPid()
	if err != nil {
		_, err := item.pidByPgrep()
		if err != nil {
			log.Println("[ERROR] pidByPgrep:", err)
			go item.toAlarm("pid")
			c1 <- 1
			return
		}
		err = item.memCPUByPid()
		if err != nil {
			log.Println("[ERROR] memCPUByPid:", err)
			c1 <- 1
			return
		}
	}

	// 检测项目的out文件大小
	err = item.outFileByPid()
	if err != nil && err.Error() != "0" {
		item.ErrMsgOutFile = err.Error()
		go item.toAlarm("outfile")
	}

	// 获取线程数、文件句柄数、tcp连接数
	item.threadByProc()
	item.fdByProc()
	item.tcpByProc()
	c1 <- 1
}

func (item *CheckResult) checkPid() error {
	// 获取项目pid机制： pid文件--> pgrep
	err := item.pidByFile()
	if err != nil {
		_, err1 := item.pidByPgrep()
		if err1 != nil {
			go item.toAlarm("pid")
			return err1
		}
	}

	procPid := "/proc/" + item.Pid
	_, err = os.Stat(procPid)
	if err != nil {
		_, err1 := item.pidByPgrep()
		if err1 != nil {
			go item.toAlarm("pid")
			return err1
		}
	}

	key := item.AppName + "-" + item.Module
	DetectedItemMap.Set(key, item)
	return nil
}

func (item *CheckResult) pidByFile() error {
	appname := item.AppName
	module := item.Module

	pidFile := fmt.Sprintf("%v/%v/%v/var/run/%v-%v.pid", Config.V4DIR, appname, module, appname, module)
	if item.AppType == "v1" {
		pidFile = fmt.Sprintf("%v/var/run/%v-%v.pid", Config.V1DIR, appname, module)
	}

	f, err := ioutil.ReadFile(pidFile)
	if err != nil {
		log.Println("[ERROR]", err)
		return err
	}

	p := fmt.Sprintf("%s", string(f))
	item.Pid = strings.Trim(p, "\n")
	return nil
}

func (item *CheckResult) pidByPgrep() (string, error) {
	var args string

	args = fmt.Sprintf("pgrep -f  Djava.apps.prog=%v-%v", item.AppName, item.Module)
	if item.Module == ScModule {
		args = fmt.Sprintf("pgrep -f  Djava.apps.prog=%v", item.AppName)
	}

	oldPid := item.Pid
	p, err := execShell(args)
	if err != nil {
		return "", err
	}
	item.Pid = p
	return oldPid, nil
}

func execShell(args string) (string, error) {
	var out bytes.Buffer
	cmd := exec.Command("/bin/sh", "-c", args)
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return strings.Trim(out.String(), "\n"), nil
}

func (item *CheckResult) threadByProc() {
	args := fmt.Sprintf("ls /proc/%v/task | wc -l", item.Pid)
	t, err := execShell(args)
	if err != nil {
		return
	}
	item.Thread = t
}

func (item *CheckResult) fdByProc() {
	args := fmt.Sprintf("ls /proc/%v/fd | wc -l", item.Pid)
	f, err := execShell(args)
	if err != nil {
		return
	}
	item.FD = f
}

func (item *CheckResult) tcpByProc() {
	args := fmt.Sprintf("ls -l /proc/%v/fd |grep -i socket |wc -l", item.Pid)
	t, err := execShell(args)
	if err != nil {
		return
	}
	item.TCP = t
}

func (item *CheckResult) memCPUByPid() error {
	p, err := strconv.Atoi(item.Pid)
	if err != nil {
		return err
	}

	pro, err := process.NewProcess(int32(p))
	if err != nil {
		return err
	}

	cpuPercent, _ := pro.CPUPercent()
	var memory uint64
	mem, _ := pro.MemoryInfo()
	if mem != nil {
		memory = mem.RSS / (1 << 20) // bytes to MB
	}

	item.PrecentOfCPU = fmt.Sprintf("%.2f", cpuPercent)
	item.Memory = fmt.Sprintf("%v", memory)

	freeMemAlarm := Config.Alarm.FreeMem
	freeMemCurrent := math.Abs(float64(Config.AllocateMem) - float64(memory))
	percentOfCPUAlarm := Config.Alarm.PercentofCPU

	if cpuPercent > percentOfCPUAlarm {
		go item.toAlarm("cpu")
	}
	if int(freeMemCurrent) < freeMemAlarm {
		go item.toAlarm("memory")
	}
	return nil
}

func (item *CheckResult) initOffset(sizeOfOut int64) (int64, error) {
	offset := &item.OffsetOfOutFile

	//检查文件是否过大
	if sizeOfOut > Config.Alarm.SizeOutFile {
		return 0, fmt.Errorf("%v-%v.out文件%vM,超过告警阈值%vM", item.AppName, item.Module, sizeOfOut, Config.Alarm.SizeOutFile)
	}

	//检查文件是否增长过快
	changeSize := sizeOfOut - *offset
	//第一次检查out文件 或者 服务重启后out文件变小,将偏移置为文件末尾
	if *offset == 0 || changeSize < 0 {
		*offset = sizeOfOut
		return 0, fmt.Errorf("0")
	}

	sizeAlarm := Config.Alarm.ChangeOutFile
	if changeSize > int64(sizeAlarm) {
		return 0, fmt.Errorf("%v-%v.out在%v分钟内增加了%vM,超过增长阈值%vM", item.AppName, item.Module, Config.CheckInterval, changeSize)
	}

	//如果out文件变化大于3M,一次最多检查3M内容
	if changeSize > 3 {
		return 3, nil
	}

	return changeSize, nil
}

func (item *CheckResult) outFileByPid() error {
	name := fmt.Sprintf("/proc/%v/fd/1", item.Pid)

	flag := os.O_RDONLY
	perm := os.FileMode(0)
	f, err := os.OpenFile(name, flag, perm)
	defer f.Close()
	if err != nil {
		log.Println("[ERROR]", err)
		return nil
	}

	fi, err := f.Stat()
	if err != nil {
		return nil
	}

	sizeOfOut := fi.Size() / (1 << 20)
	bufSize, err := item.initOffset(sizeOfOut)
	if err != nil {
		return err
	}

	offset := &item.OffsetOfOutFile
	buf := make([]byte, bufSize)

	//Read the latest file contents
	_, err = f.Seek(*offset, os.SEEK_SET)
	if err != nil {
		log.Println("[ERROR] seek", err)
		return nil
	}

	n, err := f.Read(buf)
	if err != nil {
		log.Println("[ERROR] read", err)
		return nil
	}

	*offset = *offset + int64(n)
	err = Toline(string(buf))
	if err != nil {
		return err
	}
	return nil
}

//Toline  A line of parsing file contents
func Toline(data string) error {
	for _, line := range strings.Split(string(data), "\n") {
		//ERROR keys: ERROR 、.....
		if strings.Contains(line, "ERROR") || strings.Contains(line, "memory") {
			log.Println("[ERROR] out error msg：", line)
			return fmt.Errorf("%v", line)
		}
	}
	return nil
}

func (item *CheckResult) confirmPid() bool {
	pid, err := item.pidByPgrep()
	if err != nil {
		return false
	}
	if item.Pid != pid {
		return true
	}
	return false
}

func (item *CheckResult) restartProject() {
	appname := item.AppName
	module := item.Module

	args := fmt.Sprintf("%v/%v/%v/control.sh restart", Config.V4DIR, appname, module)
	if item.AppType == "v1" {
		args = fmt.Sprintf("su - www -c '%v/apps/%v/bin/%v restart'", Config.V1DIR, appname, module)
	}

	log.Println("[WARN] restartProject:", args)

	if Config.ExecRestart {
		_, err := execShell(args)
		if err != nil {
			log.Printf("[ERROR] %v restart fail%v\n", args, err)
		}
	}
}

func (item *CheckResult) CheckJvm(c2 chan int) {

	// 检测项目jvm信息，通过jstat获取数据
	if Config.Jvm.Enable {
		err := item.checkJvm()
		if err != nil {
			if err.Error() == "invalidresultArr" {
				c2 <- 1
				log.Println("[ERROR] checkjvm", err, item, *item.Jvm)
				return
			}
			go item.toAlarm("jvm")
		}
	}

	// jvm信息录入数据库
	if Config.Influxdb.Enable {
		//保证第一次检查的数据不录入数据库
		if !item.Inflx {
			item.Inflx = true
			c2 <- 1
			return
		}
		go writeToInlfux(item)
	}

	if Config.Elastic.Enable {
		//保证第一次检查的数据不录入es
		if !item.Elast {
			item.Elast = true
			c2 <- 1
			return
		}
		item.connectbeat()
	}
	c2 <- 1
}
func (item *CheckResult) checkJvm() error {

	jvm, err := item.jvmByJstat()
	if err != nil {
		return err
	}

	changeSizeYGC := jvm.YGC - item.Jvm.YGC
	if changeSizeYGC < 0 {
		item.Jvm = &CheckJvm{}
		item.Inflx = false
		item.Elast = false
		return nil
	}

	item.Jvm.SizeYGC = jvm.YGC - item.Jvm.YGC
	item.Jvm.SizeYGCT = round(jvm.YGCT-item.Jvm.YGCT, 6)
	item.Jvm.SizeFGC = jvm.FGC - item.Jvm.FGC
	item.Jvm.SizeFGCT = round(jvm.FGCT-item.Jvm.FGCT, 6)

	item.Jvm.RatioGC = round(jvm.RatioGC, 6) //本次检查
	item.Jvm.SizeAllHeap = jvm.S0C + jvm.S1C + jvm.EC + jvm.OC
	item.Jvm.SizeUseHeap = jvm.S0U + jvm.S1U + jvm.EU + jvm.OU

	item.Jvm.SizeYGCOC = jvm.OC + jvm.S0C + jvm.S1C - item.Jvm.OC - item.Jvm.S0C - item.Jvm.S1C
	item.Jvm.AvgSizeYGCOC = 0.0
	if item.Jvm.SizeYGCOC != 0.0 && item.Jvm.SizeYGC > 0.0 {
		item.Jvm.AvgSizeYGCOC = round(item.Jvm.SizeYGCOC/item.Jvm.SizeYGC, 6)
	}
	item.Jvm.SizeOldOU = jvm.OU          //本次检查
	item.Jvm.RatioOldOU = jvm.RatioOldOU //本次检查

	item.Jvm.YGC = jvm.YGC
	item.Jvm.YGCT = jvm.YGCT
	item.Jvm.FGC = jvm.FGC
	item.Jvm.FGCT = jvm.FGCT
	item.Jvm.S0C = jvm.S0C
	item.Jvm.S0U = jvm.S0U
	item.Jvm.S1C = jvm.S1C
	item.Jvm.S1U = jvm.S1U
	item.Jvm.EC = jvm.EC
	item.Jvm.EU = jvm.EU
	item.Jvm.OC = jvm.OC
	item.Jvm.OU = jvm.OU

	if !item.Flag {
		item.Flag = true
		return nil
	}

	if item.Jvm.SizeYGC > Config.Jvm.YGC || item.Jvm.SizeYGCT > Config.Jvm.YGCT || item.Jvm.SizeFGC > Config.Jvm.FGC || item.Jvm.SizeFGCT > Config.Jvm.FGCT || item.Jvm.SizeYGCOC > Config.Jvm.SizeYGCOC || item.Jvm.AvgSizeYGCOC > Config.Jvm.AvgSizeYGCOC || item.Jvm.RatioGC > Config.Jvm.RatioGCT || item.Jvm.SizeOldOU > Config.Jvm.SizeOldOU || item.Jvm.RatioOldOU > Config.Jvm.RatioOldOU || item.Jvm.SizeAllHeap > Config.Jvm.SizeAllHeap || item.Jvm.SizeUseHeap > Config.Jvm.SizeUseHeap {
		log.Println("[WARN] jstat检测到jvm值超过阈值", *item, *item.Jvm)
		return fmt.Errorf("jvmover")
	}
	return nil
}

func (item *CheckResult) jvmByJstat() (*CheckJvm, error) {

	pid := item.Pid
	jvm := &CheckJvm{}
	// args := "jstat -gc -t " + pid + Config.Jvm.JstatInterval + Config.Jvm.JstatNum + "| grep -v OC | tr -s ' '"
	args := fmt.Sprintf("jstat -gc -t %v %v %v | grep -v OC | tr -s ' '", pid, 1, 3)
	cmd := exec.Command("/bin/sh", "-c", args)
	var out bytes.Buffer
	cmd.Stdout = &out
	var outerr bytes.Buffer
	cmd.Stderr = &outerr
	err := cmd.Run()
	if err != nil {
		log.Println("[ERROR] jstat -gc", pid)
		return jvm, err
	}

	resultArr := strings.Split(out.String(), "\n")

	if len(resultArr) != 4 {
		log.Println("[ERROR] checkjvm invalidresultArr", len(resultArr), resultArr, out.String(), args, *item, outerr.String())
		return jvm, fmt.Errorf("invalidresultArr")
	}

	jvm.S0C = calculate(resultArr, 2)
	jvm.S1C = calculate(resultArr, 3)
	jvm.S0U = calculate(resultArr, 4)
	jvm.S1U = calculate(resultArr, 5)
	jvm.EC = calculate(resultArr, 6)
	jvm.EU = calculate(resultArr, 7)
	jvm.OC = calculate(resultArr, 8)
	jvm.OU = calculate(resultArr, 9)
	jvm.YGC = calculate(resultArr, 14)
	jvm.YGCT = calculate(resultArr, 15)
	jvm.FGC = calculate(resultArr, 16)
	jvm.FGCT = calculate(resultArr, 17)

	jvm.RatioGC = calculate(resultArr, 18) / calculate(resultArr, 1)
	jvm.RatioOldOU = round(jvm.OU/jvm.OC, 3)

	return jvm, nil
}

func calculate(resultArr []string, p int) float64 {
	oneR := strings.Split(resultArr[0], " ")
	twoR := strings.Split(resultArr[1], " ")
	threeR := strings.Split(resultArr[2], " ")

	return round((stf(oneR[p])+stf(twoR[p])+stf(threeR[p]))/3, 3)
}

//round float保留小数位
func round(f float64, n int) float64 {
	pow10N := math.Pow10(n)
	return math.Trunc((f+0.5/pow10N)*pow10N) / pow10N
}

//stf string to float
func stf(str string) float64 {
	s, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0.0
	}
	return s
}

func (item *CheckResult) toAlarm(flag string) {
	appName := item.AppName
	module := item.Module
	host := item.Host

	// 验证pid，pid不存在告警，pid改变则return
	if item.confirmPid() {
		return
	}

	//	if flag == "pid" || flag == "cpu" || flag == "memory" {
	//		item.restartProject()
	//	}

	if flag == "pid" {
		item.restartProject()
	}

	var message, msg string
	var alertname string = "JAVAProjProblem"
	var lv string = "CRITICAL"
	var t = time.Now().Format("2006-01-02 15:04:05")
	switch flag {
	case "pid":
		msg = host + " 服务" + appName + "的" + module + "模块pid不存在"
	case "cpu":
		msg = host + " 服务" + appName + "的" + module + "模块CPU " + item.PrecentOfCPU + "超过阈值" + fmt.Sprintf("%v", Config.Alarm.PercentofCPU)
	case "memory":
		msg = host + " 服务" + appName + "的" + module + "模块Memory" + item.Memory + "过高"
	case "outfile":
		msg = host + " 服务" + appName + "的" + module + "的out文件中检测到error" + " 内容: " + item.ErrMsgOutFile
	case "jvm":
		data, _ := json.Marshal(*item.Jvm)
		jvmInfo := string(data)
		msg = host + " 检测到服务" + appName + "的" + module + "的jstat值超过阈值" + " 内容: " + jvmInfo
	}

	key := item.AppName + "-" + item.Module
	DetectedItemMap.Del(key)

	message = fmt.Sprintf(`[{   
		"labels":{ 
			"alertname":"%v",
			"tag":"%v",
			"level":"%v"
		},
		"annotations":{
			"group":"JAVA",
			"time":"%v",
			"msg":"%v"
		}
	}]`, alertname, appName, lv, t, msg)

	if Config.Debug {
		log.Println("[WARN] Start alarm", message)
	}

	if Config.Alarm.Enable {
		alarmURL := Config.Alarm.Alarmurl
		alarm.Weixin(alarmURL, message)
		log.Println("[WARN] Start alarm", alarmURL, message)
	}
}
