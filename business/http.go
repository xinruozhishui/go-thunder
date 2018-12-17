package business

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"github.com/xinruozhishui/go-thunder/dao"
	"github.com/xinruozhishui/go-thunder/model"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type DServ struct {
	dls []*Downloader
	opLock sync.Mutex
}

func (o *DServ) Start(listenPort int) error {
	http.HandleFunc("/", o.redirect)
	http.HandleFunc("/index.html", o.index)
	http.HandleFunc("/task/get", o.progress)
	http.HandleFunc("/task/create", o.createTask)
	http.HandleFunc("/task/delete", o.deleteTask)
	http.HandleFunc("/task/start", o.startTask)
	http.HandleFunc("/task/stop", o.stopTask)
	http.HandleFunc("/task/start_all", o.startAllTask)
	http.HandleFunc("/task/stop_all", o.stopAllTask)

	if err := http.ListenAndServe(":"+strconv.Itoa(listenPort), nil); err != nil {
		return err
	}
	return nil
}

func (o *DServ) redirect(responseWriter http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	http.Redirect(responseWriter, request, "index.html", 301)
}

func (o *DServ) index(rwr http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	rwr.Header().Set("Content-Type: text/html", "*")
	content, _ := ioutil.ReadFile("view/dist/index.html")
	rwr.Write(content)
}

func (o *DServ) createTask(rwr http.ResponseWriter, req *http.Request) {
	o.opLock.Lock()
	defer func() {
		o.opLock.Unlock()
		req.Body.Close()
	}()
	bodyData, err := ioutil.ReadAll(req.Body)
	rwr.Header().Set("Access-Control-Allow-Origin", "*")
	if err != nil {
		http.Error(rwr, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	var nt NewTask
	if err := json.Unmarshal(bodyData, &nt); err != nil {
		http.Error(rwr, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	dl, err := CreateDownloader(nt.Url, nt.FilePath, nt.PartCount)
	if err != nil {
		http.Error(rwr, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	o.dls = append(o.dls, dl)
	dl.StartAllDownloader()
	js, _ := json.Marshal("ok")
	rwr.Write(js)
}

func (o *DServ) startTask(rwr http.ResponseWriter, req *http.Request) {
	o.opLock.Lock()
	defer func() {
		o.opLock.Unlock()
		req.Body.Close()
	}()
	bodyData, err := ioutil.ReadAll(req.Body)
	rwr.Header().Set("Access-Control-Allow-Origin", "*")
	if err != nil {
		http.Error(rwr, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	var ind int
	if err := json.Unmarshal(bodyData, &ind); err != nil {
		http.Error(rwr, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if !(len(o.dls) > ind) {
		http.Error(rwr, "error: id is out of jobs list", http.StatusInternalServerError)
		return
	}

	if errs := o.dls[ind].StartAllDownloader(); len(errs) > 0 {
		http.Error(rwr, "error: can't start all part", http.StatusInternalServerError)
		return
	}
	js, _ := json.Marshal("ok")
	rwr.Write(js)
}

func (o *DServ) stopTask(rwr http.ResponseWriter, req *http.Request) {
	o.opLock.Lock()
	defer func() {
		o.opLock.Unlock()
		req.Body.Close()
	}()
	bodyData, err := ioutil.ReadAll(req.Body)
	rwr.Header().Set("Access-Control-Allow-Origin", "*")
	if err != nil {
		http.Error(rwr, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	var ind int
	if err := json.Unmarshal(bodyData, &ind); err != nil {
		http.Error(rwr, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if !(len(o.dls) > ind) {
		http.Error(rwr, "error: id is out of jobs list", http.StatusInternalServerError)
		return
	}
	task := &model.Task{
		Id: o.dls[ind].Fi.Id,
	}
	dpStr, err := json.Marshal(o.dls[ind].GetProgress())
	if err != nil {
		log.Println("jsonMarshalErr:", err)
	}
	task.DownloadProgress = string(dpStr)
	if err := dao.UpdateTask(task); err != nil {
		log.Println("UpdateTask:", err)
	}

	o.dls[ind].StopAllDownloader()
	js, _ := json.Marshal("ok")
	rwr.Write(js)
}

func (o *DServ) startAllTask(rwr http.ResponseWriter, req *http.Request) {
	defer func() {
		req.Body.Close()
	}()
	_, err := ioutil.ReadAll(req.Body)
	rwr.Header().Set("Access-Control-Allow-Origin", "*")
	if err != nil {
		http.Error(rwr, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	o.StartAllTask()
	js, _ := json.Marshal("ok")
	rwr.Write(js)
}

func (o *DServ) stopAllTask(rwr http.ResponseWriter, req *http.Request) {
	defer func() {
		req.Body.Close()
	}()
	_, err := ioutil.ReadAll(req.Body)
	rwr.Header().Set("Access-Control-Allow-Origin", "*")
	if err != nil {
		http.Error(rwr, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	o.StopAllTask()
	js, _ := json.Marshal("ok")
	rwr.Write(js)
}

func (o *DServ) StartAllTask() {
	o.opLock.Lock()
	defer func() {
		o.opLock.Unlock()
	}()
	for _, e := range o.dls {
		log.Println("info start all result:", e.StartAllDownloader())
	}
}

func (o *DServ) StopAllTask() {
	o.opLock.Lock()
	defer func() {
		o.opLock.Unlock()
	}()
	for _, e := range o.dls {
		log.Println("info stopall result:", e.StopAllDownloader())
	}
}

func (o *DServ) progress(rwr http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	rwr.Header().Set("Access-Control-Allow-Origin", "*")
	jbs := make([]TaskInfo, 0, len(o.dls))
	for ind, i := range o.dls {
		prs := i.GetProgress()
		var d int64
		var s int64
		for _, p := range prs {
			d = d + (p.Pos - p.From)
			s += p.Speed
		}
		j := TaskInfo{
			Id:         ind,
			FileName:   i.Fi.FileName,
			Size:       i.Fi.Size,
			Progress:   d * 100 / i.Fi.Size,
			Downloaded: d,
			Speed:      s,
		}
		jbs = append(jbs, j)
	}
	js, err := json.Marshal(jbs)
	if err != nil {
		http.Error(rwr, err.Error(), http.StatusInternalServerError)
		return
	}
	rwr.Write(js)
}

func (o *DServ) deleteTask(rwr http.ResponseWriter, req *http.Request) {
	o.opLock.Lock()
	defer func() {
		o.opLock.Unlock()
		req.Body.Close()
	}()
	bodyData, err := ioutil.ReadAll(req.Body)
	rwr.Header().Set("Access-Control-Allow-Origin", "*")
	if err != nil {
		http.Error(rwr, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	var ind int
	if err := json.Unmarshal(bodyData, &ind); err != nil {
		http.Error(rwr, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if !(len(o.dls) > ind) {
		http.Error(rwr, "error: id is out of jobs list", http.StatusInternalServerError)
		return
	}

	log.Printf("try stop segment download %v", o.dls[ind].StopAllDownloader())
	if err := dao.DeleteTask(o.dls[ind].Fi.Id); err != nil {
		log.Println("DeleteTaskErr:", err)
	}
	o.dls = append(o.dls[:ind], o.dls[ind+1:]...)
	js, _ := json.Marshal("ok")
	rwr.Write(js)
}

func (o *DServ) SaveAllTask(sf string) error {
	for _, i := range o.dls {
		task := &model.Task{
			Id:       i.Fi.Id,
			FileName: i.Fi.FileName,
			Size:     i.Fi.Size,
			Url:      i.Fi.Url,
		}
		dpStr, err := json.Marshal(i.GetProgress())
		if err != nil {
			return err
		}
		task.DownloadProgress = string(dpStr)
		if err := dao.UpdateTask(task); err != nil {
			return err
		}
	}
	return nil
}

func (o *DServ) GetTaskList(sf string) error {
	task, err := dao.GetTaskList()
	if err != nil {
		return err
	}
	for _, r := range task {
		var dp []*DownloadProgress
		for _, v := range gjson.Parse(r.DownloadProgress).Array() {
			dp = append(dp, &DownloadProgress{
				From:          gjson.Get(v.String(), "From").Int(),
				To:            gjson.Get(v.String(), "To").Int(),
				Pos:           gjson.Get(v.String(), "Pos").Int(),
				BytesInSecond: gjson.Get(v.String(), "BytesInSecond").Int(),
				Speed:         gjson.Get(v.String(), "Speed").Int(),
				Lsmt:          gjson.Get(v.String(), "Lsmt").Time(),
			})
		}
		dl, err := RestartDownloader(r.Id, r.Url, r.FileName, dp)
		if err != nil {
			return err
		}
		o.dls = append(o.dls, dl)
	}
	return nil
}
