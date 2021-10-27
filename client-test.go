package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

const proxyDomain = "token.zhujq.ga:9999"
const port = "9990"

func handleConnection(clientConn net.Conn) {

	b := make([]byte, 1024)
	n, err := clientConn.Read(b)

	log.Println(b[:n])

	serverConn, err := net.Dial("tcp", proxyDomain)
	serverConn.Write(b[:n])

	clientConn.Write()
	buf := bytes.NewReader(b)
	//	log.Println(buf)
	var webclient http.Client
	rsp, err := webclient.Post(proxyDomain, "application/octet-stream", buf)
	if err != nil {
		log.Println("Error post:", err)
		return
	}
	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	log.Println(body)
	clientConn.Write(body)

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
