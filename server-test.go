package main

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
)

const port = ":8080"
const target = "127.0.0.1:9979"

var conns map[string]net.Conn = make(map[string]net.Conn)
var lock sync.Mutex

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("starting process....")
	s, _ := ioutil.ReadAll(r.Body)
	clientid := r.Header.Get("Clientid")
	log.Println(clientid)

	if clientid == "" {
		log.Println("Not Get cliendId")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if string(s) == "" {
		log.Println("Get empty msg,exit....")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	_, exists := conns[clientid]
	if exists == false {
		//	delete(conns, clientid)
		serverConn, err := net.Dial("tcp", target)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		lock.Lock()
		conns[clientid] = serverConn
		lock.Unlock()
	}

	log.Println(clientid + ":Get req from webc to webs,len is " + strconv.Itoa(len(s)))

	n, err := conns[clientid].Write(s)
	if err != nil {
		log.Println(clientid+":Write to server conn error:", err)
		w.WriteHeader(http.StatusNotFound)
		delete(conns, clientid)
		//	w.Write("")
		return
	}
	log.Println(clientid + ":send req from webs to s5,len is " + strconv.Itoa(n))

	b := [65536]byte{0x0} //64k的缓冲区

	n, err = conns[clientid].Read(b[:]) //这里无法进行for循环，否则导致程序一直在for循环中，因为reader的流式没有io.EOF，
	//所以socket5 over http/ssl 无法实现，不能用中间无连接的http协议来承载有连接的socket5协议，最多实现小包的传递
	if err != nil {
		log.Println(clientid+":Geted rsp from s5 to webs conn error:", err)
		delete(conns, clientid)
		w.WriteHeader(http.StatusNotFound)
		return

	}
	//	buf := bufio.NewReader(serverConn)
	//	log.Println(b[:n])
	log.Println(clientid + ":Geted rsp from s5 to webs,len is: " + strconv.Itoa(n))
	w.Header().Set("Content-Type", "application/octet-stream")
	//	w.Header().Set("Content-Length", "999999")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Connection", "keep-alive")
	//	w.WriteHeader(http.StatusOK)
	n, err = w.Write(b[:n])
	if err != nil {
		log.Println("Write to webs->webc conn error:", err)
		delete(conns, clientid)
		return
	}
	log.Println(clientid + ":send rsp from webs to webc,len is: " + strconv.Itoa(n))

}

func main() {
	log.Println("Listenging on" + port)
	http.HandleFunc("/", IndexHandler)
	http.ListenAndServe(port, nil)

}
