### 2018/12/03 
- 兼容新框架springcloud
  - 数据库 module字段设置为sc ，app_type字段设置为sc即可


### 【agent】
#### 1. 检测服务pid是否存在，不存在拉起
```
 (1）每隔5分钟拉取Server端项目信息
 (2）检测项目pid，每个项目启动goroutine，goroutine设置超时机制，超时时间1s
 (3）先读取一次pid文件，文件不存在，步骤（4）
 (4）触发pgrep查询。查询不到拉起服务，拉起服务前pgrep再次确认pid，不存在拉起同时异步发送pid不存在告警
```

##### 2. 检测项目cpu、内存、out文件
```
 (1）每隔2分钟检测缓存中项目的cpu、内存、out文件信息
 (2）如果缓存中项目不存在，通过pgrep检测项目pid，检测失败发送异步告警
 (3）检测cpu，当cpu百分比超过100%时，校验项目pid，校验通过，重启服务并异步发送警cpu过高告警
 (4）检测内存，当内存剩余小于200M时，校验项目pid，校验通过，重启服务并异步发送内存不足告警
 (5）检测out文件：
 	当文件大于10M，不检测内容，异步告警。
 	首次检测out，直接将指针指向末尾，不检测内容
 	文件无变化，不检测内容
 	文件变化过快，在2分钟内变化3M，不检测内容，异步告警
 	文件正常变化，检测变化内容中是否含有ERROR，有则告警
```

#### 代码调试
```
机器: 10.10.10.147
目录: /root/go/src/projmonitor
打包分发脚本139: /home/vera/pack_projagent.sh

数据库
 select * from project where app_name="t8t-wkf-bpm";
 insert project(app_name,app_type,module,host,env,toggle) value("t8t-wkf-bpm","sc","sc","10.10.10.147","uat","1");

139上执行:
    IDC
	1. 分发代码
	   ansible JAVAIDC -m copy -a 'src=/home/vera.jiang/projagent-idc.tgz dest=/tmp' -k
	   ansible JAVAIDC -m shell -a 'tar xvf /tmp/projagent-idc.tgz -C /data/to8to/tools/projagent' -k

	   ansible JAVAIDC -m shell -a '/bin/cp /data/to8to/tools/projmonitor/projagent/projagent /data/back' -k
	   ansible JAVAIDC -m copy -a 'src=/home/vera.jiang/projagent dest=/data/to8to/tools/projmonitor/projagent/projagent' -k
	2. 确认projagent是在运行
	   ansible JAVAIDC -m shell -a 'ps aux |grep projagent |grep -v grep 
	3. kill掉projagent后，监控自动拉起
	   ansible JAVAIDC -m shell -a 'ps aux |grep projagent |grep -v grep| awk "{print \$2}"| xargs kill -9' -k

    UAT
	1. 分发代码
	   ansible JAVAUAT -m copy -a 'src=/home/vera.jiang/projagent-idc.tgz dest=/tmp' -k
	   ansible JAVAUAT -m shell -a 'tar xvf /tmp/projagent-idc.tgz -C /data/to8to/tools/projagent' -k
	2. 确认projagent是在运行
	   ansible JAVAUAT -m shell -a 'ps aux |grep projagent |grep -v grep 
	3. kill掉projagent后，监控自动拉起
	   ansible JAVAUAT -m shell -a 'ps aux |grep projagent |grep -v grep| awk "{print \$2}"| xargs kill -9' -k

或者你可以clone代码
	git clone ssh://git@repo.we.com:22/vera.jiang/projmonitor.git
	git checkout -b test origin/test

```
