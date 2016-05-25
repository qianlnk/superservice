package api

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
)

func CmdHandle(res http.ResponseWriter, req *http.Request) {
	glog.Infof("cmd handle start")
	fmt.Fprintf(res, "OK")
}

func ReleaseHandle(res http.ResponseWriter, req *http.Request) {
	glog.Infof("release handle start")
}
