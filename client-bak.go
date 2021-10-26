package main

import (
	"bufio"
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
	defer serverConn.Close()
	log.Println("succed dial to server-relay")

	buf := bufio.NewReader(connect)
	io.Copy(serverConn, buf)
	go io.Copy(connect, serverConn)
	//	b := make([]byte, 1024)

	/*	n, err := connect.Read(b)
		log.Println(b)
		if err != nil {
			return
		}
		//	time.Sleep(3 * time.Second)
		if b[0] == 0x05 {
			log.Println("starting to proxy....")

		/*	c := make([]byte, 1024)
			//	connect.Write([]byte{0x05, 0x00})
			for {
				n, err =serverConn.Write(b[:n])

				n, err = serverConn.Read(c)
				log.Println(c)
				if err != nil {
					log.Println("error:", err)
					return
				}
				if c[0] == 0x05 {
					log.Println("get rsp from remote socket5 server")
					break
				}
			}
	*/

	//	connect.Write(c[:n])
	//	log.Println(c)
	//	connect.Write([]byte{0x05, 0x00})

	//	go io.Copy(connect, serverConn)
	//	io.Copy(serverConn, connect)

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
