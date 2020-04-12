package conf

import (
	"log"
	"time"
	"fmt"
	"strconv"
	"github.com/influxdata/influxdb/client/v2"
)

func StringTofloat(str string) (f float64) {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		log.Println("[ERROR] string to float", err)
		return 0.0
	}
	return f
}

func writeToInlfux(proj *CheckResult) {
	
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:  Config.Influxdb.Addr,
	})
	if err != nil {
		log.Println("[ERROR] creating InfluxDB Client: ", err.Error())
	}
	defer c.Close()


	sql:=fmt.Sprintf(`CREATE DATABASE "%s"`,  Config.Influxdb.DataBase )
	q := client.NewQuery(sql, "", "")
	if response, err := c.Query(q); err != nil && response.Error() != nil {
		log.Println("[ERROR] inflxudb database fail",err.Error())
	}

	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database: Config.Influxdb.DataBase,
		Precision: "s",
	})

	name := "project"
	lbs := map[string]string{}
	fields := map[string]interface{}{}

	lbs["appservice"] = proj.AppService
	lbs["appname"] = proj.AppName
	lbs["module"] = proj.Module
	lbs["host"] = proj.Host
	lbs["env"] = proj.Env

	fields["cpu"] = StringTofloat(proj.PrecentOfCPU)
	fields["mem"] = StringTofloat(proj.Memory)
	fields["fd"] = StringTofloat(proj.FD)
	fields["thread"] = StringTofloat(proj.Thread)
	fields["tcp"] = StringTofloat(proj.TCP)
	fields["ygc"] = proj.Jvm.SizeYGC
	fields["ygct"] = proj.Jvm.SizeYGCT
	fields["fgc"] = proj.Jvm.SizeFGC
	fields["fgct"] = proj.Jvm.SizeFGCT
	fields["sizeygc"] = proj.Jvm.SizeYGCOC
	fields["avgsizeygc"] = proj.Jvm.AvgSizeYGCOC
	fields["ratiogct"] = proj.Jvm.RatioGC
	fields["sizeoldou"] = proj.Jvm.SizeOldOU
	fields["ratiooldou"] = proj.Jvm.RatioOldOU
	fields["sizeallheap"] = proj.Jvm.SizeAllHeap
	fields["sizeuseheap"] = proj.Jvm.SizeUseHeap

	if Config.Debug {
		log.Println("[DEBUG] writeToInlfux:", "labels:", lbs, "fields:", fields)
	}

	pt, err := client.NewPoint(name , lbs, fields, time.Now())
	if err != nil {
		log.Println("[ERROR] writeToInlfux", err.Error())
	}
	bp.AddPoint(pt)
	
	// Write the batch
	c.Write(bp)	
}
