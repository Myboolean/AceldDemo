package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// 创建一个用户user

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	// 启动监听当前user channel消息的goroutine
	go ListenMessage(user)
	return user
}

// 监听当前User channel的方法
func ListenMessage(this *User) {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}

// 用户上线业务
func (this *User) Online() {
	this.server.maplock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.maplock.Unlock()
	this.server.BroadCast(this, "已上线")
}

// 用户下线业务

func (this *User) Offline() {
	this.server.maplock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.maplock.Unlock()
	this.server.BroadCast(this, "下线")
}

func (this *User) SendMessage(msg string) {
	_, err := this.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("send message err : ", err)
		return
	}

}

// 用户处理消息的业务

func (this *User) DoMessage(msg string) {
	// 查询在线用户
	if msg == "who" {
		// 查询当前在线用户都有哪些

		this.server.maplock.Lock()

		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线\n"
			this.SendMessage(onlineMsg)
		}

		this.server.maplock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename" {
		// 消息格式： rename 张三
		newName := strings.Split(msg, "|")[1]

		// 判断name是否存在
		this.server.maplock.Lock()
		_, ok := this.server.OnlineMap[newName]

		if ok {
			this.SendMessage("当前用户名被使用\n")
		} else {
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
		}
		this.server.maplock.Unlock()
	} else if len(msg) > 4 && msg[:3] == "to|" {
		userName := strings.Split(msg, "|")[1]
		if userName == "" {
			this.SendMessage("消息格式不正确， 请使用\"to|张三|你好啊\"格式\n")

		}
		remoteUser, ok := this.server.OnlineMap[userName]
		// 1.根据用户名得到当前user的对象
		if !ok {
			this.SendMessage("用户不存在\n")
			return
		}

		sendMsg := strings.Split(msg, "|")[2]
		if sendMsg == "" {
			this.SendMessage("无消息内容，请重发 \n")
			return
		}
		remoteUser.SendMessage(this.Name + "对您说" + sendMsg)
	} else {
		this.server.BroadCast(this, msg)
	}
}
