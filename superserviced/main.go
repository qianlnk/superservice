package main

import (
	"flag"
	"net/http"
	"os/user"
	"time"

	"github.com/golang/glog"

	"golang.org/x/net/websocket"

	"qianno.xie/superservice/superserviced/api"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			glog.Errorf("recover err: %+v", err)
		}
	}()
	defer glog.Flush()
	ticker := time.NewTicker(1000 * 1000 * 100)
	go func() {
		for _ = range ticker.C {
			glog.Flush()
		}
	}()
	defer ticker.Stop()
	flag.Parse()

	selfuser, err := user.Current()
	if err != nil {
		glog.Errorf("get user info err: %s", err.Error())
	}
	if selfuser.Username != "root" {
		glog.Errorf("superservice request to run as root.")
		return
	}

	//	err = service.ServiceList.UpdateService("lnk", "./hello/hello lnk", "", "root", true, true)
	//	if err != nil {
	//		glog.Errorf("update service err: %s", err.Error())
	//	}
	//	time.Sleep(time.Second * 5)
	//	err = service.ServiceList.UpdateService("lnk", "./hello/hello newlnk", "", "root", true, true)
	//	if err != nil {
	//		glog.Errorf("update service err: %s", err.Error())
	//	}
	//	err = service.ServiceList.UpdateService("xzj", "./hello/hello xzj", "", "xiezhenjia", true, true)
	//	if err != nil {
	//		glog.Errorf("update service err: %s", err.Error())
	//	}
	//	for {
	//		time.Sleep(1 * time.Minute)
	//	}
	http.Handle("/Cmd", websocket.Handler(api.CmdHandle))
	http.HandleFunc("Release", api.ReleaseHandle)

	// initialize server
	srv := &http.Server{
		Addr:           ":5260",
		Handler:        nil,
		ReadTimeout:    time.Duration(30) * time.Second,
		WriteTimeout:   time.Duration(30) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// start listen
	err = srv.ListenAndServe()
	if err != nil {
		glog.Errorf("ERROR:main listen and serve failed!err:%+v", err)
		return
	}

	time.Sleep(10)
}
