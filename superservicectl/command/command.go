package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"qianno.xie/longsocket"
	"qianno.xie/superservice/superservicectl/service"
)

/*******************************************
local command:
1.
host
name	host	port
2.
service
hostname service.name1 service.command1....
hostname service.name2
3.
connect [name1 name2...]/all	default all
4.
disconnect [name1 name2...]/all no default, must be call
service command:
list list all services, include detail information
start servicename/all
stop servicename/all
restart servicename/all
delete servicename/all
update servicename/all
add servicename/all
log servicename tailf service's log
********************************************/
const (
	LOCAL_CMD_HOSTS      = "HOSTS"
	LOCAL_CMD_SERVICES   = "SERVICES"
	LOCAL_CMD_CONNECTS   = "CONNECTS"
	LOCAL_CMD_CONNECT    = "CONNECT"
	LOCAL_CMD_DISCONNECT = "DISCONNECT"
	LOCAL_CMD_RELEASE    = "RELEASE"
	LOCAL_CMD_EXIT       = "EXIT"
	SERVICE_CMD_LIST     = "LIST"
	SERVICE_CMD_START    = "START"
	SERVICE_CMD_STOP     = "STOP"
	SERVICE_CMD_RESTART  = "RESTART"
	SERVICE_CMD_DELETE   = "DELETE"
	SERVICE_CMD_UPDATE   = "UPDATE"
	SERVICE_CMD_ADD      = "ADD"
)

var CommandList []string

type Cmd struct {
	Type string `json: "Type"`
	service.Service
}

type Command struct {
	ConnectMachineList service.Machines
	ServiceMachineList service.Machines
}

var ctlCommand *Command

func init() {
	ctlCommand = new(Command)
	ctlCommand.ConnectMachineList = make(map[string]*service.Machine)
	ctlCommand.ServiceMachineList = service.GetServiceMachineList()
	CommandList = []string{"hosts", "services", "connects", "connect", "disconnect", "release", "list", "start", "stop", "restart", "delete", "update", "add", "log", "exit"}
}

func DealCommand(cmd string) {
	cmds := strings.Fields(cmd)
	if len(cmds) <= 0 {
		return
	}
	switch strings.ToUpper(cmds[0]) {
	case LOCAL_CMD_HOSTS:
		ctlCommand.showHosts()
		break
	case LOCAL_CMD_SERVICES:
		ctlCommand.showServices()
		break
	case LOCAL_CMD_CONNECTS:
		ctlCommand.showConnects()
		break
	case LOCAL_CMD_CONNECT:
		var tmpcmds []string
		tmpcmds = append(tmpcmds, cmds[1:]...)
		if len(tmpcmds) <= 0 {
			fmt.Println("ERROR: command 'connect' nead hostname as parameter but not found.")
			break
		}
		if strings.ToLower(tmpcmds[0]) == "all" {
			tmpcmds = service.MachineList.GetAllMachines()
		}
		ctlCommand.connect(tmpcmds...)
		break
	case LOCAL_CMD_DISCONNECT:
		var tmpcmds []string
		tmpcmds = append(tmpcmds, cmds[1:]...)
		if len(tmpcmds) <= 0 {
			fmt.Println("ERROR: command 'disconnect' nead hostname as parameter but not found.")
			break
		}
		if strings.ToLower(tmpcmds[0]) == "all" {
			tmpcmds = ctlCommand.ConnectMachineList.GetAllMachines()
		}
		ctlCommand.disconnect(tmpcmds...)
		break
	case LOCAL_CMD_RELEASE:
		ctlCommand.relaseVersion(cmds[1:]...)
		break
	default:
		ctlCommand.sendCommand(cmd)
	}
}
func show(fields []string, datas [][]string) {
	maxlen := make(map[int]int)
	for i, data := range datas {
		for j, dt := range data {
			if i == 0 {
				if len(dt) > len(fields[j]) {
					maxlen[j] = len(dt)
				} else {
					maxlen[j] = len(fields[j])
				}
			} else {
				if len(dt) > maxlen[j] {
					maxlen[j] = len(dt)
				}
			}
		}
	}
	line := "+"
	for i := 0; i < len(maxlen); i++ {
		if _, ok := maxlen[i]; ok {
			for j := 0; j <= maxlen[i]; j++ {
				line += "-"
			}
		}
		line += "+"
	}
	//	for _, v := range maxlen {
	//		for i := 0; i <= v; i++ {
	//			line += "-"
	//		}
	//		line += "+"
	//	}
	if line == "+" {
		return
	}
	fmt.Println(line)
	fmt.Printf("|")
	for i, f := range fields {
		format := "%-" + strconv.Itoa(maxlen[i]+1) + "s|"
		fmt.Printf(format, f)
	}
	fmt.Printf("\n")
	fmt.Println(line)
	count := 0
	for _, data := range datas {
		fmt.Printf("|")
		for i, dt := range data {
			format := "%-" + strconv.Itoa(maxlen[i]+1) + "s|"
			fmt.Printf(format, dt)
		}
		fmt.Printf("\n")
		count++
	}
	fmt.Println(line)
}
func (ctl *Command) showHosts() {
	fields := []string{"name", "host", "port"}
	var datas [][]string
	for _, v := range ctl.ServiceMachineList {
		var data []string
		data = append(data, v.Name)
		data = append(data, v.Host)
		data = append(data, v.Port)
		datas = append(datas, data)
	}
	show(fields, datas)
}

func (ctl *Command) showServices() {
	fields := []string{"hostname", "name", "version", "command", "directory", "user", "autostart", "autorestart"}
	var datas [][]string
	for _, v := range ctl.ServiceMachineList {
		for _, s := range v.ServiceList {
			var data []string
			data = append(data, v.Name)
			data = append(data, s.Name)
			data = append(data, s.Version)
			data = append(data, s.Command)
			data = append(data, s.Directory)
			data = append(data, s.User)
			data = append(data, fmt.Sprint(s.AutoStart))
			data = append(data, fmt.Sprint(s.AutoRestart))
			datas = append(datas, data)
		}
	}
	show(fields, datas)
}

func (ctl *Command) showConnects() {
	fields := []string{"name", "host", "port"}
	var datas [][]string
	for _, v := range ctl.ConnectMachineList {
		var data []string
		data = append(data, v.Name)
		data = append(data, v.Host)
		data = append(data, v.Port)
		datas = append(datas, data)
	}
	show(fields, datas)
}

func postFile(fileName string, targetUrl string) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile("releasefile", fileName)
	if err != nil {
		return err
	}

	f, err := os.Open(fileName)
	if err != nil {
		return err
	}

	_, err = io.Copy(fileWriter, f)
	if err != nil {
		return err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(targetUrl, contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)
	fmt.Println(string(resp_body))
	return nil
}

func (ctl *Command) relaseVersion(services ...string) {
	for _, m := range ctl.ConnectMachineList {
		targetUrl := fmt.Sprintf("http://%s:%s/Release", m.Host, m.Port)
		for _, s := range m.ServiceList {
			for _, sn := range services {
				if s.Name == sn {
					err := postFile(fmt.Sprintf("%s/%s.tar", s.Directory, s.Name), targetUrl)
					if err != nil {
						fmt.Println("err = ", err)
					}
				}
			}
		}
	}
}

func (ctl *Command) showResMessage(msg []byte, l *longsocket.Longsocket) {
	type list struct {
		Host    string
		Service string
		Status  string
		Pid     int
		User    string
		Command string
	}
	var resList list
	err := json.Unmarshal(msg, &resList)
	if err != nil {
		for _, v := range ctl.ConnectMachineList {
			if v.Ls == l {
				fmt.Println(v.Name, ":", string(msg))
				break
			}
		}
	} else {
		for _, v := range ctl.ConnectMachineList {
			if v.Ls == l {
				resList.Host = v.Name
				break
			}
		}
		fmt.Printf("host: %s\tservice:%s\t\t%s\t%d\t%s\t%s\n", resList.Host, resList.Service, resList.Status, resList.Pid, resList.User, resList.Command)
	}

}

func (ctl *Command) connect(machines ...string) {
	for _, m := range machines {
		if _, ok := ctl.ConnectMachineList[m]; ok {
			fmt.Printf("WARNING: host '%s' already connected.", m)
		}
		if v, ok := ctl.ServiceMachineList[m]; ok {
			go func() {
				errcount := 0
				for {
					wsAddr := fmt.Sprintf("ws://%s:%s/Cmd", v.Host, v.Port)
					httpAddr := fmt.Sprintf("http://%s:%s/Cmd?user=%s&password=%s", v.Host, v.Port, "qianlnk", "123456")
					v.Ls = longsocket.NewConn(wsAddr, "", httpAddr, true, 128*1024)
					err := v.Ls.Dial(true)
					if err != nil {
						errcount++
						fmt.Println("err:", err)
					}
					if v.Ls.Status == longsocket.STATUS_INIT {
						if errcount >= 2 {
							delete(ctl.ConnectMachineList, m)
							return
						}
						time.Sleep(2 * time.Second)
						continue
					}
					ctl.ConnectMachineList[m] = v

					reConn := make(chan bool)
					go func() {
						go v.Ls.ReadLoop()
						go v.Ls.WriteLoop()
						v.Ls.Read(dealResMessage)
						reConn <- false
						return
					}()

					select {
					case <-v.CloseConn:
						v.Ls.Close()
						return
					case <-reConn:
						close(reConn)
						break
					}
				}
			}()
			if _, ok := ctl.ConnectMachineList[m]; !ok {
				time.Sleep(3 * time.Second)
				if _, ok := ctl.ConnectMachineList[m]; !ok {
					fmt.Printf("host '%s' connection refused.\n", m)
				}
			}
		} else {
			fmt.Printf("host '%s' is not exist.\n", m)
		}
	}
}

func (ctl *Command) disconnect(machines ...string) {
	for _, m := range machines {
		if v, ok := ctl.ConnectMachineList[m]; ok {
			v.CloseConn <- true
			time.Sleep(2 * time.Nanosecond)
			delete(ctl.ConnectMachineList, m)
		}
	}
}

func (ctl *Command) sendCommand(cmd string) {
	cmds := strings.Fields(cmd)
	if len(cmds) != 2 {
		fmt.Printf("ERROR:command '%s' need one servicename as parameter only.\n", cmds[0])
		return
	}
	if strings.ToUpper(cmds[0]) == SERVICE_CMD_LIST {
		for _, v := range ctl.ConnectMachineList {
			if v.Name == cmds[1] || strings.ToLower(cmds[1]) == "all" {
				var cmdMsg Cmd
				cmdMsg.Type = cmds[0]
				msg, err := json.Marshal(cmdMsg)
				if err != nil {
					fmt.Println("ERROR:", err)
					continue
				}
				err = v.Ls.Write(msg)
				if err != nil {
					fmt.Println("ERROR:", err)
				}
			}
		}
		return
	}
	for _, v := range ctl.ConnectMachineList {
		for _, s := range v.ServiceList {
			if s.Name == cmds[1] {
				var cmdMsg Cmd
				cmdMsg.Type = cmds[0]
				cmdMsg.Name = s.Name
				cmdMsg.Command = s.Command
				cmdMsg.Directory = s.Directory
				cmdMsg.User = s.User
				cmdMsg.AutoStart = s.AutoStart
				cmdMsg.AutoRestart = s.AutoRestart

				msg, err := json.Marshal(cmdMsg)
				if err != nil {
					fmt.Println("ERROR:", err)
					break
				}
				err = v.Ls.Write(msg)
				if err != nil {
					fmt.Println("ERROR:", err)
				}
				break
			}
		}
	}

}

func dealResMessage(msg []byte, l *longsocket.Longsocket) error {
	if len(msg) == 0 {
		return nil
	}
	ctlCommand.showResMessage(msg, l)
	return nil
}
func main() {
	fmt.Println("test")
	DealCommand("hosts")
	DealCommand("services")
	DealCommand("connect testMachine1")
	time.Sleep(1 * time.Second)
	DealCommand("connects")
	//DealCommand("disconnect testMachine1")
	time.Sleep(1 * time.Second)
	DealCommand("connects")
	for {
		time.Sleep(2 * time.Second)
	}
}
