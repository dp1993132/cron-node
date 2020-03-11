package util

import (
	"fmt"
	"os"
	"path"
)

func GetTaskConfigPath() string {
	if cacheDir, err := os.UserCacheDir(); err == nil {
		return path.Join(cacheDir)
	}
	if userHomeDir, err := os.UserHomeDir(); err == nil {
		return path.Join(userHomeDir)
	}
	return path.Join(path.Dir(os.Args[0]))
}

func GetTaskConfigFile(flag int) (*os.File, error) {
	tp := path.Join(GetTaskConfigPath(), "cron-task")
	return os.OpenFile(tp, flag, 0644)
}

func RmTaskConfigFile() error {
	tp := path.Join(GetTaskConfigPath(), "cron-task")
	return os.Remove(tp)
}

func GetPID() (int, error) {
	pp := path.Join(GetTaskConfigPath(), ".cron-pid")
	fl, err := os.OpenFile(pp, os.O_RDONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer fl.Close()

	var pid int
	_, err = fmt.Fscanf(fl, "%d", &pid)

	return pid, err
}

func SetPID(pid int) error {
	pp := path.Join(GetTaskConfigPath(), ".cron-pid")
	fl, err := os.OpenFile(pp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fl.Close()

	fmt.Fprintf(fl, "%d", pid)

	return nil
}

func RmPID() error {
	pp := path.Join(GetTaskConfigPath(), ".cron-pid")
	return os.Remove(pp)
}
