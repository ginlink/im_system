package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
}

func (this *User) listenMessage() {
	// 阻塞读取管道消息
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}

func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
	}

	go user.listenMessage()

	return user
}
