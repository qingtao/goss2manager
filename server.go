package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

const (
	// ss://${CRYPTOMOTHED}:${PASSWORD}@${ADDR}:${PORT}
	_srvStr = `ss://%s:%s@%s:%s`
)

// server 服务器
type server struct {
	wg     sync.WaitGroup
	cancel context.CancelFunc
	exit   chan struct{}
	c      chan string
}

// newServer 新建服务器，初始化exit
func newServer() *server {
	return &server{
		exit: make(chan struct{}),
		c:    make(chan string, 1),
	}
}

// url 根据method,password,ip,port生成命令行"-s"的url
func (s *server) url(method, password, ip, port string) string {
	return fmt.Sprintf(_srvStr, method, password, ip, port)
}

// logFile 日志文件
func (s *server) logFile(port string) string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "logs", fmt.Sprintf("go-shadowsocks2.%s.log", port))
}

// start 启动服务器
func (s *server) start(cfg *Config) {
	go s.output()
	s.run(cfg)
}

var (
	errNoCmd    = errors.New("未配置命令路径")
	errNoServer = errors.New("未配置服务")
	errNoPorts  = errors.New("未配置端口")
)

// run 启动服务器
func (s *server) run(cfg *Config) {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	if cfg.Cmd == "" {
		s.c <- errNoCmd.Error()
	}
	o := cfg.Option
	if o == nil {
		s.c <- errNoServer.Error()
	}
	if len(o.PortPassword) == 0 {
		s.c <- errNoPorts.Error()
	}
	for k, v := range o.PortPassword {
		args := []string{
			"-s",
			s.url(o.Method, v, o.IP, k),
			"-udptimeout",
			o.Timeout,
		}
		if cfg.Verbose {
			args = append(args, "-verbose")
		}
		cmd := exec.CommandContext(ctx, cfg.Cmd, args...)
		// 取标准输出
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			s.c <- err.Error()
		}
		// 取错误输出
		stderr, err := cmd.StderrPipe()
		if err != nil {
			s.c <- err.Error()
		}
		// 启动命令行
		if err := cmd.Start(); err != nil {
			s.c <- fmt.Sprintf("启动 %s %v 发生错误: %s", filepath.Base(cfg.Cmd), args, err)
		}
		s.wg.Add(1)
		go func(k string) {
			defer s.wg.Done()
			w := ioutil.Discard
			if cfg.Verbose {
				file, err := os.Create(s.logFile(k))
				if err != nil {
					s.c <- fmt.Sprintf("打开日志文件发生错误: %s", err)
					return
				}
				defer file.Close()
				w = file
			}
			// 写日志
			go io.Copy(w, stdout)
			go io.Copy(w, stderr)

			// 等待释放资源
			if err := cmd.Wait(); err != nil {
				log.Printf("子进程(%s)退出%s", k, err)
				return
			}
		}(k)
	}
	s.c <- "程序已启动"
	// 等待全部进程退出
	go s.wait()
}

// wait 等待服务器退出
func (s *server) wait() {
	s.wg.Wait()
	// 关闭退出通道
	close(s.exit)
	log.Println("程序终止")
}

// quit 服务器退出
func (s *server) quit() {
	s.cancel()
	<-s.exit
}

// output 打印信息
func (s *server) output() {
	for {
		s, ok := <-s.c
		if !ok {
			break
		}
		log.Printf("%s\n", s)
		return
	}
}
