package core

import (
	"bufio"
	"fmt"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
)

var CronNode *cron.Cron
var taskList []string
var entryIDS []cron.EntryID
var mutex sync.Mutex
var logger = log.New(os.Stdout, "cron: ", log.LstdFlags)

func init(){
	CronNode = cron.New(cron.WithLogger(cron.VerbosePrintfLogger(logger)))
	log.SetOutput(os.Stdout)
}

func LoadTask(){
	tp:=getTaskConfigPath()
	fl,err:=os.OpenFile(tp,os.O_CREATE|os.O_RDONLY,0644)
	if err != nil {
		return
	}
	defer fl.Close()
	log.Println("定时服务配置于",tp)

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
	CronNode.Stop()
	defer CronNode.Start()

	for _,id:=range entryIDS{
		CronNode.Remove(id)
	}
	LoadTask()
}

func SaveTask(){
	tp:=getTaskConfigPath()
	fl,err:=os.OpenFile(tp,os.O_CREATE|os.O_WRONLY|os.O_TRUNC,0644)
	if err != nil {
		return
	}
	defer log.Println("保存定时任务于",tp,taskList)
	defer fl.Close()

	for _,line:= range taskList {
		if _,err:=fmt.Fprintln(fl,line);err!=nil {
			log.Println(err)
			continue
		}
	}
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
		cmd.Stdout = os.Stdout
		cmd.Stderr=os.Stderr

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
		taskList = append(taskList,task)
		log.Println("添加定时任务",task)
	}
	return err
}

func getTaskConfigPath() string {
	if cacheDir,err:= os.UserCacheDir();err == nil {
		return path.Join(cacheDir,"cron-task")
	}
	if userHomeDir,err:= os.UserHomeDir();err == nil{
		return path.Join(userHomeDir,"cron-task")
	}
	return path.Join(path.Dir(os.Args[0]),".cron-task")
}
