### **项目目录结构**
```
projmonitor
├── projagent
│   ├── alarm
│   │   ├── weixin.go
│   │   └── weixin_test.go
│   ├── backend
│   │   └── rpc.go
│   ├── cfg.json
│   ├── checkprojagent.sh
│   ├── conf
│   │   ├── config.go
│   │   ├── const.go
│   │   ├── project.go
│   │   └── var.go
│   ├── control
│   ├── cron
│   │   ├── check.go
│   │   ├── fetch.go
│   │   └── init.go
│   ├── main.go
│   ├── README.md
├── projserver
│   ├── alarm
│   │   └── weixin.go
│   ├── api
│   │   ├── item.go
│   │   └── rpc.go
│   ├── cfg.default.json
│   ├── cfg.json
│   ├── conf
│   │   ├── config.go
│   │   ├── const.go
│   │   └── var.go
│   ├── control
│   ├── cron
│   │   ├── fetch.go
│   │   ├── init.go
│   │   ├── pull.go
│   │   └── pull_test.go
│   ├── main.go
│   ├── model
│   │   ├── mysql.go
│   │   └── project.go
│   ├── phttp
│   │   └── http.go
│   ├── README.md
└── README.md

```

### **功能介绍**
#### 【Server】
##### 功能：
1. 负责分发project信息，包括服务名、服务版本等
2. 获取janson的项目数据
3. 发送告警

##### 配置文件：
```
    {
        "debug": true,
        "rpc": {
            "listen": "0.0.0.0:1990"     //开发端口，允许agent连接
        },
        "mysql": {
            "addr": "root:root@tcp(192.168.1.60:3306)/project_monitor?charset=utf8&&loc=Asia%2FShanghai",
            "interval": 300,             //定期每隔5分钟获取project信息，单位ms
            "idle": 10,
            "max": 20
        }
        "alarm":{
           "enable": false,             //启用告警，用于场景：当运行信息与部署信息不一致
           "alarmurl": "http://alarm.we.com/api/v1/alerts"
        },
        "pull":{
            "enable": true,            //获取janson的项目数据
            "pullurl":"http://192.168.3.189:9001/api/v1/getallproject"
        },
        "web":"0.0.0.0:8080"
    }
```

#### 【Agent】
##### 功能：
1. 检查project的pid信息
2. 检查project的CPU百分比
3. 检查project物理Memory使用值
4. 检查out文件大小、以及其中出现的错误
5. 发送告警

##### 配置文件：

```
   {
        "debug": true,
        "hostname": "192.168.3.73",
        "v1dir":"/usr/local/java",         //老服务部署目录
        "v4dir":"/usr/local/release",      //新服务部署目录
        "worker": 100,                     //工作队列长度
        "gointerval":1000,                 //单个goroutine超时时间，单位ms
        "checkinterval":120,               //定期每隔2分钟检查一次project，单位s
        "allocatemem":2000,                //设定服务被分配的内存值，默认分配2G内存
        "web": {
            "addrs": ["127.0.0.1:1990"],   //服务端的ip:port
            "interval": 300,               //定期每隔5分钟检测服务端的project信息，单位s
            "timeout": 1000
        },
        "alarm":{
            "enable": false,
            "cpu":100,                    //cpu占用百分比 > 100 时触发告警
            "mem":100,                    //内存剩余值 < 100M 时触发告警
            "outfilesize":10,             //out文件大于10M 触发告警
            "outfilechange":3,            //2分钟内out文件增加大于3M 触发告警
            "alarmurl": "http://alarm.we.com/api/v1/alerts"   //告警接口
        }
    }
```
