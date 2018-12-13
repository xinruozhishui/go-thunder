package business

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"github.com/xinruozhishui/go-thunder/dao"
	"github.com/xinruozhishui/go-thunder/model"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
)

type DServ struct {
	dls []*Downloader
	// Mutex （互斥锁）和RWMutex（读写锁）
	oplock sync.Mutex
}

// Start is to handle the requests on incoming connections
func (srv *DServ) Start(listenPort int) error {
	http.HandleFunc("/", srv.Redirect)
	http.HandleFunc("/task/get", srv.progress)
	http.HandleFunc("/task/create", srv.createTask)
	http.HandleFunc("/task/delete", srv.deleteTask)
	http.HandleFunc("/task/start", srv.startTask)
	http.HandleFunc("/task/stop", srv.stopTask)
	http.HandleFunc("/task/start_all", srv.startAllTask)
	http.HandleFunc("/task/stop_all", srv.stopAllTask)
	http.HandleFunc("/index.html", srv.index)

	// listens on the TCP network address
	if err := http.ListenAndServe(":"+strconv.Itoa(listenPort), nil); err != nil {
		return err
	}
	return nil
}

func (service *DServ) Redirect(responseWriter http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	http.Redirect(responseWriter, request, "index.html", 301)
}

func (srv *DServ) index(rwr http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	rwr.Header().Set("Content-Type: text/html", "*")
	content, _ := ioutil.ReadFile("view/dist/index.html")
	rwr.Write(content)
}

// SaveSetting is to save all tasks'progresses in a file when the program exits
func (srv *DServ) SaveSetting(sf string) error {
	for _, i := range srv.dls {
		task := &model.Task{
			Id: i.Fi.Id,
			FileName: i.Fi.FileName,
			Size: i.Fi.Size,
			Url: i.Fi.Url,
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

// LoadSetting is to load all task'progress from a file when the program starts
func (srv *DServ) GetTaskList(sf string) error {
	task, err := dao.GetTaskList()
	if err != nil {
		return err
	}
	for _, r := range task {
		var dp []*DownloadProgress
		for _, v := range gjson.Parse(r.DownloadProgress).Array() {
			dp = append(dp, &DownloadProgress{
				From: gjson.Get(v.String(), "From").Int(),
				To: gjson.Get(v.String(), "To").Int(),
				Pos:gjson.Get(v.String(), "Pos").Int(),
				BytesInSecond: gjson.Get(v.String(), "BytesInSecond").Int(),
				Speed: gjson.Get(v.String(), "Speed").Int(),
				Lsmt: gjson.Get(v.String(), "Lsmt").Time(),
			})
		}
		dl, err := RestartDownloader(r.Id, r.Url, r.FileName, dp)
		if err != nil {
			return err
		}
		srv.dls = append(srv.dls, dl)
	}
	return nil
}
