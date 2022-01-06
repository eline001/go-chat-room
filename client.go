package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	fmt.Println("正在连接服务器。。。")
	conn, err := net.Dial("tcp", "127.0.0.1:8011")
	if err != nil {
		log.Fatal("net.Dial error :", err)
	}

	defer conn.Close()
	fmt.Printf("连接服务器成功\n")
	fmt.Println("先起一个名字吧:")

	var userName string
	// 使用 Scan 输入， 不允许出现空格
	_, err = fmt.Scan(&userName)
	if err != nil {
		log.Fatal("fmt.Scan Err", err)
	}
	_, _ = conn.Write([]byte(userName))

	buf2 := make([]byte, 4096)
	n, err := conn.Read(buf2)
	if err != nil {
		log.Fatal("conn.Read err", err)
	}

	// 客户端收到 提示信息后， 就可以发送聊天了
	fmt.Println(string(buf2[:n]))
	fmt.Println("⚠提示:长时间没有发送消息会被系统强制踢出")

	// 客户端发送消息到服务器
	go func() {
		for {
			buffer1 := make([]byte, 4096)
			// 这里使用 Stdin 标准输入， 因为scanf 无法识别空格
			n, err := os.Stdin.Read(buffer1)
			if n == 2 {
				continue
			}
			if err != nil {
				log.Printf("os.Stdin.Read error:%v", err)
				continue
			}
			_, _ = conn.Write(buffer1[:n])
		}
	}()

	// 接受服务器发来的数据
	for {
		buffer2 := make([]byte, 4096)
		n, err := conn.Read(buffer2)
		if n == 0 {
			fmt.Println("服务器已关闭当前连接，正在退出")
			return
		}
		if err != nil {
			fmt.Println("conn.Read error:", err)
			return
		}
		fmt.Print(string(buffer2[:n]))
	}

}
