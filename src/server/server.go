package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	maplock   sync.RWMutex
	// 消息广播的channel

	Message chan string
}

// 创建一个server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

// 监听message 广播channel的goroutine， 一旦有消息就发送给全部的在线user
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message

		// 将msg发送给全部的user
		this.maplock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.maplock.Unlock()
	}
}

func (this *Server) Handler(conn net.Conn) {
	// 当前连接的业务
	// 用户上线将用户加入到onlineMap中
	user := NewUser(conn, this)

	user.Online()
	isLive := make(chan bool)
	go func() {
		buf := make([]byte, 4096)

		for {
			read, err := conn.Read(buf)
			if err != nil && err != io.EOF {
				fmt.Printf("读取数据错误 + %s\n", conn.RemoteAddr().String())
				return
			}
			if read == 0 {
				user.Offline()
				return
			}

			msg := string(buf[:read-1])

			// 将得到的消息进行广播
			user.DoMessage(msg)
			isLive <- true
		}
	}()

	// 当前handler阻塞
	for {
		select {
		case <-isLive:
			//什么都不做，用户活跃，为了激活select下面的定时器
		case <-time.After(time.Second * 10):
			// 已经超时，将当前的User强制关闭
			user.SendMessage("你被踢了\n")
			close(user.C)
			err := conn.Close()
			if err != nil {
				fmt.Println("连接关闭失败")
				return
			}

			// 退出当前handler

			return

		}
	}
}

// 广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

// 启动服务端接口
func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net listen err: ", err)
		return
	}
	// close listen socket
	defer listener.Close()

	// 启动监听message的goroutine
	go this.ListenMessager()
	for {
		// accept
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("listener accept err : ", err)
			continue
		}

		//do handler
		go this.Handler(conn)
	}

}
