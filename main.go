package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	flag.Parse()
	if _cfgFile == "" {
		log.Println("未指定配置文件")
		return
	}
	cfg, err := ReadConfig(_cfgFile)
	if err != nil {
		log.Printf("读取配置文件失败: %s\n", err)
		return
	}

	s := newServer()
	s.start(cfg)

	// 监听系统信号
	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGQUIT,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGHUP,
	)
	for si := range c {
		switch si {
		case syscall.SIGHUP:
		default:
			s.quit()
			return
		}
	}
}
