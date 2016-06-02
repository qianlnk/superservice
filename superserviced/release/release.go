package release

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/golang/glog"
)

func ReleaseHandle(res http.ResponseWriter, req *http.Request) {
	glog.Infof("release handle start")
	if "POST" == req.Method {
		file, _, err := req.FormFile("releasefile")
		if err != nil {
			http.Error(res, err.Error(), 500)
			return
		}
		defer file.Close()
		f, err := os.Create("/tmp/helloLnk.tar")
		if err != nil {
			http.Error(res, err.Error(), 500)
			return
		}
		defer f.Close()
		n, err := io.Copy(f, file)
		if err != nil {
			http.Error(res, err.Error(), 500)
			return
		}
		fmt.Fprintf(res, "file size is %d", n)
	}
}
