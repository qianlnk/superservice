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
		user := req.URL.Query().Get("user")
		pwd := req.URL.Query().Get("password")
		tmpfile := req.URL.Query().Get("file")
		filename := req.URL.Query().Get("filename")
		fmt.Println(user, pwd, tmpfile, filename)
		file, _, err := req.FormFile(filename)
		if err != nil {
			http.Error(res, err.Error(), 500)
			return
		}
		defer file.Close()
		f, err := os.Create(tmpfile)
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
