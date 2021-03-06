package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"
	"strconv"
)

var (
	no_auth = []byte{0x05, 0x00}
	//	with_auth = []byte{0x05, 0x02}
	//	auth_success = []byte{0x05, 0x00}
	//	auth_failed  = []byte{0x05, 0x01}
	gen_failed      = []byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	connect_success = []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
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

	b := [1024]byte{0x0}

	_, err := connect.Read(b[:])
	if err != nil {
		return
	}

	if b[0] != 0x05 {
		connect.Close()
		return
	}
	log.Println("get socket5 conn reqest,sending rsp....")
	connect.Write(no_auth)

	n, err := connect.Read(b[:])
	var host string
	switch b[3] {
	case 0x01: //IP V4
		host = net.IPv4(b[4], b[5], b[6], b[7]).String()
	case 0x03: //domain
		host = string(b[5 : n-2]) //b[4] length of domain
	case 0x04: //IP V6
		host = net.IP{b[4], b[5], b[6], b[7], b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15], b[16], b[17], b[18], b[19]}.String()
	default:
		return
	}
	//	port := strconv.Itoa(int(b[n-2])<<8 | int(b[n-1]))
	var port uint16
	binary.Read(bytes.NewReader(b[(n-2):]), binary.BigEndian, &port)

	log.Println(host + ":")
	log.Println(port)
	//	log.Println(b[1])
	server, err := net.Dial("tcp", host+":"+strconv.Itoa(int(port)))
	if server != nil {
		defer server.Close()
	}
	if err != nil {
		log.Println("error:", err)
		connect.Write(gen_failed)
		return
	}
	if b[1] == 0x01 { //只支持connect
		connect.Write(connect_success)
		go io.Copy(server, connect)
		io.Copy(connect, server)

	} else {
		connect.Write(gen_failed)
	}
	return
}

func main() {

	socket, err := net.Listen("tcp", ":9979")
	if err != nil {
		return
	}
	log.Println("socks5 proxy server running on port 9979, listening ...\n")

	for {
		client, err := socket.Accept()
		if err != nil {
			return
		}

		var handler Handler = new(Socks5ProxyHandler)
		go handler.Handle(client)

		log.Println(client, " request handling...")
	}

}
