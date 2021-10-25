package main

import (
	"log"
	"net"

	"./proxy"
)

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

		var handler proxy.Handler = new(proxy.Socks5ProxyHandler)

		go handler.Handle(client)

		log.Println(client, " request handling...")
	}

}
