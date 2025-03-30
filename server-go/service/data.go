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
	/*	// TODO localhost
		if strings.EqualFold("127.0.0.1", ip) {
			return true
		}
		if strings.EqualFold("::1", ip) {
			return true
		}*/

	return false
}

type UserManager interface {
	RegisterUser(ip string, roomID string, socket *websocket.Conn, r *http.Request) string
	UnregisterUser(ip, roomID, id string) []User
	GetUserList(ip, roomID string) []User
	GetUser(ip, roomID, uid string) *User
	UpdateNickname(ip, roomID, id, nickname string) bool
}

type SimpleUserManager struct {
	// 存储用户数据的映射
	data map[string][]User
}

func NewSimpleUserManager() *SimpleUserManager {
	return &SimpleUserManager{
		data: make(map[string][]User),
	}
}

func (manager *SimpleUserManager) RegisterUser(ip string, roomID string, socket *websocket.Conn, r *http.Request) string {
	key := GetKey(ip, roomID)

	// 如果房间不存在，创建房间
	if _, exists := manager.data[key]; !exists {
		manager.data[key] = []User{}
	}

	// 生成随机ID
	// rand.Seed(time.Now().UnixNano())
	id := fmt.Sprintf("%02d%03d", rand.Intn(100), time.Now().Nanosecond()/1000000)

	// 确保ID唯一
	for _, exists := manager.data[id]; exists; {
		id = fmt.Sprintf("%02d%03d", rand.Intn(100), time.Now().Nanosecond()/1000000)
		_, exists = manager.data[id]
	}

	// 获取昵称
	nickname := ""
	cookie, err := r.Cookie("nickname")
	if err == nil {
		nickname = cookie.Value
	}
	// 创建用户并添加到房间
	user := User{
		ID:       id,
		Socket:   socket,
		Targets:  make(map[string]interface{}),
		Nickname: nickname,
	}
	manager.data[key] = append(manager.data[key], user)

	return id
}
func (manager *SimpleUserManager) UnregisterUser(ip, roomID, id string) []User {
	key := GetKey(ip, roomID)
	room, exists := manager.data[key]
	if !exists {
		return nil
	}

	for i, user := range room {
		if user.ID == id {
			// 移除用户
			manager.data[key] = append(room[:i], room[i+1:]...)
			return []User{user}
		}
	}

	return nil
}
func (manager *SimpleUserManager) GetUserList(ip, roomID string) []User {
	key := GetKey(ip, roomID)
	room, exists := manager.data[key]
	if !exists {
		return []User{}
	}
	return room
}
func (manager *SimpleUserManager) GetUser(ip, roomID, uid string) *User {
	key := GetKey(ip, roomID)
	room, exists := manager.data[key]
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
func (manager *SimpleUserManager) UpdateNickname(ip, roomID, id, nickname string) bool {
	key := GetKey(ip, roomID)
	room, exists := manager.data[key]
	if !exists {
		return false
	}

	for i, user := range room {
		if user.ID == id {
			manager.data[key][i].Nickname = nickname
			return true
		}
	}

	return false
}
