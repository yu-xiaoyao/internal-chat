package service

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

// RoomConfig 房间密码配置
type RoomConfig struct {
	RoomID string     `json:"roomId"`
	Pwd    string     `json:"pwd"`
	Turns  []TurnInfo `json:"turns"`
	Remark string     `json:"remark"`
}
type TurnInfo struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username"`
	Credential string   `json:"credential"`
}

type RoomManager interface {
	Init()
	GetByRoomId(roomId string) (RoomConfig, bool)
}

var roomManager RoomManager

// region JsonRoomManager
type JsonRoomManager struct {
	ConfigPath string
	roomPwd    map[string]RoomConfig
}

func InitRoomManager() {
	initJsonRoomManager(serverConfig.RoomPwdPath)
}

func initJsonRoomManager(configPath string) {
	roomManager = &JsonRoomManager{
		ConfigPath: configPath,
	}
	roomManager.Init()
}

func (room *JsonRoomManager) Init() {
	log.Debugf("开始加载房间数据. path = %v", room.ConfigPath)

	file, err := os.ReadFile(room.ConfigPath)
	if err != nil {
		log.Errorf("Failed to load %v. err: %v", room.ConfigPath, err)
		return
	}

	var configs []RoomConfig
	if err := json.Unmarshal(file, &configs); err != nil {
		log.Errorf("Failed to parse %v. err: %v", room.ConfigPath, err)
		return
	}

	room.roomPwd = make(map[string]RoomConfig)

	var roomIds []string
	for _, config := range configs {
		roomIds = append(roomIds, config.RoomID)
		room.roomPwd[config.RoomID] = RoomConfig{
			Pwd:   config.Pwd,
			Turns: config.Turns,
		}
	}
	log.Infof("加载房间数据: %s", strings.Join(roomIds, ","))
}

func (room *JsonRoomManager) GetByRoomId(roomId string) (RoomConfig, bool) {
	roomConfig, exists := room.roomPwd[roomId]
	return roomConfig, exists
}

//endregion
