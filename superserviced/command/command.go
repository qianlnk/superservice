package command

import (
	"encoding/json"
	"fmt"
	//"net/http"
	"net/url"
	"strings"

	//"github.com/golang/glog"

	"golang.org/x/net/websocket"

	"github.com/qianlnk/longsocket"
	"github.com/qianlnk/superservice/superserviced/service"
)

type User struct {
	Name     string
	Password string
}

type Message struct {
}

//list start stop restart delete update add log
type Cmd struct {
	Type string `json: "Type"`
	service.Service
}

//var message chan string

//func init() {
//	message = make(chan string, 1)

//}

//exec command
func dealCommand(cmd Cmd, l *longsocket.Longsocket) {
	fmt.Println(cmd)
	message := make(chan string, 1)
	//defer close(message)
	go sendMessage(message, l)
	switch strings.ToUpper(cmd.Type) {
	case "ADD":
		service.ServiceList.UpdateService(cmd.Name, cmd.Version, cmd.Command, cmd.Directory, cmd.User, cmd.AutoStart, cmd.AutoRestart, message)
		break
	case "DELETE":
		service.ServiceList.Delete(cmd.Name, message)
		break
	case "UPDATE":
		service.ServiceList.UpdateService(cmd.Name, cmd.Version, cmd.Command, cmd.Directory, cmd.User, cmd.AutoStart, cmd.AutoRestart, message)
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

//deal the message for super service control client
func dealMsg(msg []byte, l *longsocket.Longsocket) error {
	fmt.Println("dealMsg", string(msg))
	if string(msg) == longsocket.SHAKE_HANDS_MSG || len(msg) == 0 {
		return nil
	}
	var cmd Cmd
	json.Unmarshal(msg, &cmd)
	dealCommand(cmd, l)
	return nil
}

//call func with gorouting, it will send result message to client
func sendMessage(msg chan string, l *longsocket.Longsocket) {
	fmt.Println("sendMessage")
	for {
		select {
		case m, ok := <-msg:
			if !ok {
				return
			}
			l.Write([]byte(m))
			fmt.Println(m)
		}
	}
}

//check user and password
func Verify(user, password string) bool {
	fmt.Println(user, password)
	return true
}

//accept connect for client
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
	apiSocket.SetSocket(ws)
	defer apiSocket.Close()
	go apiSocket.WriteLoop()
	go apiSocket.ReadLoop()
	apiSocket.Read(dealMsg)
}
