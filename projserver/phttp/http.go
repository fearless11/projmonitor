package phttp

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"projmonitor/projserver/conf"
	"projmonitor/projserver/model"
)

type AddProj struct{}
type AddProjs struct{}
type ReadProj struct{}

func (*AddProj) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	var item *model.Project
	err = json.Unmarshal(body, &item)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	item.Add()
	io.WriteString(w, "status:success")
}

func (*AddProjs) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := Receive(r)
	if err != nil {
		io.WriteString(w, err.Error())
	}
	io.WriteString(w, "status:success")
}

func Receive(r *http.Request) (err error) {
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	projects := make([]*model.Project, 0)

	_, err = dec.Token()
	if err != nil {
		log.Println("[ERROR] decoder token fail", err)
		return
	}
	for dec.More() {
		item := &model.Project{}
		if err = dec.Decode(&item); err != nil {
			log.Println(err)
			return
		}
		projects = append(projects, item)
	}
	model.BatchAdd(projects)
	return nil
}

type AllProj struct {
	Project []*conf.DetectedItem `json:"project"`
}

func (*ReadProj) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var items map[string][]*conf.DetectedItem
	items = conf.DetectedItemMap.GetALL()

	proj := &AllProj{}
	for _, item := range items {
		proj.Project = item
	}
	b, err := json.Marshal(proj)
	if err != nil {
		log.Println(err)
	}
	content := string(b)
	io.WriteString(w, content)
}

func Start() {
	mux := http.NewServeMux()
	mux.Handle("/v1/proj", &AddProj{})
	mux.Handle("/v1/apps", &ReadProj{})
	mux.Handle("/v1/projs", &AddProjs{})

	s := &http.Server{
		Addr:           conf.Config.Web,
		Handler:        mux,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
