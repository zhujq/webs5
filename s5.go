package main

import (
	"log"
	"net"

	"io"
	"strconv"
	//	"./proxy"
)

var (
	no_auth = []byte{0x05, 0x00}
	//	with_auth = []byte{0x05, 0x02}

	//	auth_success = []byte{0x05, 0x00}
	//	auth_failed  = []byte{0x05, 0x01}

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

	b := make([]byte, 1024)

	n, err := connect.Read(b)
	if err != nil {
		return
	}

	if b[0] == 0x05 {

		connect.Write(no_auth)

		n, err = connect.Read(b)
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
		port := strconv.Itoa(int(b[n-2])<<8 | int(b[n-1]))
		log.Println(host)
		server, err := net.Dial("tcp", net.JoinHostPort(host, port))
		if server != nil {
			defer server.Close()
		}
		if err != nil {
			return
		}
		connect.Write(connect_success)

		go io.Copy(server, connect)
		io.Copy(connect, server)
	}
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
