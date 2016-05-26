package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/golang/glog"

	"golang.org/x/net/websocket"

	"qianno.xie/superservice/longsocket"
	"qianno.xie/superservice/superserviced/service"
)

type User struct {
	Name     string
	Password string
}

type Message struct {
}

//list start stop restart delete update add log
type Cmd struct {
	Type string
	service.Service
}

func dealCommand(cmd Cmd) {
	fmt.Println(cmd)
	message := make(chan string)
	go sendMessage(message)
	switch strings.ToUpper(cmd.Type) {
	case "ADD":
		service.ServiceList.UpdateService(cmd.Name, cmd.Command, cmd.Directory, cmd.User, cmd.AutoStart, cmd.AutoRestart, message)
		break
	case "DELETE":
		service.ServiceList.Delete(cmd.Name, message)
		break
	case "UPDATE":
		service.ServiceList.UpdateService(cmd.Name, cmd.Command, cmd.Directory, cmd.User, cmd.AutoStart, cmd.AutoRestart, message)
		break
	case "LIST":
		service.ServiceList.List(message)
		break
	case "START":
		service.ServiceList[cmd.Name].Start(message)
		break
	case "STOP":
		service.ServiceList[cmd.Name].Stop(message)
		break
	case "RESTART":
		service.ServiceList[cmd.Name].Restart(message)
		break
	case "LOG":
		break
	default:
		break
	}
}

func dealMsg(msg []byte) error {
	fmt.Println(msg)
	return nil
}

func sendMessage(msg chan string) {
	fmt.Println("sendMessage")
	for {
		select {
		case m, ok := <-msg:
			if !ok {
				return
			}
			fmt.Println(m)
		}
	}
}

func Verify(user, password string) bool {
	return true
}

func CmdHandle(ws *websocket.Conn) {
	req := ws.Request()
	u, err := url.Parse(req.Header.Get("Origin"))
	if err != nil {
		ws.Close()
		return
	}

	user := u.Query().Get("user")
	password := u.Query().Get("password")

	if Verify(user, password) == false {
		ws.Write([]byte("user name or password is not right."))
		ws.Close()
	}

	apiSocket := longsocket.NewConn("", "", "", false, 128*1024)
	defer apiSocket.Close()
	go apiSocket.WriteLoop()
	go apiSocket.ReadLoop()
	apiSocket.Read(dealMsg)
}

func ReleaseHandle(res http.ResponseWriter, req *http.Request) {
	glog.Infof("release handle start")
}
