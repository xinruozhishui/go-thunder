package main

import (
	"github.com/xinruozhishui/go-thunder/business"
	"net/http"
)

func main() {
	//
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("view/dist/static"))))
	gdownsrv := new(business.DServ)
	// Judging the condition of exit
	business.ExitJudgment(gdownsrv)
	gdownsrv.LoadSetting(business.GetSettingPath())
	business.Open("http://localhost:9988/index.html")
	gdownsrv.Start(9988)
}