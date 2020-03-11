package core

import (
	"bufio"
	"fmt"
	"github.com/dp1993132/cron-node/m/v2/util"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var CronNode *cron.Cron
var entryIDS []cron.EntryID
var mutex sync.Mutex

func init() {
	CronNode = cron.New(cron.WithSeconds())
}

func LoadTask() {
	fl, err := util.GetTaskConfigFile(os.O_RDONLY | os.O_CREATE)
	if err != nil {
		log.Println("读取任务列表失败")
		return
	}
	log.Println("读取任务列表")

	sc := bufio.NewScanner(fl)
	for sc.Scan() {
		line := sc.Text()

		err := AddTask(line)

		if err != nil {
			continue
		}
	}
}

func ReloadTask() {
	log.Println("重新加载任务列表")
	CronNode.Stop()
	defer CronNode.Start()

	for _, id := range entryIDS {
		CronNode.Remove(id)
	}
	LoadTask()
}

func ParseLine(line string) ([]string, error) {
	var err error
	defer func() {
		e := recover()
		if e != nil {
			err = fmt.Errorf("parse line error")
		}
	}()

	arr := strings.Split(line, " ")
	spec := strings.Join(arr[:6], " ")
	cmd := strings.Join(arr[6:], " ")
	return []string{spec, cmd}, err
}

func parseCMD(cmdstr string) (cron.FuncJob, error) {
	var err error
	defer func() {
		e := recover()
		if e != nil {
			err = fmt.Errorf("parse cmd error")
		}
	}()

	arr := strings.Split(cmdstr, " ")
	return func() {
		cmd := exec.Command(arr[0], arr[1:]...)

		err := cmd.Run()
		if err != nil {
			log.Println(err)
		}
	}, err
}

func AddTask(line string) error {
	mutex.Lock()
	defer mutex.Unlock()

	res, err := ParseLine(line)
	if err != nil {
		return err
	}

	cmdFunc, err := parseCMD(res[1])
	if err != nil {
		return err
	}

	id, err := CronNode.AddFunc(res[0], cmdFunc)
	if err == nil {
		entryIDS = append(entryIDS, id)
		task := strings.Join(res, " ")
		log.Println("添加定时任务", task)
	}
	return err
}
