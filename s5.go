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
	gen_failed      = []byte{0x05, 0x02, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	connect_success = []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
)

type Socks5ProxyHandler struct{}

type Handler interface {
	Handle(connect net.Conn)
}

func readLen(con net.Conn, len int) (buf []byte) {
	buf = make([]byte, len)

	n, _ := con.Read(buf)

	return buf[:n]
}

func (socks5 *Socks5ProxyHandler) Handle(connect net.Conn) {
	defer connect.Close()
	if connect == nil {
		return
	}

	var b []byte

	b = readLen(connect, 1)
	if b[0] != 0x05 {
		log.Println("only support socket5")
		_ = connect.Close()
		return
	}

	connect.Write(no_auth)

	b = readLen(connect, 4)

	cmd := b[1]
	switch cmd {
	case 0x01: //tcp
	case 0x02: //bind
		log.Println("不支持BIND")
		connect.Write(gen_failed)
		connect.Close()
		return
	case 0x03: //udp
		log.Println("不支持UDP")
		connect.Write(gen_failed)
		connect.Close()
		return
	}

	atyp := b[3]
	var host string
	var port uint16
	b = readLen(connect, 1024)
	switch atyp {
	case 0x01: //ipv4地址
		host = net.IP(b[:3]).String()
	case 0x03: //域名
		host = string(b[1 : len(b)-2])
	case 0x04: //ipv6地址
		host = net.IP(b[:15]).String()
	}
	_ = binary.Read(bytes.NewReader(b[len(b)-2:]), binary.BigEndian, &port)

	log.Println(host + ":" + string(port))
	//	log.Println(b[1])
	server, err := net.Dial("tcp", host+":"+strconv.Itoa(int(port)))
	if server != nil {
		defer server.Close()
	}
	if err != nil {
		log.Println("error:", err)
		return
	}

	_, _ = connect.Write([]byte{0x05, 0x00, 0x00, atyp})
	//把地址写回去
	_, _ = connect.Write(b)
	go io.Copy(server, connect)
	io.Copy(connect, server)

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
