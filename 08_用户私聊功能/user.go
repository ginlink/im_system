package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

func (this *User) listenMessage() {
	// 阻塞读取管道消息
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}

func (this *User) Online() {
	// 加入用户列表
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// 广播用户上线消息
	this.server.BroadCast(this, "已上线")
}

func (this *User) Offline() {
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	this.server.BroadCast(this, "下线")
}

// 广播本用户消息
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			sendMsg := "[" + user.Name + "]:" + "在线..."
			this.SendMsg(sendMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[0:7] == "rename|" {
		// rename|xxx
		newName := strings.Split(msg, "|")[1]
		_, ok := this.server.OnlineMap[newName]

		if ok {
			this.SendMsg("当前用户名已被使用")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("更改用户名成功:" + this.Name)
		}
	} else if len(msg) > 3 && msg[0:3] == "to|" {
		// to|xxx|msg

		targetName := strings.Split(msg, "|")[1]
		targetMsg := strings.Split(msg, "|")[2]
		target, ok := this.server.OnlineMap[targetName]

		if !ok {
			fmt.Printf("User not exist: %s\n", targetName)
			return
		}

		target.SendMsg(targetMsg)
	} else {
		this.server.BroadCast(this, msg)
	}
}

// 发送消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg + "\n"))
}

// 通过管道发送消息
func (this *User) SendMsgByChan(msg string) {
	this.C <- msg
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}

	go user.listenMessage()

	return user
}
