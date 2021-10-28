package main

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
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
	r := bufio.NewReader(clientConn)
	b := [1024]byte{0x0}
	clientId := randSeq(20)

	for {

		n, err := r.Read(b[:])

		if err == io.EOF {
			log.Println(clientId+":clientConn has closed:", err)
			return
		}
		if err != nil {
			log.Println(clientId+":Read clientConn error:", err)
			return
		}
		//	if n == 0 {
		//		continue
		//	}

		log.Println(clientId + ":get req from s5client to webc,len is: " + strconv.Itoa(n))
		//	log.Println(b[:n])

		var webclient http.Client
		req, err := http.NewRequest("GET", "https://"+proxyDomain, bytes.NewReader(b[:n]))
		if err != nil {
			log.Println("Error post:", err)
			break
		}
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Host", proxyDomain)
		req.Header.Set("Content-Type", "multipart/form-data")
		req.Header.Add("Clientid", clientId)
		rsp, err := webclient.Do(req)
		if err != nil {
			log.Println(clientId+":Error post:", err)
			break
		}
		log.Println(clientId + ":Send req from webc to webs,len is: " + strconv.Itoa(n))
		defer rsp.Body.Close()
		/*	if rsp.StatusCode != http.StatusOK {
				log.Println("Get post err result:" + rsp.Status)
				break
			}
		*/
		body, err := ioutil.ReadAll(rsp.Body)
		log.Println(clientId + ":Geted rsp from webs to webc,len is " + strconv.Itoa(len(body)))
		//	log.Println(body)
		n, err = clientConn.Write(body)
		if err != nil {
			log.Println(clientId+":Write from webc to s5client error:", err)
			break
		}
		log.Println(clientId + ":Send from webc to s5client,len is " + strconv.Itoa(n))
		//	log.Println(body)

		continue
	}

	return

}

func main() {
	log.Println("Listening on " + port + "....")
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
