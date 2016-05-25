package service

import (
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
)

type Service struct {
	Name        string
	Command     string
	Directory   string
	User        string
	AutoStart   bool
	AutoRestart bool
	Status      string
	Cmd         *exec.Cmd
	StartTime   time.Time
	killSelf    bool
}

type Services map[string]*Service

var (
	Mu          sync.Mutex
	ServiceList Services
)

const (
	RUNNING = "RUNNING"
	STOP    = "STOP"
)

func init() {
	ServiceList = make(map[string]*Service)
	go ServiceList.AotoRestart()
}

func newService(name, command, dir, username string, start, restart bool) (*Service, error) {
	return &Service{
		Name:        name,
		Command:     command,
		Directory:   dir,
		User:        username,
		AutoStart:   start,
		AutoRestart: restart,
		Status:      STOP,
		Cmd:         nil,
		StartTime:   time.Now(),
		killSelf:    false,
	}, nil
}
func (svclst Services) UpdateService(name, command, dir, user string, start, restart bool) error {
	runbefore := false
	if v, ok := svclst[name]; ok {
		if v.Command != command || v.Directory != dir || v.User != user || v.AutoStart != start || v.AutoRestart != restart {
			if v.Status == RUNNING {
				runbefore = true
				err := svclst[name].Stop()
				if err != nil {
					return err
				}
			}
			service, err := newService(name, command, dir, user, start, restart)
			if err != nil {
				return err
			}
			svclst[name] = service
		}
	} else {
		service, err := newService(name, command, dir, user, start, restart)
		if err != nil {
			return err
		}
		svclst[name] = service
	}
	fmt.Println(name)
	if svclst[name].AutoStart == true || runbefore == true {
		return svclst[name].Start()
	}
	return nil
}

func (svclst Services) AotoRestart() {
	ticker := time.NewTicker(time.Second * time.Duration(20))
	for range ticker.C {
		for k, v := range svclst {
			if v.Status == STOP && v.AutoRestart {
				err := svclst[k].Start()
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}

func (svc *Service) Start() error {
	if svc.Status == RUNNING {
		return errors.New(fmt.Sprintf("%s already running.", svc.Name))
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
			fmt.Println("user", err)
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
		svc.Cmd = cmd
		err = svc.Cmd.Start()
		if err != nil {
			fmt.Println("start:", err)
			return
		}
		svc.Status = RUNNING
		svc.StartTime = time.Now()
		err = svc.Cmd.Wait()
		svc.Status = STOP
		svc.Cmd.Process.Release()
		if err != nil && svc.killSelf == false {
			fmt.Println("wait:", err)
			return
		}
	}()
	return nil
}

func (svc *Service) Stop() error {
	if svc.Status == STOP {
		return nil
	}
	err := svc.Cmd.Process.Kill()
	if err != nil {
		return err
	}
	svc.Status = STOP
	svc.killSelf = true
	return nil
}

func (svc *Service) Restart() error {
	err := svc.Stop()
	if err != nil {
		return err
	}
	return svc.Start()
}
