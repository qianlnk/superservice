package main

import (
	"fmt"
	"os/user"
	"time"
)

func main() {
	selfuser, err := user.Current()
	fmt.Println(selfuser.Username)
	if err != nil {
		fmt.Println(err)
	}
	if selfuser.Username != "root" {
		fmt.Println("superservice request to run as root.")
		return
	}

	err = ServiceList.UpdateService("lnk", "./hello/hello lnk", "", "root", true, true)
	if err != nil {
		fmt.Println(err)
	}
	time.Sleep(time.Second * 5)
	err = ServiceList.UpdateService("lnk", "./hello/hello newlnk", "", "root", true, true)
	if err != nil {
		fmt.Println(err)
	}
	err = ServiceList.UpdateService("xzj", "./hello/hello xzj", "", "xiezhenjia", true, true)
	if err != nil {
		fmt.Println(err)
	}
	for {
		time.Sleep(1 * time.Minute)
	}
}
