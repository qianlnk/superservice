package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"qianno.xie/superservice/longsocket"
)

type Service struct {
	Name        string `json:"Name"`
	Command     string `json:"Command"`
	Directory   string `json:"Directory"`
	User        string `json:"User"`
	AutoStart   bool   `json:"AutoStart"`
	AutoRestart bool   `json:"AutoRestart"`
}
type Machine struct {
	Name        string    `json:"Name", "MachineName"`
	Host        string    `json:"Host"`
	Port        string    `json:"Port"`
	ServiceList []Service `json:"ServiceList"`
	Ls          *longsocket.Longsocket
	CloseConn   chan bool
}

type Machines_conf struct {
	MachineList []Machine
}

type Machines map[string]*Machine

var (
	MachineList Machines
)

func init() {
	cfg, err := config()
	if err != nil {
		return
	}
	initMachineList(cfg)
	for k, v := range MachineList {
		fmt.Println(k, v.Host, v.Port, v.ServiceList)
	}
}

func GetServiceMachineList() Machines {
	return MachineList
}

//函数功能：读取配置文件
func config() (*Machines_conf, error) {
	var cfg Machines_conf
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func initMachineList(cfg *Machines_conf) {
	MachineList = make(map[string]*Machine)
	for _, v := range cfg.MachineList {
		v.CloseConn = make(chan bool)
		MachineList[v.Name] = &v
	}
}

func (m Machines) GetAllMachines() []string {
	var res []string
	for k, _ := range m {
		res = append(res, k)
	}
	return res
}

func main() {
	fmt.Println("test")
}
