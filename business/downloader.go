package business

import (
	"encoding/json"
	"errors"
	"github.com/xinruozhishui/go-thunder/dao"
	"github.com/xinruozhishui/go-thunder/library"
	"github.com/xinruozhishui/go-thunder/model"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"time"
)

type FileInfo struct {
	Id int64 `json:"id"`
	Size int64 `json:"Size"`
	FileName string `json:"FileName"`
	Url string `json:"Url"`
}

type Downloader struct {
	sf *library.SafeFile
	wp *WorkerPool
	Fi FileInfo
}

type DownloadProgress struct {
	From int64 `json:"From"`
	To int64 `json:"To"`
	Pos int64 `json:"Pos"`
	BytesInSecond int64
	Speed int64
	Lsmt  time.Time
}

type PartialDownloader struct {
	dp     DownloadProgress
	client http.Client
	req    http.Response
	url    string
	file   *library.SafeFile
}

type DownloadSettings struct {
	FI FileInfo           `json:"FileInfo"`
	Dp []DownloadProgress `json:"DownloadProgress"`
}

type ServiceSettings struct {
	Ds []DownloadSettings
}

func getDown() string {
	usr, _ := user.Current()
	st := strconv.QuoteRune(os.PathSeparator)
	st = st[1 : len(st)-1]
	return usr.HomeDir + st + "Downloads" + st
}

func CreateDownloader(url string, path string, count int64) (*Downloader, error) {
	c, err := library.GetSize(url)
	if err != nil {
		return nil, err
	}

	dfs := getDown() + path
	sf, err := library.CreateSafeFile(dfs)
	if err != nil {
		return nil, err
	}

	if err := sf.Truncate(c); err != nil {
		return nil, err
	}
	ps := c / count
	wp := new(WorkerPool)
	for i := int64(0); i < count-int64(1); i++ {
		d := CreatePartialDownloader(url, sf, ps*i, ps*i, ps*i+ps)
		mv := MonitoredWorker{dw: d}
		wp.AppendWork(&mv)
	}
	lastSeg := int64(ps * (count - 1))
	dow := CreatePartialDownloader(url, sf, lastSeg, lastSeg, c)
	mv := MonitoredWorker{dw: dow}

	wp.AppendWork(&mv)
	task, err := dao.CreateTask(&model.Task{
		FileName: path,
		Size:     c,
		Url:      url,
	})
	if err != nil {
		return nil, err
	}
	d := Downloader{
		sf: sf,
		wp: wp,
		Fi: FileInfo{Id: task.Id, FileName: task.FileName, Size: task.Size, Url: task.Url},
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

func (o *PartialDownloader) DownloadSegment() (bool, error) {
	buffer := make([]byte, library.FlushDiskSize, library.FlushDiskSize)

	count, err := o.req.Body.Read(buffer)
	log.Println(count)
	if (err != nil) && (err.Error() != "EOF") {
		o.req.Body.Close()
		o.file.Sync()
		return true, err
	}
	if o.dp.Pos+int64(count) > o.dp.To {
		count = int(o.dp.To - o.dp.Pos)
		log.Printf("warning: server return to much for me i give only %v bytes", count)
	}

	realc, err := o.file.WriteAt(buffer[:count], o.dp.Pos)
	if err != nil {
		o.file.Sync()
		o.req.Body.Close()
		return true, err
	}
	o.dp.Pos = o.dp.Pos + int64(realc)
	o.CalculationSpeed(realc)
	if o.dp.Pos == o.dp.To {
		o.file.Sync()
		o.req.Body.Close()
		o.dp.Speed = 0
		log.Printf("info: download complete normal")
		return true, nil
	}
	return false, nil
}

func (o *PartialDownloader) BeforeDownload() error {
	r, err := http.NewRequest("GET", o.url, nil)
	if err != nil {
		return err
	}

	r.Header.Add("Range", "bytes="+strconv.FormatInt(o.dp.Pos, 10)+"-"+strconv.FormatInt(o.dp.To, 10))
	resp, err := o.client.Do(r)
	if err != nil {
		log.Printf("error: error download part file%v \n", err)
		return err
	}
	if resp.StatusCode != 206 {
		log.Printf("error: file not found or moved status:%d", resp.StatusCode)
		return errors.New("error: file not found or moved")
	}
	o.req = *resp
	return nil
}

func (o *PartialDownloader) AfterStopDownload() error {
	log.Println("info: try sync file")
	log.Println(o.req.Body.Close())
	return o.file.Sync()
}

func RestartDownloader(id int64, url string, fileName string, dp []*DownloadProgress) (dl *Downloader, err error) {
	dfs := getDown() + fileName
	sf, err := library.OpenSafeFile(dfs)
	if err != nil {
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
		Fi: FileInfo{Id: id, FileName: fileName, Size: s.Size(), Url: url},
	}
	return &d, nil
}

func (o *Downloader) GetProgress() []DownloadProgress {
	pr := o.wp.GetAllProgress().([]interface{})
	re := make([]DownloadProgress, len(pr))
	for i, val := range pr {
		re[i] = val.(DownloadProgress)
	}
	return re
}

func (o *PartialDownloader) CalculationSpeed(realc int) {
	if time.Since(o.dp.Lsmt).Seconds() > 0.5 {
		o.dp.Speed = 2 * o.dp.BytesInSecond
		o.dp.Lsmt = time.Now()
		o.dp.BytesInSecond = 0
	} else {
		o.dp.BytesInSecond += int64(realc)
	}

}

func (o *Downloader) StopAllDownloader() []error {
	return o.wp.StopAll()
}

func (o *Downloader) StartAllDownloader() []error {
	return o.wp.StartAll()
}

func (o *ServiceSettings) SaveToFile(fp string) error {
	dat, err := json.Marshal(o)
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

func GetSettingPath() string {
	u, _ := user.Current()
	st := strconv.QuoteRune(os.PathSeparator)
	st = st[1 : len(st)-1]
	return u.HomeDir + st + library.SettingFileName
}

func (o *PartialDownloader) DoWork() (bool, error) {
	return o.DownloadSegment()
}

func (o PartialDownloader) GetProgress() interface{} {
	return o.dp
}

func (o *PartialDownloader) BeforeRun() error {
	return o.BeforeDownload()
}

func (o *PartialDownloader) AfterStop() error {
	return o.AfterStopDownload()
}
