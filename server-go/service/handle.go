package service

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"strings"
	"time"
)

// 消息类型常量
const (
	// SendTypeReg 发送类型
	SendTypeReg            = "1001" // 注册后发送用户id
	SendTypeRoomInfo       = "1002" // 发送房间信息
	SendTypeJoinedRoom     = "1003" // 加入房间后的通知
	SendTypeNewCandidate   = "1004" // offer
	SendTypeNewConnection  = "1005" // new connection
	SendTypeConnected      = "1006" // new connection
	SendTypeNicknameUpdate = "1007" // 昵称更新通知

	// ReceiveTypeNewCandidate 接收类型
	ReceiveTypeNewCandidate   = "9001" // offer
	ReceiveTypeNewConnection  = "9002" // new connection
	ReceiveTypeConnected      = "9003" // joined
	ReceiveTypeKeepAlive      = "9999" // keep-alive
	ReceiveTypeUpdateNickname = "9004" // 更新昵称请求
)
const MaxMessageSize = 10 * 1024
const KeepAliveInterval = 30 * 1000

// Message 消息结构
type Message struct {
	Type     string          `json:"type"`
	UID      string          `json:"uid"`
	TargetID string          `json:"targetId"`
	Data     json.RawMessage `json:"data"`
}

// SendMessage 发送消息结构
type SendMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// WebSocket升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有跨域请求
	},
}
var roomManager RoomManager

// HandleWebSocket WebSocket连接处理
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Debugf("Url: %s", r.URL)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorln("Failed to upgrade connection:", err)
		return
	}
	defer conn.Close()

	ip := getRequestClientIp(r)

	// 解析URL路径获取房间ID和密码
	pathParts := strings.Split(r.URL.Path, "/")
	var roomId, pwd string

	if len(pathParts) > 1 && pathParts[1] != "" && len(pathParts[1]) <= 32 {
		roomId = strings.TrimSpace(pathParts[1])
	}
	if len(pathParts) > 2 && pathParts[2] != "" && len(pathParts[2]) <= 32 {
		pwd = strings.TrimSpace(pathParts[2])
	}

	var turns []TurnInfo
	// 兼容旧版本
	if roomId == "ws" || roomId == "" {
		roomId = ""
	} else {
		// 验证房间密码
		roomConfig, exists := roomManager.GetByRoomId(roomId)
		if !exists || pwd == "" || strings.ToLower(roomConfig.Pwd) != strings.ToLower(pwd) {
			roomId = ""
		} else {
			turns = roomConfig.Turns
		}
	}

	// 注册用户
	currentId := RegisterUser(ip, roomId, conn, func(key string) (*http.Cookie, error) {
		return r.Cookie(key)
	})

	// 向客户端发送自己的ID
	socketSendUserId(conn, currentId, roomId, turns)

	log.Infof("%s@%s. roomId = [%s] connected", currentId, ip, roomIdString(roomId))

	// 向所有用户发送房间信息
	for _, user := range GetUserList(ip, roomId) {
		socketSendRoomInfo(user.Socket, ip, roomId)
	}

	// 通知用户已加入房间
	socketSendJoinedRoom(conn, currentId)

	// 处理消息
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		if len(msg) == 0 {
			continue
		}
		// 消息长度限制
		if len(msg) > MaxMessageSize {
			// close connection
			break
		}

		// 解析消息
		var message Message
		if err := json.Unmarshal(msg, &message); err != nil {
			log.Debugln("Invalid JSON:", string(msg))
			continue
		}

		// 验证消息
		if message.Type == "" {
			continue
		}
		if message.Type == ReceiveTypeKeepAlive {
			_ = conn.SetReadDeadline(time.Now().Add(KeepAliveInterval * time.Millisecond))
			continue
		}

		if message.UID == "" || message.TargetID == "" {
			continue
		}

		// 获取发送者和接收者
		me := GetUser(ip, roomId, message.UID)
		target := GetUser(ip, roomId, message.TargetID)
		if me == nil || target == nil {
			continue
		}

		// 处理不同类型的消息
		switch message.Type {
		case ReceiveTypeNewCandidate:
			var data struct {
				Candidate json.RawMessage `json:"candidate"`
			}
			if err := json.Unmarshal(message.Data, &data); err != nil {
				continue
			}
			socketSendCandidate(target.Socket, message.UID, data.Candidate)

		case ReceiveTypeNewConnection:
			var data struct {
				TargetAddr json.RawMessage `json:"targetAddr"`
			}
			if err := json.Unmarshal(message.Data, &data); err != nil {
				continue
			}
			socketSendConnectInvite(target.Socket, message.UID, data.TargetAddr)

		case ReceiveTypeConnected:
			var data struct {
				TargetAddr json.RawMessage `json:"targetAddr"`
			}
			if err := json.Unmarshal(message.Data, &data); err != nil {
				continue
			}
			socketSendConnected(target.Socket, message.UID, data.TargetAddr)

		case ReceiveTypeUpdateNickname:
			var data struct {
				Nickname string `json:"nickname"`
			}
			if err := json.Unmarshal(message.Data, &data); err != nil {
				continue
			}

			success := UpdateNickname(ip, roomId, message.UID, data.Nickname)
			if success {
				// 通知所有用户昵称更新
				for _, user := range GetUserList(ip, roomId) {
					socketSendNicknameUpdated(user.Socket, message.UID, data.Nickname)
				}
			}
		}
	}

	// 用户断开连接
	UnregisterUser(ip, roomId, currentId)
	for _, user := range GetUserList(ip, roomId) {
		socketSendRoomInfo(user.Socket, ip, roomId)
	}
	log.Infof("%s@%s. roomId = [%s] disconnected", currentId, ip, roomIdString(roomId))
}

// 辅助函数：格式化房间ID字符串
func roomIdString(roomId string) string {
	if roomId == "" {
		return ""
	}
	return "/" + roomId
}

// 发送消息的辅助函数
func send(conn *websocket.Conn, msgType string, data interface{}) {
	msg := SendMessage{
		Type: msgType,
		Data: data,
	}
	if err := conn.WriteJSON(msg); err != nil {
		log.Errorln("Failed to send message:", err)
	}
}

// 发送用户ID
func socketSendUserId(conn *websocket.Conn, id, roomId string, turns []TurnInfo) {
	send(conn, SendTypeReg, map[string]interface{}{
		"id":     id,
		"roomId": roomId,
		"turns":  turns,
	})
}

// 发送房间信息
func socketSendRoomInfo(conn *websocket.Conn, ip, roomId string) {
	users := GetUserList(ip, roomId)
	result := make([]map[string]string, 0, len(users))
	for _, user := range users {
		result = append(result, map[string]string{
			"id":       user.ID,
			"nickname": user.Nickname,
		})
	}
	send(conn, SendTypeRoomInfo, result)
}

// 发送加入房间通知
func socketSendJoinedRoom(conn *websocket.Conn, id string) {
	send(conn, SendTypeJoinedRoom, map[string]string{
		"id": id,
	})
}

// 发送候选信息
func socketSendCandidate(conn *websocket.Conn, targetId string, candidate json.RawMessage) {
	send(conn, SendTypeNewCandidate, map[string]interface{}{
		"targetId":  targetId,
		"candidate": candidate,
	})
}

// 发送连接邀请
func socketSendConnectInvite(conn *websocket.Conn, targetId string, offer json.RawMessage) {
	send(conn, SendTypeNewConnection, map[string]interface{}{
		"targetId": targetId,
		"offer":    offer,
	})
}

// 发送已连接通知
func socketSendConnected(conn *websocket.Conn, targetId string, answer json.RawMessage) {
	send(conn, SendTypeConnected, map[string]interface{}{
		"targetId": targetId,
		"answer":   answer,
	})
}

// 发送昵称更新通知
func socketSendNicknameUpdated(conn *websocket.Conn, id, nickname string) {
	send(conn, SendTypeNicknameUpdate, map[string]string{
		"id":       id,
		"nickname": nickname,
	})
}

func getRequestClientIp(r *http.Request) string {
	// 获取客户端IP

	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For 格式：client, proxy1, proxy2
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 2. 尝试从 X-Real-Ip 获取
	xRealIP := r.Header.Get("X-Real-Ip")
	if xRealIP != "" {
		return xRealIP
	}

	// 3. 尝试从 Cloudflare 特定头部获取
	cfConnectingIP := r.Header.Get("CF-Connecting-IP")
	if cfConnectingIP != "" {
		return cfConnectingIP
	}

	// 4. 回退到 RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}

	if strings.HasPrefix(ip, "[") && strings.HasSuffix(ip, "]") {
		return ip[1 : len(ip)-1]
	}

	return ip
}
