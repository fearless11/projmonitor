#### 【server】
1. 定期每5分钟读取数据库项目信息
2. 定期每7分钟拉取janson的数据信息
3. 比较部署信息与运行信息，不一致异步告警通知

##### 接口
```
http://localhost:8080/
http://localhost:8080/v1/proj
http://localhost:8080/v1/projs
```


##### [example]
###### 单条
```
curl -XPOST  -H"Content-Type":"application/json" http://localhost:8080/v1/proj -d
'{ 
    "app_service":"com.hello", 
    "app_name":"hello-server", 
    "app_type":"v1","module":"1", 
    "mod_type":"1",  
    "mod_version":"-1", 
    "host":"192.168.3.73",     
    "instance":"/com.hello/1/server", 
    "env":"test","toggle":"1"
}'
```

###### 批量
```
curl -XPOST  -H"Content-Type":"application/json" http://localhost:8080/v1/projs -d
'[{
        "app_service":"com.dst", 
        "app_name":"dst-server", 
        "app_type":"v1",
        "module":"1", 
        "mod_type":"1",  
        "mod_version":"-1", 
        "host":"192.168.3.73",     
        "instance":"/com.hello/1/server", 
        "env":"test",
        "toggle":"1"
    }, 
    {  
        "app_service":"com.abc", 
        "app_name":"abc-server", 
        "app_type":"v1",
        "module":"1", 
        "mod_type":"1",  
        "mod_version":"-1", 
        "host":"192.168.3.73",     
        "instance":"/com.hello/1/server", 
        "env":"test",
        "toggle":"1"
    }
]'
```

### projserver
- 增加界面查询功能
- 提供接口插入功能