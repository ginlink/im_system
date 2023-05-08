package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

func (this *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	// close listen
	defer listener.Close()

	// 监听Message管道
	go this.ListenMessage()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept err:", err)
			continue
		}

		// TODO 尝试使用指针
		go this.Handler(conn)
	}
}

func (this *Server) ListenMessage() {
	for {
		sendMsg := <-this.Message

		this.mapLock.Lock()
		for _, user := range this.OnlineMap {
			user.C <- sendMsg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Name + "]" + ":" + msg

	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	user := NewUser(conn, this)

	user.Online()

	isLive := make(chan bool)

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			// 提取发送消息(去除\n)
			msg := string(buf[:n-1])

			user.DoMessage(msg)

			isLive <- true
		}
	}()

	// 阻塞当前handler，防止子协程exit
	for {
		select {
		case <-isLive:
		case <-time.After(time.Second * 30):
			fmt.Println("超时")
			// 已经超时
			user.SendMsg("超时未活动，断开连接！")

			// defer func() {
			// 	// 销毁用户资源
			// 	close(user.C)
			// 	// 关闭连接
			// 	conn.Close()
			// }()

			// 销毁用户资源
			close(user.C)
			// 关闭连接
			conn.Close()

			// 退出handler
			// runtime.Goexit()
			return
		}
	}
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}
