package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

var (
	// 配置文件路径
	_cfgFile string
)

// Config 配置
type Config struct {
	// go-shadowsocks2的命令行路径
	Cmd string `toml:"cmd"`
	// 是否启动详细模式
	Verbose bool `toml:"verbose"`
	// 服务器选项
	Option *Option `toml:"server"`
}

// Option 服务器配置
type Option struct {
	// 服务监听IP
	IP string `toml:"ip"`
	// 超时时间
	Timeout string `toml:"timeout"`
	// 加密方法
	Method string `toml:"method"`
	// 快速连接
	FastOpen bool `toml:"fast_open"`
	// 端口配置
	PortPassword map[string]string `toml:"port_password"`
}

// ReadConfig 读取toml格式的配置文件
func ReadConfig(file string) (*Config, error) {
	t, err := toml.LoadFile(file)
	if err != nil {
		return nil, fmt.Errorf("ReadConfig failed: %w", err)
	}

	var cfg Config
	if err = t.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("tree.Unmarshal failed: %w", err)
	}
	return &cfg, nil
}

func init() {
	flag.StringVar(&_cfgFile, "c", "conf/config.toml", "配置文件")
	wd, _ := os.Getwd()
	// 创建日志目录
	if err := os.MkdirAll(filepath.Join(wd, "logs"), 0755); err != nil {
		log.Printf("创建日志目录发生错误: %s", err)
	}
}
