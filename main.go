package main

import (
	"fmt"
	. "github.com/dp1993132/cron-node/m/v2/core"
	"log"
	"os"
	"os/signal"
	"syscall"
)



func main(){
	defer CronNode.Stop()
	defer SaveTask()
	LoadTask()
	CronNode.Start()

	var c = make(chan os.Signal)
	signal.Notify(c,syscall.SIGINT)
	sig:
	<-c
	var char string
	ask:
	fmt.Println("确实要退出吗?（y/n）")
	fmt.Scanf("%s\n",&char)
	switch char {
	case "y","Y":
		log.Println("退出")
		return
	case "n","N":
		goto sig
	default:
		goto ask
	}
}