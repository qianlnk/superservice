package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"qianno.xie/superservice/superserviced/storage"
)

type Service struct {
	Name        string
	Version     string
	Command     string
	Directory   string
	User        string
	AutoStart   bool
	AutoRestart bool
	status      string
	cmd         *exec.Cmd
	startTime   time.Time
	killSelf    bool
}

type Services map[string]*Service

var (
	Mu          sync.Mutex
	ServiceList Services
	Message     chan string
)

const (
	RUNNING = "RUNNING"
	STOP    = "STOP"
)

func MSG_UPDATE(name string) string {
	return fmt.Sprintf("service %s updating", name)
}

func MSG_ADD(name string) string {
	return fmt.Sprintf("service %s adding", name)
}

func MSG_DELETE(name string) string {
	return fmt.Sprintf("service %s deleting", name)
}

func MSG_START(name string) string {
	return fmt.Sprintf("service %s starting", name)
}

func MSG_STOP(name string) string {
	return fmt.Sprintf("service %s stoping", name)
}

func MSG_OK(name string) string {
	return fmt.Sprintf("service %s ok", name)
}

func MSG_NO_CHANGE(name string) string {
	return fmt.Sprintf("service %s no change", name)
}

func MSG_ERROR(err error) string {
	return fmt.Sprintf("Err: %s", err.Error())
}

func MSG_EXIT(funcname string) string {
	return fmt.Sprintf("func %s exit", funcname)
}
func init() {
	ServiceList = make(map[string]*Service)
	ServiceList.ReadStorage()
	go ServiceList.AotoRestart()
}

func newService(name, version, command, dir, username string, start, restart bool) *Service {
	return &Service{
		Name:        name,
		Version:     version,
		Command:     command,
		Directory:   dir,
		User:        username,
		AutoStart:   start,
		AutoRestart: restart,
		status:      STOP,
		cmd:         nil,
		startTime:   time.Now(),
		killSelf:    false,
	}
}
func (svclst Services) ReadStorage() {
	bolt := storage.GetBolt("service")
	kvs, err := bolt.List()
	if err != nil {
		fmt.Println(err)
		return
	}
	msg := make(chan string)
	go discardMsg(msg)
	for _, v := range kvs {
		var service Service
		err := json.Unmarshal(v.Value, &service)
		if err != nil {
			continue
		}
		svclst.UpdateService(service.Name, service.Version, service.Command, service.Directory, service.User, service.AutoStart, service.AutoRestart, msg)
	}
}

func (svclst Services) UpdateService(name, version, command, dir, user string, start, restart bool, msg chan string) {
	bolt := storage.GetBolt("service")
	runbefore := false
	if v, ok := svclst[name]; ok {
		msg <- MSG_UPDATE(name)
		if v.Version != version || v.Command != command || v.Directory != dir || v.User != user || v.AutoStart != start || v.AutoRestart != restart {
			if v.status == RUNNING {
				runbefore = true
				err := svclst[name].Stop(msg)
				if err != nil {
					return
				}
			}
			svclst[name] = newService(name, version, command, dir, user, start, restart)
			bolt.Update(name, *svclst[name])
			msg <- MSG_OK(name)
		} else {
			msg <- MSG_NO_CHANGE(name)
		}
	} else {
		msg <- MSG_ADD(name)
		svclst[name] = newService(name, version, command, dir, user, start, restart)
		bolt.Put(name, *svclst[name])
		msg <- MSG_OK(name)
	}

	if svclst[name].AutoStart == true || runbefore == true {
		svclst[name].Start(msg)
	}
}

func discardMsg(msg chan string) {
	for {
		select {
		case m := <-msg:
			if strings.Contains(m, "exit") || strings.Contains(m, "stop ok") {
				close(msg)
				return
			}
		}
	}
}
func (svclst Services) AotoRestart() {
	ticker := time.NewTicker(time.Second * time.Duration(20))
	for range ticker.C {
		for k, v := range svclst {
			fmt.Println(k, v)
			if v.status == STOP && v.AutoRestart {
				msg := make(chan string)
				go discardMsg(msg)
				svclst[k].Start(msg)
			}
		}
	}
}

func (svclst Services) Close() {
	for k, _ := range svclst {
		msg := make(chan string)
		go discardMsg(msg)
		err := svclst[k].Stop(msg)
		if err != nil {
			fmt.Println(err)
		}
		delete(svclst, k)
	}
}

func (svclst Services) Delete(serviceName string, msg chan string) {
	if _, ok := svclst[serviceName]; ok {
		bolt := storage.GetBolt("service")
		msg <- MSG_DELETE(serviceName)
		err := svclst[serviceName].Stop(msg)
		if err != nil {
			return
		}
		svclst[serviceName] = nil
		delete(svclst, serviceName)
		bolt.Delete(serviceName)
		msg <- MSG_OK(serviceName)

	}
}
func (svclst Services) List(msg chan string) {
	for _, v := range svclst {
		//{"Name":"xzj","Command":"./hello/hello xzj","Directory":"","User":"xiezhenjia","AutoStart":true,"AutoRestart":true}
		var tmpmsg string
		if v.status == RUNNING {
			tmpmsg = fmt.Sprintf("{\"Service\":\"%s\",\"Status\":\"%s\",\"Pid\":%d,\"User\":\"%s\", \"Command\":\"%s\"}",
				v.Name, v.status, v.cmd.Process.Pid, v.User, v.Command)
		} else {
			tmpmsg = fmt.Sprintf("{\"Service\":\"%s\",\"Status\":\"%s\",\"Pid\":%d,\"User\":\"%s\", \"Command\":\"%s\"}",
				v.Name, v.status, 0, v.User, v.Command)
		}
		msg <- string(tmpmsg)
	}
}

func (svc *Service) Start(msg chan string) {
	msg <- MSG_START(svc.Name)
	if svc.status == RUNNING {
		msg <- MSG_ERROR(errors.New(fmt.Sprintf("%s already running.", svc.Name)))
		return
	}
	go func() {
		commands := strings.Fields(svc.Command)
		var arg []string
		for i := 1; i < len(commands); i++ {
			arg = append(arg, commands[i])
		}
		cmd := exec.Command(commands[0], arg...)
		runuser, err := user.Lookup(svc.User)
		if err != nil {
			msg <- MSG_ERROR(err)
			return
		}
		uid, _ := strconv.Atoi(runuser.Uid)
		gid, _ := strconv.Atoi(runuser.Gid)
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential = &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		}
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		svc.cmd = cmd
		err = svc.cmd.Start()
		if err != nil {
			msg <- MSG_ERROR(err)
			return
		}
		svc.status = RUNNING
		svc.startTime = time.Now()
		msg <- MSG_OK(svc.Name)
		err = svc.cmd.Wait()
		svc.status = STOP
		if err != nil && svc.killSelf == false {
			msg <- MSG_ERROR(err)
			return
		}
		msg <- MSG_EXIT("start")
	}()
}

func (svc *Service) Stop(msg chan string) error {
	msg <- MSG_STOP(svc.Name)
	if svc.status == STOP {
		msg <- MSG_OK(svc.Name)
		return nil
	}
	err := svc.cmd.Process.Kill()
	if err != nil {
		msg <- MSG_ERROR(err)
		return err
	}
	svc.status = STOP
	svc.killSelf = true
	msg <- MSG_OK(svc.Name + " stop")
	return nil
}

func (svc *Service) Restart(msg chan string) {
	err := svc.Stop(msg)
	if err != nil {
		return
	}
	svc.Start(msg)
}
