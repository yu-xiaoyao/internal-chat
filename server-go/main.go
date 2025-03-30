package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"internal-chat/server-go/service"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		PadLevelText:    true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
}

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "", "Path to the configuration file")
	port := flag.Int("port", 8082, "Port to listen on")
	flag.Parse()
	// 加载配置
	service.LoadConfig(*configPath, *port)
	// 启动 web server
	service.StartWebServer()
}
