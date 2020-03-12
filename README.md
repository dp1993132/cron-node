# cron-node
定时任务执行器

# 使用
1. 启动程序会在用户目录创建任务列表文件
2. 修改任务列表可以添加定时任务

# 任务格式
```
* * * * * * cmd args...
秒 分 时 日 月 周 命令 参数1 ... 参数n

例如：* * * * * * get report.yottachain.net
每秒调用一次get report.yottachain.net一次
```

# 现有命令
```
Usage:
   [command]

Available Commands:
  add         添加定时任务
  clear       清空任务列表
  daemon      启动定时任务调度器
  help        Help about any command
  list        列出现有任务
  rm          删除指定任务

Flags:
  -h, --help   help for this command

```
