package business

import (
	"errors"
	"github.com/xinruozhishui/go-thunder/library"
	"log"
	"sync"
)

const (
	Stopped = iota
	Running
	Failed
	Completed
)

//
type DiscretWork interface {
	DoWork() (bool, error)
	GetProgress() interface{}
	BeforeRun() error
	AfterStop() error
}

type MonitoredWorker struct {
	lc    sync.Mutex
	dw    DiscretWork
	wgRun sync.WaitGroup
	guid  string
	state int
	chsig chan int
	stwg  sync.WaitGroup
}

func (mw *MonitoredWorker) wgoroute() {
	log.Println("info: work start", mw.GetId())
	defer func() {
		log.Print("info: realease work guid ", mw.GetId())
		mw.wgRun.Done()
	}()

	for {
		select {
		case newState := <-mw.chsig:
			if newState == Stopped {
				mw.state = newState
				log.Println("info: work stopped")
				return
			}
		default:
			{
				isdone, err := mw.dw.DoWork()
				if err != nil {
					log.Println("error: guid", mw.guid, " work failed", err)
					mw.state = Failed
					return
				}
				if isdone {
					mw.state = Completed
					log.Println("info: work done")
					return
				}
			}

		}
	}
}
func (mw MonitoredWorker) GetState() int {
	return mw.state
}
func (mw *MonitoredWorker) GetId() string {
	if len(mw.guid) == 0 {
		mw.guid = library.GenUid()
	}
	return mw.guid

}

func (mw *MonitoredWorker) Start() error {
	mw.lc.Lock()
	defer mw.lc.Unlock()
	if mw.state == Completed {
		return errors.New("error: try run completed job")
	}
	if mw.state == Running {
		return errors.New("error: try run runing job")
	}
	if err := mw.dw.BeforeRun(); err != nil {
		mw.state = Failed
		return err
	}
	mw.chsig = make(chan int, 1)
	mw.state = Running
	mw.wgRun.Add(1)
	go mw.wgoroute()

	return nil
}

func (mw *MonitoredWorker) Stop() error {
	mw.lc.Lock()
	defer mw.lc.Unlock()
	if mw.state != Running {
		return errors.New("error: imposible stop non runing job")

	}
	mw.chsig <- Stopped
	mw.wgRun.Wait()
	close(mw.chsig)
	if err := mw.dw.AfterStop(); err != nil {
		return err
	}
	return nil
}

func (mw *MonitoredWorker) Wait() {
	mw.wgRun.Wait()
}

func (mw MonitoredWorker) GetProgress() interface{} {
	return mw.dw.GetProgress()

}
