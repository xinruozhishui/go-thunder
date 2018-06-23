package business

import (
	"log"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
)

// a task download process info
type TaskInfo struct {
	// task'ID
	Id         int
	// task download file name
	FileName   string
	// task download file size
	Size       int64
	// task has downloaded size
	Downloaded int64
	// task has downloaded size percentage
	Progress   int64
	// task current download speed
	Speed      int64
}

// a new task info
type NewTask struct {
	// download file url
	Url       string
	// The number of partitioned copies of the downloaded file
	PartCount int64
	// file local save path
	FilePath  string
}

// createTask is to create a download task
func (srv *DServ) createTask(rwr http.ResponseWriter, req *http.Request) {
	srv.oplock.Lock()
	defer func() {
		srv.oplock.Unlock()
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
	srv.dls = append(srv.dls, dl)
	js, _ := json.Marshal("ok")
	rwr.Write(js)
}

// startTask is to start a download task
func (srv *DServ) startTask(rwr http.ResponseWriter, req *http.Request) {
	srv.oplock.Lock()
	defer func() {
		srv.oplock.Unlock()
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
	if !(len(srv.dls) > ind) {
		http.Error(rwr, "error: id is out of jobs list", http.StatusInternalServerError)
		return
	}

	if errs := srv.dls[ind].StartAllDownloader(); len(errs) > 0 {
		http.Error(rwr, "error: can't start all part", http.StatusInternalServerError)
		return
	}
	js, _ := json.Marshal("ok")
	rwr.Write(js)
}

// stopTask is to stop a download task
func (srv *DServ) stopTask(rwr http.ResponseWriter, req *http.Request) {
	srv.oplock.Lock()
	defer func() {
		srv.oplock.Unlock()
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
	if !(len(srv.dls) > ind) {
		http.Error(rwr, "error: id is out of jobs list", http.StatusInternalServerError)
		return
	}

	srv.dls[ind].StopAllDownloader()
	js, _ := json.Marshal("ok")
	rwr.Write(js)
}

// startAllTask is to start All download tasks
func (srv *DServ) startAllTask(rwr http.ResponseWriter, req *http.Request) {
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
	srv.StartAllTask()
	js, _ := json.Marshal("ok")
	rwr.Write(js)
}

// StopAllTask is to stop all tasks
func (srv *DServ) stopAllTask(rwr http.ResponseWriter, req *http.Request) {
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
	srv.StopAllTask()
	js, _ := json.Marshal("ok")
	rwr.Write(js)
}

// StartAllTask is to start all Tasks
func (srv *DServ) StartAllTask() {
	srv.oplock.Lock()
	defer func() {
		srv.oplock.Unlock()
	}()
	for _, e := range srv.dls {
		log.Println("info start all result:", e.StartAllDownloader())
	}
}

// StopAllTask is to stop all tasks
func (srv *DServ) StopAllTask() {
	srv.oplock.Lock()
	// 在该函数程序结束的时候执行收尾
	defer func() {
		srv.oplock.Unlock()
	}()
	for _, e := range srv.dls {
		log.Println("info stopall result:", e.StopAllDownloader())
	}
}

// progress is to get all tasks real time progress
func (srv *DServ) progress(rwr http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	rwr.Header().Set("Access-Control-Allow-Origin", "*")
	jbs := make([]TaskInfo, 0, len(srv.dls))
	for ind, i := range srv.dls {
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
			Progress:   (d * 100 / i.Fi.Size),
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

// deleteTask is to delete a download task
func (srv *DServ) deleteTask(rwr http.ResponseWriter, req *http.Request) {
	srv.oplock.Lock()
	defer func() {
		srv.oplock.Unlock()
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
	if !(len(srv.dls) > ind) {
		http.Error(rwr, "error: id is out of jobs list", http.StatusInternalServerError)
		return
	}

	log.Printf("try stop segment download %v", srv.dls[ind].StopAllDownloader())
	srv.dls = append(srv.dls[:ind], srv.dls[ind+1:]...)
	js, _ := json.Marshal("ok")
	rwr.Write(js)
}


// Exitjudgment is to judge the condition of exit
func ExitJudgment(gdownsrv *DServ)  {
	c := make(chan os.Signal, 1)
	// monitor the end signal
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		// If there is an end signal，stop all download tasks,and save all download tasks'progress to a file
		<-c
		func() {
			gdownsrv.StopAllTask()
			gdownsrv.SaveSetting(GetSettingPath())
		}()
		os.Exit(1)
	}()
}

