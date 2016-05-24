package main

type ServiceInfo struct {
	Name   string
	Status string
	Pid    int64
	Uptime string
}

type Services map[string]ServiceInfo

var ServiceList Services

const (
	RUNNING = "RUNNING"
	STOP    = "RUNNING"
)

func init() {
	ServiceList = make(map[string]ServiceInfo)
}

func (svc ServiceInfo) Running() {
	svc.Status = RUNNING
}

func (svc ServiceInfo) Stop() {
	svc.Status = STOP
}
