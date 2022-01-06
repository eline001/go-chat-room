package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

// 定义一个去哪聚的 channel, 用户处理各个客户端读到的消息
var message = make(chan []byte)

// 定义一个结构体 userInfo, 用户存储每位处聊天室用户的信息
type userInfo struct {
	name    string
	C       chan []byte
	NewUser chan []byte // 用户广播用户进入如退出当前聊天室的信息
}

// 定义一个map, 用户存储里聊天室中所有在线的用户和用户信息
var onlineUsers = make(map[string]userInfo)

// 管家循环监听管道message, 然后将消息写入每个用户的消息channel中
func manager() {
	for {
		select {
		case msg := <-message:
			for _, v := range onlineUsers {
				v.C <- msg
			}
		}
	}
}

// HandleConnect 这个函数完成服务器对一个客户端的整体处理流程
func HandleConnect(conn net.Conn) {
	defer conn.Close()
	// 管道 overTime 用户处理超时
	overTime := make(chan bool)

	// 用户存储用户名信息
	buf1 := make([]byte, 4096)
	n, err := conn.Read(buf1)
	if err != nil {
		fmt.Println("nnnnnnnnnnnnnnnnnnn",n)
		log.Println("conn.ReadErr", err)
		return
	}
	userName := string(buf1[:n]) // n-1 是为了去掉末尾的\n
	perC := make(chan []byte)
	perNerUser := make(chan []byte)
	user := userInfo{name: userName, C: perC, NewUser: perNerUser}
	onlineUsers[conn.RemoteAddr().String()] = user
	fmt.Printf("用户[%s]注册成功\n", userName)
	_, err = conn.Write([]byte(fmt.Sprintf("???????????你好,%s 欢迎来到聊天室\n", userName)))
	if err != nil {
		log.Println("conn.Write err", err)
		return
	}
	// 广播通知所有人  遍历 map
	go func() {
		for _, v := range onlineUsers {
			v.NewUser <- []byte(fmt.Sprintf("用户[%s] 加入了聊天室\n", userName))
		}
	}()
	// 监听每个用户自己的channel
	go func() {
		for {
			select {
			case msg1 := <-user.NewUser:
				_, _ = conn.Write(msg1)
			case msg2 := <-user.C:
				_, _ = conn.Write(msg2)
			}
		}
	}()

	// 循环读取客户端发来的消息
	go func() {
		buf2 := make([]byte, 4096)
		for {
			n, err := conn.Read(buf2)
			fmt.Printf("%s 说:%s", onlineUsers[conn.RemoteAddr().String()].name, string(buf2))
			// 用户存储当前服务器通信的客户端上的那个同户名
			thisUser := onlineUsers[conn.RemoteAddr().String()].name
			switch {
			case n == 0:
				fmt.Println(conn.RemoteAddr(), "已断开连接")
				for _, v := range onlineUsers {
					if thisUser != "" {
						v.NewUser <- []byte(fmt.Sprintf("???用户[%s] 已退出当前聊天\n", v.name))
					}
				}
				delete(onlineUsers, conn.RemoteAddr().String())
				return
			case string(buf2[:n]) == "who\n":
				_, _ = conn.Write([]byte("当前在线用户：\n"))
				for _, v := range onlineUsers {
					_, _ = conn.Write([]byte(fmt.Sprintf("????%s???\n", v.name)))
				}
			case len(string(buf2[:n])) > 7 && string(buf2[:n])[:7] == "rename|":
				// n-1 去掉buf2 里的空格
				onlineUsers[conn.RemoteAddr().String()] = userInfo{name: string(buf2[:n-1])[7:], C: perC, NewUser: perNerUser}
				_, _ = conn.Write([]byte("您已成功修改用户名！\n"))
			}

			if err != nil {
				fmt.Println("conn.Read error:", err)
				return
			}

			var msg []byte
			if buf2[0] != 10 && string(buf2[:n]) != "who\n" {
				if len(string(buf2[:n])) <= 7 || string(buf2[:n])[:7] != "rename|" {
					msg = append([]byte("["+thisUser+"]对大家说:"), buf2[:n]...)
				}
			} else {
				msg = nil
			}
			//
			overTime <- true
			message <- msg
		}
	}()

	// 等待消息
	for {
		select {
		case <-overTime:
		case <-time.After(time.Second * 3600):
			_, _ = conn.Write([]byte("抱歉，由于长时间未发送聊天内容，您已被系统踢出"))
			thisUser := onlineUsers[conn.RemoteAddr().String()].name
			for _, v := range onlineUsers {
				if thisUser != "" {
					v.NewUser <- []byte("????用户[" + thisUser + "]由于长时间未发送消息已被踢出当前聊天室\n")
				}
			}
			delete(onlineUsers, conn.RemoteAddr().String())
			return
		}
	}

}

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:8011")
	if err != nil {
		log.Fatal("net.Listen err:", err)
	}
	fmt.Println("聊天室-服务器已启动")
	fmt.Println("正在监听客户端连接请求。。。。")

	// 启动管家go程序，  不断监听全局channel-----message
	go manager()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("listener.Accept error:", err)
			continue
		}
		fmt.Printf("地址为[%v]的客户端已经连接成功\n", conn.RemoteAddr())
		go HandleConnect(conn)
	}

}
