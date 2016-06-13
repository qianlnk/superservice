package main

import (
	"flag"
	"net/http"
	//"os/user"
	"time"

	"github.com/golang/glog"

	"golang.org/x/net/websocket"

	"github.com/qianlnk/superservice/superserviced/command"
	"github.com/qianlnk/superservice/superserviced/release"
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

	//	selfuser, err := user.Current()
	//	if err != nil {
	//		glog.Errorf("get user info err: %s", err.Error())
	//	}
	//	if selfuser.Username != "root" {
	//		glog.Errorf("superservice request to run as root.")
	//		return
	//	}

	http.Handle("/Cmd", websocket.Handler(command.CmdHandle))
	http.HandleFunc("/Release", release.ReleaseHandle)

	// initialize server
	srv := &http.Server{
		Addr:           ":5260",
		Handler:        nil,
		ReadTimeout:    time.Duration(5) * time.Minute,
		WriteTimeout:   time.Duration(5) * time.Minute,
		MaxHeaderBytes: 1 << 20,
	}

	// start listen
	err := srv.ListenAndServe()
	if err != nil {
		glog.Errorf("ERROR:main listen and serve failed!err:%+v", err)
		return
	}

	time.Sleep(10)
}
