package main

import (
	"bufio"
	"crypto/tls"
	"io"
	"log"
	"math/rand"
	"net"
	"strings"
)

const proxyDomain = "zjqzjq2018-ray.run.goorm.io"
const port = "9990"

var letters = []rune("abcdefghijklmnopqrstuvwyz1234567890")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func handleConnection(clientConn net.Conn) {
	defer clientConn.Close()

	conf := &tls.Config{InsecureSkipVerify: true}

	destSvr := proxyDomain

	if strings.Contains(destSvr, ":") == false { //地址信息不含端口号
		destSvr += ":"
	}
	if strings.HasSuffix(destSvr, ":") { //默认443端口
		destSvr += "443"
	}

	serverSend, err := tls.Dial("tcp", destSvr, conf)
	if err != nil {
		log.Println("Failed to connect to send proxy up server!")
		return
	}

	serverListen, err := tls.Dial("tcp", destSvr, conf)
	if err != nil {
		log.Println("Failed to connect to listen proxy down server!")
		return
	}
	log.Println("Succes dail to dual-way ssl server")
	defer serverListen.Close()
	defer serverSend.Close()

	clientId := randSeq(20)

	b := make([]byte, 1024)
	n, err := clientConn.Read(b)
	log.Println(b[:n])
	wait := make(chan bool)

	go func() {
		log.Println("starting to post up http")

		_, err = serverSend.Write([]byte("GET /transmit HTTP/1.1\r\n" + "Host: " + proxyDomain + "\r\n" + "Accept: */*\r\n" + "Upgrade: websocket\r\n" + "Connection: Upgrade\r\n" + "Clientid: " + clientId + "\r\n" + "Connection: keep-alive\r\n" + "Sec-WebSocket-Version: 13\r\n" + "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n" + "\r\n"))
		if err != nil {
			log.Println(proxyDomain+":Error write to serversend", err)
		}

		buf := bufio.NewReader(serverSend)
		success := false
		//	log.Println(buf)
		for line, err := buf.ReadString('\n'); true; line, err = buf.ReadString('\n') {
			log.Println(line)
			if err != nil {
				log.Println("error:", err)
				log.Println(proxyDomain + ":Failed to read following lines")
				return
			}
			if line == "HTTP/1.1 101 Switching Protocols\r\n" {
				success = true
				log.Println("succed get rsp")
			}

			if line == "\r\n" {
				break
			}
		}
		//	fmt.Fprintf(serverSend,
		//		"Content-Type: multipart/form-data; boundary=----------SWAG------BOUNDARY----\r\n")
		// fmt.Fprintf(serverSend, "Transfer-Encoding: chunked\r\n")
		//	fmt.Fprintf(serverSend, "Content-Length: 12345789000\r\n\r\n")
		//	fmt.Fprintf(serverSend, "----------SWAG------BOUNDARY----\r\n")

		if success && b[0] == 0x05 {
			log.Println("entering up syn process...")
			serverSend.Write(b[:n])
			_, err = io.Copy(serverSend, clientConn)
			if err != nil {
				log.Println(proxyDomain+":Error copying client to server stream", err)
			}
		} else {
			log.Println(proxyDomain + ":Failed to bind send connection!")
		}

		wait <- true
	}()

	go func() {
		log.Println("starting to post listen http")
		_, err = serverListen.Write([]byte("GET /listen HTTP/1.1\r\n" + "Host: " + proxyDomain + "\r\n" + "Accept: */*\r\n" + "Upgrade: websocket\r\n" + "Connection: Upgrade\r\n" + "Clientid: " + clientId + "\r\n" + "Connection: keep-alive\r\n" + "Sec-WebSocket-Version: 13\r\n" + "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n" + "\r\n"))
		//	fmt.Fprintf(serverListen, "GET /listen HTTP/1.1\r\n")
		//	fmt.Fprintf(serverListen, "Host: "+proxyDomain+"\r\n")
		//	fmt.Fprintf(serverListen, "Accept: */*\r\n")
		//	fmt.Fprintf(serverListen, "Clientid: "+clientId+"\r\n")
		//	fmt.Fprintf(serverListen, "Connection: keep-alive\r\n")
		//	fmt.Fprintf(serverListen, "\r\n")

		log.Println("succed get listen")
		buf := bufio.NewReader(serverListen)

		success := false
		//	log.Println(buf)
		for line, err := buf.ReadString('\n'); true; line, err = buf.ReadString('\n') {
			if err != nil {
				log.Println("Failed to read following lines")
				return
			}

			if line == "HTTP/1.1 101 Switching Protocols\r\n" {
				success = true
				log.Println("succed get rsp")
			}

			if success && line == "\r\n" {
				break
			}
		}

		if success {
			log.Println("entering downlink sync process...")
			_, err = io.Copy(clientConn, buf)

			if err != nil {
				log.Println("Error copying server to client stream", err)
			}
		} else {
			log.Println("Failed to bind listen connection!")
		}

		wait <- true

	}()

	<-wait
}

func main() {
	log.Println("Listening...")
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Println("Error listening!", err)
		return
	}

	for true {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error accepting connection", err)
			continue
		}
		go handleConnection(conn)
	}

}
