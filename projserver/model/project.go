package model

import "log"

type Project struct {
	Id         int64  `json:"id"`
	AppService string `json:"app_service"`
	AppName    string `json:"app_name"`
	AppType    string `json:"app_type"`
	Module     string `json:"module"`
	ModType    string `json:"mod_type"`
	ModVersion string `json:"mod_version"`
	Host       string `json:"host"`
	Instance   string `json:"instance"`
	Env        string `json:"env"`
	Toggle     string `json:"toggle"`
}

func GetAllProjectByCron() ([]*Project, error) {
	projects := make([]*Project, 0)
	err := Orm.Find(&projects)
	return projects, err
}

func GetbyToggle() ([]*Project, error) {
	projects := make([]*Project, 0)
	err := Orm.Find(&projects)
	return projects, err
}

func BatchAdd(items []*Project) {
	_, err := Orm.Insert(items)
	if err != nil {
		log.Println("[ERROR] batch insert mysql fail", err)
	}
}

func (this *Project) Add() (int64, error) {
	id, err := Orm.InsertOne(this)
	if err != nil {
		log.Println("[ERROR]", this.AppName, this.Module, this.Host, "insert into mysql fail", err)
	}
	return id, err
}
