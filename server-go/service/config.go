package service

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"strings"
)

type LogFileConfig struct {
	LogDir     string `json:"logDir"`
	Filename   string `json:"filename"`
	MaxSize    int    `json:"maxSize"`    // 每个日志文件的最大大小（MB）
	MaxBackups int    `json:"maxBackups"` // 保留的旧日志文件的最大数量
	MaxAge     int    `json:"maxAge"`     // 保留旧日志文件的最大天数
	Compress   bool   `json:"compress"`   // 是否压缩/归档旧日志文件
}

type ServerConfig struct {
	Port        int            `json:"port"`
	LogType     string         `json:"logType"` // console,file
	LogFile     *LogFileConfig `json:"logFile"`
	LogLevel    string         `json:"logLevel"` // panic,fatal,error,warn,info,debug,trace
	RoomPwdPath string         `json:"RoomPwdPath"`
}

var serverConfig *ServerConfig

func LoadConfig(path string, port int) {
	// default
	serverConfig = &ServerConfig{
		Port:    port,
		LogType: "file",
		LogFile: &LogFileConfig{
			LogDir:     "./logs",
			Filename:   "chat.log",
			MaxSize:    10,
			MaxBackups: 10,
			MaxAge:     10,
			Compress:   true,
		},
	}
	if path != "" {
		bytes, err := os.ReadFile(path)
		if err == nil {
			if strings.HasSuffix(path, ".json") {
				err := json.Unmarshal(bytes, serverConfig)
				if err != nil {
					log.Errorf("Load config error. path = %s, err: %v", path, err)
				}
			}
			// more config file type
		}
	}

	logConfig()

	// 加载房间密码配置
	if serverConfig.RoomPwdPath == "" {
		exePath, err := os.Executable()
		if err != nil {
			exePath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
		} else {
			exePath = filepath.Dir(exePath)
		}
		serverConfig.RoomPwdPath = filepath.Join(exePath, ".room_pwd.json")
	}
}

func logConfig() {
	level, err := log.ParseLevel(serverConfig.LogLevel)
	if err != nil {
		// default
		level = log.TraceLevel
	}
	log.SetLevel(level)

	// 根据 LogType 设置日志输出
	if serverConfig.LogType == "file" {
		logFileConfig := serverConfig.LogFile
		_, err := os.Stat(logFileConfig.LogDir)
		if os.IsNotExist(err) {
			// 创建文件夹
			err := os.MkdirAll(filepath.Dir(logFileConfig.LogDir), os.ModePerm)
			if err != nil {
				log.Warnf("create log dir error. path = %s, err: %v", logFileConfig.LogDir, err)
				return
			}
		} else if err != nil {
			log.Warnf("check log dir error. path = %s, err: %v", logFileConfig.LogDir, err)
			return
		}

		logFile := filepath.Join(logFileConfig.LogDir, logFileConfig.Filename)
		fmt.Println(logFile)

		log.SetOutput(&lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    logFileConfig.MaxSize,    // 每个日志文件的最大大小（MB）
			MaxBackups: logFileConfig.MaxBackups, // 保留的旧日志文件的最大数量
			MaxAge:     logFileConfig.MaxAge,     // 保留旧日志文件的最大天数
			Compress:   logFileConfig.Compress,   // 是否压缩/归档旧日志文件
		})
	}
}
