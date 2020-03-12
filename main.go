package main

import (
	"bufio"
	"bytes"
	"fmt"
	. "github.com/dp1993132/cron-node/m/v2/core"
	"github.com/dp1993132/cron-node/m/v2/util"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var rootCmd = &cobra.Command{
	Short: "定时调度程序",
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "启动定时任务调度器",
	Run: func(cmd *cobra.Command, args []string) {
		daemon()
	},
}

func daemon() {
	pid, err := util.GetPID()
	if err == nil && pid != 0 {
		_, err := os.FindProcess(pid)
		if err == nil {
			log.Println("进程已存在", pid)
			return
		}
	}

	util.SetPID(os.Getpid())
	defer util.RmPID()

	defer func() {
		err := recover()
		if err != nil {
			log.Println(err)
		}
	}()

	defer CronNode.Stop()
	LoadTask()
	CronNode.Start()

	var c = make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGUSR1)
sig:
	switch s := <-c; s {
	case syscall.SIGHUP, syscall.SIGINT:
		if !quit() {
			goto sig
		}
	case syscall.SIGUSR1:
		ReloadTask()
		goto sig
	default:
		goto sig
	}
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "添加定时任务",
	Long:  "添加一条定时调度任务，形如 * * * * * * cmd arg1 arg2 args...",
	Run: func(cmd *cobra.Command, args []string) {
		err := add(strings.Join(args, " "))
		if err != nil {
			log.Println(err)
		} else {
			log.Println("添加成功")
		}
	},
}

func sendReloadSig() error {
	pid, err := util.GetPID()
	if err != nil {
		return fmt.Errorf("cron task daemon 进程不存在")
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("cron task daemon 进程不存在")
	}

	err = p.Signal(syscall.SIGUSR1)
	if err != nil {
		return fmt.Errorf("cron task daemon 进程不存在")
	}
	return nil
}
func add(task string) error {

	if find(task) {
		return fmt.Errorf("任务已存在")
	}

	fl, err := util.GetTaskConfigFile(os.O_CREATE | os.O_APPEND | os.O_WRONLY)
	if err != nil {
		return err
	}
	defer fl.Close()

	_, err = fmt.Fprintln(fl, task)
	if err != nil {
		return err
	}

	err = sendReloadSig()
	if err != nil {
		return err
	}

	return err
}

func find(task string) bool {
	fl, err := util.GetTaskConfigFile(os.O_RDONLY)
	if err != nil {
		return false
	}
	defer fl.Close()

	sc := bufio.NewScanner(fl)
	for sc.Scan() {
		line := sc.Text()
		if line == task {
			return true
		}
	}
	return false
}

var lsCmd = &cobra.Command{
	Use:   "list",
	Short: "列出现有任务",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ls(); err != nil {
			log.Println("读取任务列表失败", err)
		}
	},
}

func ls() error {
	fl, err := util.GetTaskConfigFile(os.O_CREATE | os.O_RDONLY)
	if err != nil {
		return err
	}
	sc := bufio.NewScanner(fl)
	for sc.Scan() {
		fmt.Println(sc.Text())
	}
	return nil
}

var clsCmd = &cobra.Command{
	Use:   "clear",
	Short: "清空任务列表",
	Run: func(cmd *cobra.Command, args []string) {
		if err := util.RmTaskConfigFile(); err != nil {
			log.Println("清空失败")
		} else {
			if err := sendReloadSig(); err != nil {
				log.Println("重新加载任务失败")
			}
		}
	},
}

var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "删除指定任务",
	Run: func(cmd *cobra.Command, args []string) {
		if err := rm(args...); err != nil {
			log.Println("删除失败")
		} else {
			if err := sendReloadSig(); err != nil {
				log.Println("重新加载任务失败")
			}
		}
	},
}

func rm(tasks ...string) error {
	fr, err := util.GetTaskConfigFile(os.O_RDWR | os.O_CREATE)
	if err != nil {
		return err
	}

	tb := bytes.NewBuffer([]byte{})

	sc := bufio.NewScanner(fr)
	for sc.Scan() {
		line := sc.Text()
		for _, task := range tasks {
			if task != line {
				fmt.Fprintln(tb, line)
			}
		}
	}

	fr.Close()
	fw, err := util.GetTaskConfigFile(os.O_WRONLY | os.O_CREATE | os.O_TRUNC)
	if err != nil {
		return err
	}
	_, err = io.Copy(fw, tb)

	return err
}

func quit() bool {
ask:
	var char string
	fmt.Println("确实要退出吗?（y/n）")
	fmt.Scanf("%s\n", &char)
	switch char {
	case "y", "Y":
		log.Println("退出")
		return true
	case "n", "N":
		return false
	default:
		goto ask
	}
}

func main() {
	rootCmd.AddCommand(daemonCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(clsCmd)
	rootCmd.Execute()
}
