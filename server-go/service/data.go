package service

import (
	"fmt"
	"github.com/gorilla/websocket"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// User 用户结构体
type User struct {
	ID       string
	Socket   *websocket.Conn
	Targets  map[string]interface{}
	Nickname string
}

// 存储用户数据的映射
var data = make(map[string][]User)

// InternalNet 判断IP是否为内网IP
/*
  A类地址：10.0.0.0–10.255.255.255
  B类地址：172.16.0.0–172.31.255.255
  C类地址：192.168.0.0–192.168.255.255
*/
func InternalNet(ip string) bool {
	if strings.HasPrefix(ip, "10.") {
		return true
	}
	if strings.HasPrefix(ip, "172.") {
		parts := strings.Split(ip, ".")
		if len(parts) > 1 {
			second, err := strconv.Atoi(parts[1])
			if err == nil && second >= 16 && second <= 31 {
				return true
			}
		}
	}
	if strings.HasPrefix(ip, "192.168.") {
		return true
	}
	// TODO localhost
	if strings.EqualFold("127.0.0.1", ip) {
		return true
	}
	if strings.EqualFold("::1", ip) {
		return true
	}

	return false
}

// GetKey 获取房间键值
func GetKey(ip, roomID string) string {
	if roomID != "" {
		return roomID
	}
	isInternalNet := InternalNet(ip)
	if isInternalNet {
		return "internal"
	}
	return ip
}

// RegisterUser 注册用户
func RegisterUser(ip string, roomID string, socket *websocket.Conn, nickFn func(key string) (*http.Cookie, error)) string {
	key := GetKey(ip, roomID)

	// 如果房间不存在，创建房间
	if _, exists := data[key]; !exists {
		data[key] = []User{}
	}

	// 生成随机ID
	// rand.Seed(time.Now().UnixNano())
	id := fmt.Sprintf("%02d%03d", rand.Intn(100), time.Now().Nanosecond()/1000000)

	// 确保ID唯一
	for _, exists := data[id]; exists; {
		id = fmt.Sprintf("%02d%03d", rand.Intn(100), time.Now().Nanosecond()/1000000)
		_, exists = data[id]
	}

	// 获取昵称
	cn, err := nickFn("nickname")
	nickname := ""
	if err == nil {
		nickname = cn.Value
	}

	// 创建用户并添加到房间
	user := User{
		ID:       id,
		Socket:   socket,
		Targets:  make(map[string]interface{}),
		Nickname: nickname,
	}
	data[key] = append(data[key], user)

	return id
}

// UnregisterUser 注销用户
func UnregisterUser(ip, roomID, id string) []User {
	key := GetKey(ip, roomID)
	room, exists := data[key]
	if !exists {
		return nil
	}

	for i, user := range room {
		if user.ID == id {
			// 移除用户
			data[key] = append(room[:i], room[i+1:]...)
			return []User{user}
		}
	}

	return nil
}

// GetUserList 获取用户列表
func GetUserList(ip, roomID string) []User {
	key := GetKey(ip, roomID)
	room, exists := data[key]
	if !exists {
		return []User{}
	}
	return room
}

// GetUser 获取用户
func GetUser(ip, roomID, uid string) *User {
	key := GetKey(ip, roomID)
	room, exists := data[key]
	if !exists {
		return nil
	}

	for i, user := range room {
		if user.ID == uid {
			return &room[i]
		}
	}

	return nil
}

// UpdateNickname 更新用户昵称
func UpdateNickname(ip, roomID, id, nickname string) bool {
	key := GetKey(ip, roomID)
	room, exists := data[key]
	if !exists {
		return false
	}

	for i, user := range room {
		if user.ID == id {
			data[key][i].Nickname = nickname
			return true
		}
	}

	return false
}
