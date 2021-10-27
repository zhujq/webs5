package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

const port = "80"
const target = "127.0.0.1:9979"

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	s, _ := ioutil.ReadAll(r.Body)
	if string(s) == "" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, string(""))
		return
	}

	serverConn, err := net.Dial("tcp", target)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, string(""))
		return
	}
	log.Println(s)
	serverConn.Write([]byte(s))

	buf := bufio.NewReader(serverConn)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", "999999")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)
	var buff []byte
	for {
		_, err := buf.Read(buff[:])
		if err == io.EOF {
			break
		}
	}
	w.Write(buff)
	return
}

func main() {
	http.HandleFunc("/", IndexHandler)
	http.ListenAndServe("127.0.0.0:9990", nil)
}
