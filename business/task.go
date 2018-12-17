package business

import (
	"os"
	"os/signal"
	"syscall"
)

type TaskInfo struct {
	Id         int
	FileName   string
	Size       int64
	Downloaded int64
	Progress   int64
	Speed      int64
}

type NewTask struct {
	Url       string
	PartCount int64
	FilePath  string
}

func ExitJudgment(gdownsrv *DServ)  {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		func() {
			gdownsrv.StopAllTask()
			gdownsrv.SaveAllTask(GetSettingPath())
		}()
		os.Exit(1)
	}()
}

