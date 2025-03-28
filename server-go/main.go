package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"internal-chat/server-go/service"
	"net/http"
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
	port := flag.String("port", "8082", "Port to listen on")
	flag.Parse()

	// 如果有命令行参数但没有使用flag格式，则第一个参数作为端口
	if flag.NArg() > 0 {
		*port = flag.Arg(0)
	}

	log.SetLevel(log.TraceLevel)

	// 加载房间密码配置
	service.LoadRoomConfig()

	// 设置WebSocket处理函数
	http.HandleFunc("/", service.HandleWebSocket)
	log.Infof("Signaling server running on ws://localhost:%s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
