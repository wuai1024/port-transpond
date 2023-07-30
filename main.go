package main

import (
	"io"
	"log"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 4 {
		log.Fatalf("用法： %s [tcp/udp] [本地监听端口] [目标地址]", os.Args[0])
	}

	protocol := os.Args[1]      // 命令行参数中的协议（tcp/udp）
	localPort := os.Args[2]     // 命令行参数中的本地监听端口
	targetAddress := os.Args[3] // 命令行参数中的目标地址

	switch protocol {
	case "tcp":
		startTCPServer(localPort, targetAddress)
	case "udp":
		startUDPServer(localPort, targetAddress)
	default:
		log.Fatalf("不支持的协议：%s", protocol)
	}
}

func startTCPServer(localPort, targetAddress string) {
	listener, err := net.Listen("tcp", ":"+localPort)
	if err != nil {
		log.Fatalf("创建TCP监听器时发生错误：%s", err)
	}
	defer listener.Close()

	log.Printf("TCP端口转发服务器正在监听 :%s 并转发到 %s", localPort, targetAddress)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("接受TCP连接时发生错误：%s", err)
			continue
		}
		go handleTCPConnection(conn, targetAddress)
	}
}

func handleTCPConnection(conn net.Conn, targetAddress string) {
	defer conn.Close()

	// 连接到目标服务器
	targetConn, err := net.Dial("tcp", targetAddress)
	if err != nil {
		log.Printf("连接到目标服务器时发生错误：%s", err)
		return
	}
	defer targetConn.Close()

	// 开始在两个连接之间转发数据
	errChan := make(chan error, 2)

	go func() {
		_, err := io.Copy(conn, targetConn)
		errChan <- err
	}()
	go func() {
		_, err := io.Copy(targetConn, conn)
		errChan <- err
	}()

	err = <-errChan
	if err != nil {
		log.Printf("TCP数据转发时发生错误：%s", err)
	}
}

func startUDPServer(localPort, targetAddress string) {
	localUDPAddr, err := net.ResolveUDPAddr("udp", ":"+localPort)
	if err != nil {
		log.Fatalf("解析本地UDP地址时发生错误：%s", err)
	}

	targetUDPAddr, err := net.ResolveUDPAddr("udp", targetAddress)
	if err != nil {
		log.Fatalf("解析目标UDP地址时发生错误：%s", err)
	}

	conn, err := net.ListenUDP("udp", localUDPAddr)
	if err != nil {
		log.Fatalf("创建UDP监听器时发生错误：%s", err)
	}
	defer conn.Close()

	log.Printf("UDP端口转发服务器正在监听 %s 并转发到 %s", localUDPAddr.String(), targetUDPAddr.String())

	buffer := make([]byte, 65507)

	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("读取UDP数据时发生错误：%s", err)
			continue
		}

		_, err = conn.WriteToUDP(buffer[:n], targetUDPAddr)
		if err != nil {
			log.Printf("向目标地址转发UDP数据时发生错误：%s", err)
		}
		log.Printf("转发UDP数据：%d 字节从 %s 到 %s", n, clientAddr.String(), targetUDPAddr.String())
	}
}
