package release

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/golang/glog"
	"github.com/qianlnk/compress"
	"github.com/qianlnk/superservice/superserviced/service"
)

func ReleaseHandle(res http.ResponseWriter, req *http.Request) {
	glog.Infof("release handle start")
	if "POST" == req.Method {
		user := req.URL.Query().Get("user")
		pwd := req.URL.Query().Get("password")
		tmpfile := req.URL.Query().Get("file")
		filename := req.URL.Query().Get("filename")
		fmt.Println(user, pwd, tmpfile, filename)
		fmt.Println(service.ServiceList)
		v, ok := service.ServiceList[filename]
		if !ok {
			fmt.Fprintf(res, "service %s not exist, please Add first", filename)
			return
		}
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
		err = compress.Uncompress(tmpfile, v.Directory)
		if err != nil {
			fmt.Fprintf(res, "err: %s", err.Error())
			return
		}
		fmt.Fprintf(res, "file size is %d", n)
	}
}
