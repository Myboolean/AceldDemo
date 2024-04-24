package main

import (
	"fmt"
	"net"
)

type Client struct {
	ServerIP   string
	ServerPort int
	Name       string
	conn       net.Conn
}

func NewClient(severIp string, serverPort int) *Client {

	client := &Client{
		ServerIP:   severIp,
		ServerPort: serverPort,
	}

	// 连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", severIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial err : ", err)
		return nil
	}

	client.conn = conn
	// 返回对象
	return client
}

var serverIP string
var serverPort int

func init() {

}

func main() {
	client := NewClient("127.0.0.1", 8888)

	if client == nil {
		fmt.Println("连接服务器失败")
		return
	}

	fmt.Println("连接服务器成功")

	// 启动客户端的业务
	select {}
}
