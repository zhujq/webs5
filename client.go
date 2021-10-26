package main

import (
	"io"
	"log"
	"net"
)

type Socks5ProxyHandler struct{}

type Handler interface {
	Handle(connect net.Conn)
}

func (socks5 *Socks5ProxyHandler) Handle(connect net.Conn) {
	defer connect.Close()
	if connect == nil {
		return
	}

	serverConn, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		return
	}
	log.Println("succed dial to server-relay")
	b := make([]byte, 1024)

	n, err := connect.Read(b)
	if err != nil {
		return
	}

	if b[0] == 0x05 {
		serverConn.Write(b[:n])
	}
	go io.Copy(connect, serverConn)
	io.Copy(serverConn, connect)
}

func main() {
	ln, err := net.Listen("tcp", ":9990")
	if err != nil {
		log.Println("Error listening!", err)
		return
	}
	log.Println("This program is designed by zhujq for relay socket5,started listening 9990...")

	/*	if err != nil {
			log.Println("Error connect to remote websocket server:", err)
			return
		}
	*/
	//	log.Println("Succed to dail to remote websocket server")

	for true {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error accepting connection", err)
			continue
		}

		var handler Handler = new(Socks5ProxyHandler)

		go handler.Handle(conn)
		log.Println(conn)

	}

}
