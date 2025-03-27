package service

import (
	"os"
	"path/filepath"
)

// LoadRoomConfig 加载房间密码配置
func LoadRoomConfig() {
	// check is Json

	exePath, err := os.Executable()
	if err != nil {
		exePath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	} else {
		exePath = filepath.Dir(exePath)
	}

	configPath := filepath.Join(exePath, ".room_pwd.json")
	roomManager = &JsonRoomManager{
		ConfigPath: configPath,
	}

	roomManager.Init()
}
