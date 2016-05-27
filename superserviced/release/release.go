package release

import (
	"net/http"

	"github.com/golang/glog"
)

func ReleaseHandle(res http.ResponseWriter, req *http.Request) {
	glog.Infof("release handle start")
}
