package command

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"qianno.xie/superservice/longsocket"
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
)

var connectMachineList service.Machines

func init() {
	connectMachineList = make(map[string]*service.Machine)
}
func KeepConnection() {
	for _, v := range service.MachineList {
		fmt.Println(v)
		go func() {
			for {
				wsAddr := fmt.Sprintf("ws://%s:%s/Cmd", v.Host, v.Port)
				httpAddr := fmt.Sprintf("http://%s:%s/Cmd?user=%s&password=%s", v.Host, v.Port, "qianlnk", "123456")
				v.Ls = longsocket.NewConn(wsAddr, "", httpAddr, true, 128*1024)
				err := v.Ls.Dial(true)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(123)
				go v.Ls.ReadLoop()
				go v.Ls.WriteLoop()
				<-v.CloseConn
				v.Ls.Close()
			}
		}()
	}
	for {
		time.Sleep(2 * time.Second)
	}
}

func DealCommand(cmd string) {
	cmds := strings.Fields(cmd)
	if len(cmds) < 0 {
		return
	}
	switch strings.ToUpper(cmds[0]) {
	case LOCAL_CMD_HOSTS:
		showHosts()
		break
	case LOCAL_CMD_SERVICES:
		showServices()
		break
	case LOCAL_CMD_CONNECTS:
		showConnects()
		break
	case LOCAL_CMD_CONNECT:
		var tmpcmds []string
		tmpcmds = append(tmpcmds, cmds[1:]...)
		connect(tmpcmds...)
		break
	case LOCAL_CMD_DISCONNECT:
		var tmpcmds []string
		tmpcmds = append(tmpcmds, cmds[1:]...)
		disconnect(tmpcmds...)
		break
	case LOCAL_CMD_RELEASE:
		break
	default:
		sendCommand(cmd)
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
	for _, v := range maxlen {
		for i := 0; i <= v; i++ {
			line += "-"
		}
		line += "+"
	}
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
func showHosts() {
	fields := []string{"name", "host", "port"}
	var datas [][]string
	for _, v := range service.MachineList {
		var data []string
		data = append(data, v.Name)
		data = append(data, v.Host)
		data = append(data, v.Port)
		datas = append(datas, data)
	}
	show(fields, datas)
}

func showServices() {
	fields := []string{"hostname", "name", "command", "directory", "user", "autostart", "autorestart"}
	var datas [][]string
	for _, v := range service.MachineList {
		for _, s := range v.ServiceList {
			var data []string
			data = append(data, v.Name)
			data = append(data, s.Name)
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

func showConnects() {
	fields := []string{"name", "host", "port"}
	var datas [][]string
	for _, v := range connectMachineList {
		var data []string
		data = append(data, v.Name)
		data = append(data, v.Host)
		data = append(data, v.Port)
		datas = append(datas, data)
	}
	show(fields, datas)
}

func connect(machines ...string) {
	for _, m := range machines {
		if v, ok := service.MachineList[m]; ok {
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
							return
						}
						time.Sleep(2 * time.Second)
						continue
					}
					connectMachineList[m] = v

					reConn := make(chan bool)
					go func() {
						go v.Ls.ReadLoop()
						v.Ls.WriteLoop()
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
			if _, ok := connectMachineList[m]; !ok {
				time.Sleep(3 * time.Second)
				if _, ok := connectMachineList[m]; !ok {
					fmt.Printf("host '%s' connection refused.\n", m)
				}
			}
		} else {
			fmt.Printf("host '%s' is not exist.\n", m)
		}
	}
}

func disconnect(machines ...string) {
	for _, m := range machines {
		if v, ok := connectMachineList[m]; ok {
			v.CloseConn <- true
			time.Sleep(2 * time.Nanosecond)
			delete(connectMachineList, m)
		}
	}
}

func sendCommand(cmd string) {}

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
