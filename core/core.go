package core

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
)

var CronNode *cron.Cron
var entryIDS []cron.EntryID
var mutex sync.Mutex
var logger = log.New(os.Stdout, "cron: ", log.LstdFlags)
var debugMode bool

func init(){
	flag.BoolVar(&debugMode,"debug",false,"是否显示debug信息")
	flag.Parse()

	if debugMode {
		CronNode = cron.New(cron.WithLogger(cron.VerbosePrintfLogger(logger)))
		log.SetOutput(os.Stdout)
	} else {
		CronNode = cron.New()
	}
}

func LoadTask(){
	fl,err:=GetTaskConfigFile()
	if err != nil {
		log.Println("读取任务列表失败")
		return
	}

	sc:=bufio.NewScanner(fl)
	for sc.Scan(){
		line:=sc.Text()

		err:= AddTask(line)

		if err != nil {
			continue
		}
	}
}

func ReloadTask() {
	log.Println("重新加载任务列表")
	CronNode.Stop()
	defer CronNode.Start()

	for _,id:=range entryIDS{
		CronNode.Remove(id)
	}
	LoadTask()
}

func ParseLine(line string)([]string,error){
	var err error
	defer func() {
		e:=recover()
		if e != nil {
			err = fmt.Errorf("parse line error")
		}
	}()

	arr:=strings.Split(line," ")
	spec := strings.Join(arr[0:5]," ")
	cmd := strings.Join(arr[5:], " ")
	return []string{spec,cmd},err
}

func parseCMD(cmdstr string) (cron.FuncJob,error){
	var err error
	defer func() {
		e:=recover()
		if e != nil {
			err = fmt.Errorf("parse cmd error")
		}
	}()

	arr:=strings.Split(cmdstr," ")
	return func() {
		cmd:=exec.Command(arr[0],arr[1:]...)

		err:=cmd.Run()
		if err != nil {
			log.Println(err)
		}
	},err
}

func AddTask(line string) error {
	mutex.Lock()
	defer mutex.Unlock()

	res,err:= ParseLine(line)
	if err != nil {
		return err
	}

	cmdFunc,err:= parseCMD(res[1])
	if err != nil {
		return err
	}

	id,err:=CronNode.AddFunc(res[0],cmdFunc)
	if err == nil {
		entryIDS = append(entryIDS,id)
		task:=strings.Join(res," ")
		log.Println("添加定时任务",task)
	}
	return err
}

func GetTaskConfigPath() string {
	//if cacheDir,err:= os.UserCacheDir();err == nil {
	//	return path.Join(cacheDir,"cron-task")
	//}
	if userHomeDir,err:= os.UserHomeDir();err == nil{
		return path.Join(userHomeDir,"cron-task")
	}
	return path.Join(path.Dir(os.Args[0]),".cron-task")
}

func GetTaskConfigFile()(*os.File,error) {
	tp:= GetTaskConfigPath()
	log.Println("读取任务列表",tp)
	return os.OpenFile(tp,os.O_CREATE|os.O_RDONLY,0644)
}

func WatchTaskList(ctx context.Context){
	wc,err:=fsnotify.NewWatcher()
	if err != nil {
		return
	}
	wc.Add(GetTaskConfigPath())
	for {
		select {
		case evt:=<-wc.Events:
			if evt.Op&fsnotify.Create == fsnotify.Create {
				ReloadTask()
			}
		case <-ctx.Done():
			break
		}
	}
}
