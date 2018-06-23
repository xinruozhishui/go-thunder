package business

import (
	"os/user"
	"strconv"
	"os"
	"net/http"
	"log"
	"errors"
	"time"
	"github.com/xinruozhishui/go-thunder/library"
	"encoding/json"
	"io/ioutil"
)

// download file information
type FileInfo struct {
	// download file size
	Size     int64  `json:"Size"`
	// download file name
	FileName string `json:"FileName"`
	// download file url
	Url      string `json:"Url"`
}

// a downloader information
type Downloader struct {
	// a downloader can write at a file to mark download infomation
	sf *library.SafeFile
	// a downloader has a worker pool
	wp *WorkerPool
	// a downloader has a file information that will be downloaded
	Fi FileInfo
}

// downloader'all works progress
type DownloadProgress struct {
	// work'start position
	From          int64 `json:"From"`
	// work'end position
	To            int64 `json:"To"`
	// work' downloaded position
	Pos           int64 `json:"Pos"`
	// work' download bytes per second
	BytesInSecond int64
	// work' download speed
	Speed         int64
	Lsmt          time.Time
}

type PartialDownloader struct {
	dp     DownloadProgress
	client http.Client
	req    http.Response
	url    string
	file   *library.SafeFile
}

func getDown() string {
	usr, _ := user.Current()
	st := strconv.QuoteRune(os.PathSeparator)
	st = st[1 : len(st)-1]
	return usr.HomeDir + st + "Downloads" + st
}



func (pd *PartialDownloader) DoWork() (bool, error) {
	return pd.DownloadSergment()
}

func (pd PartialDownloader) GetProgress() interface{} {
	return pd.dp
}

func (pd *PartialDownloader) BeforeRun() error {
	return pd.BeforeDownload()
}

func (pd *PartialDownloader) AfterStop() error {
	return pd.AfterStopDownload()
}


// CreateDownloader is to creating a new Downloader
func CreateDownloader(url string, path string, count int64) (*Downloader, error)  {
	c, err := library.GetSize(url)
	if err != nil {
		//can't get file size
		return nil, err
	}

	dfs := getDown() + path
	sf, err := library.CreateSafeFile(dfs)
	if err != nil {
		//can't create file on path
		return nil, err
	}

	if err := sf.Truncate(c); err != nil {
		//can't truncate file
		return nil, err
	}
	//create part-downloader fsoreach segment
	ps := c / count
	wp := new(WorkerPool)
	for i := int64(0); i < count-int64(1); i++ {
		d := CreatePartialDownloader(url, sf, ps*i, ps*i, ps*i+ps)
		mv := MonitoredWorker{dw: d}
		wp.AppendWork(&mv)
	}
	d := Downloader{
		sf: sf,
		wp: wp,
		Fi: FileInfo{FileName: path, Size: c, Url: url},
	}
	return &d, nil
}



func CreatePartialDownloader(url string, file *library.SafeFile, from int64, pos int64, to int64) *PartialDownloader {
	var pd PartialDownloader
	pd.file = file
	pd.url = url
	pd.dp.From = from
	pd.dp.To = to
	pd.dp.Pos = pos
	return &pd
}

func (pd *PartialDownloader) DownloadSergment() (bool, error) {
	// write flush data to disk
	buffer := make([]byte, library.FlushDiskSize, library.FlushDiskSize)

	count, err := pd.req.Body.Read(buffer)
	log.Println(count)
	if (err != nil) && (err.Error() != "EOF") {
		pd.req.Body.Close()
		pd.file.Sync()
		return true, err
	}
	//log.Printf("returned from server %v bytes", count)
	if pd.dp.Pos+int64(count) > pd.dp.To {
		count = int(pd.dp.To - pd.dp.Pos)
		log.Printf("warning: server return to much for me i give only %v bytes", count)
	}

	// write the bytes to the file, and return the number of bytes written and an error
	realc, err := pd.file.WriteAt(buffer[:count], pd.dp.Pos)
	if err != nil {
		pd.file.Sync()
		pd.req.Body.Close()
		return true, err
	}
	pd.dp.Pos = pd.dp.Pos + int64(realc)
	pd.CalculationSpeed(realc)
	//log.Printf("writed %v pos %v to %v", realc, pd.dp.Pos, pd.dp.To)
	if pd.dp.Pos == pd.dp.To {
		//ok download part complete normal
		pd.file.Sync()
		pd.req.Body.Close()
		pd.dp.Speed = 0
		log.Printf("info: download complete normal")
		return true, nil
	}
	//not full download next segment
	return false, nil
}

func (pd *PartialDownloader) BeforeDownload() error {
	//create new req
	r, err := http.NewRequest("GET", pd.url, nil)
	if err != nil {
		return err
	}

	r.Header.Add("Range", "bytes="+strconv.FormatInt(pd.dp.Pos, 10)+"-"+strconv.FormatInt(pd.dp.To, 10))
	f,_ := library.CreateSafeFile("test")
	r.Write(f)
	f.Close()
	resp, err := pd.client.Do(r)
	if err != nil {
		log.Printf("error: error download part file%v \n", err)
		return err
	}
	//check response
	if resp.StatusCode != 206 {
		log.Printf("error: file not found or moved status:", resp.StatusCode)
		return errors.New("error: file not found or moved")
	}
	pd.req = *resp
	return nil
}

func (pd *PartialDownloader) AfterStopDownload() error {
	log.Println("info: try sync file")
	log.Println(pd.req.Body.Close())
	return pd.file.Sync()
}


func RestoreDownloader(url string, fp string, dp []DownloadProgress) (dl *Downloader, err error) {
	dfs := getDown() + fp
	sf, err := library.OpenSafeFile(dfs)
	if err != nil {
		//can't create file on path
		return nil, err
	}
	s, err := sf.Stat()
	if err != nil {
		return nil, err
	}
	wp := new(WorkerPool)
	for _, r := range dp {
		dow := CreatePartialDownloader(url, sf, r.From, r.Pos, r.To)
		mv := MonitoredWorker{dw: dow}

		//add to worker pool
		wp.AppendWork(&mv)

	}
	d := Downloader{
		sf: sf,
		wp: wp,
		Fi: FileInfo{FileName: fp, Size: s.Size(), Url: url},
	}
	return &d, nil
}

func (dl *Downloader) GetProgress() []DownloadProgress {
	pr := dl.wp.GetAllProgress().([]interface{})
	re := make([]DownloadProgress, len(pr))
	for i, val := range pr {
		re[i] = val.(DownloadProgress)
	}
	return re
}


// CalculationSpeed is to alculation download speed
func (pd *PartialDownloader) CalculationSpeed(realc int) {
	if time.Since(pd.dp.Lsmt).Seconds() > 0.5 {
		pd.dp.Speed = 2 * pd.dp.BytesInSecond
		pd.dp.Lsmt = time.Now()
		pd.dp.BytesInSecond = 0
	} else {
		pd.dp.BytesInSecond += int64(realc)
	}

}

// StopAllDownloader is to stop all downloaders
func (dl *Downloader) StopAllDownloader() []error {
	return dl.wp.StopAll()
}

// StartAllDownloader is to start all downloaders
func (dl *Downloader) StartAllDownloader() []error {
	return dl.wp.StartAll()
}


type DownloadSettings struct {
	FI FileInfo           `json:"FileInfo"`
	Dp []DownloadProgress `json:"DownloadProgress"`
}

type ServiceSettings struct {
	Ds []DownloadSettings
}

func LoadFromFile(s string) (*ServiceSettings, error) {
	sb, err := ioutil.ReadFile(s)
	if err != nil {
		library.CreateSafeFile(s)
	}

	var ss ServiceSettings
	err = json.Unmarshal(sb, &ss)
	if err != nil {
		return nil, err
	}
	return &ss, nil
}

func (s *ServiceSettings) SaveToFile(fp string) error {
	dat, err := json.Marshal(s)
	if err != nil {
		return err
	}
	log.Println("info: try save settings")
	log.Println(string(dat))
	err = ioutil.WriteFile(fp, dat, 0664)
	if err != nil {
		return err
	}
	return nil
}

// GetSettingPath is to get setting'path
func GetSettingPath() string  {
	u, _ := user.Current()
	st := strconv.QuoteRune(os.PathSeparator)
	st = st[1 : len(st)-1]
	return u.HomeDir + st + library.SETTING_FILE_NAME
}
