package business

import (
	"net/http"
	"strconv"
	"io/ioutil"
	"sync"
	"log"
)

type DServ struct {
	dls    []*Downloader
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
	var ss ServiceSettings
	for _, i := range srv.dls {
		ss.Ds = append(ss.Ds, DownloadSettings{
			FI: i.Fi,
			Dp: i.GetProgress(),
		})
	}

	return ss.SaveToFile(sf)
}

// LoadSetting is to load all task'progress from a file when the program starts
func (srv *DServ) LoadSetting(sf string) error {
	ss, err := LoadFromFile(sf)
	if err != nil {
		//log.Println("error: when try load settings", err)
		return err
	}
	log.Println(ss)
	for _, r := range ss.Ds {
		dl, err := RestoreDownloader(r.FI.Url, r.FI.FileName, r.Dp)
		if err != nil {
			return err
		}
		srv.dls = append(srv.dls, dl)
	}
	return nil
}


