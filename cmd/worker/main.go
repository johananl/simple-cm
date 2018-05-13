package main

import (
	"log"
	"net"
	"net/rpc"

	"github.com/johananl/simple-cm/worker"
)

func main() {
	w := new(worker.Worker)
	rpc.Register(w)

	tcpAddr, err := net.ResolveTCPAddr("tcp", ":1234")
	if err != nil {
		log.Fatalf("error resolving TCP address: %v", err)
	}

	l, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatalf("error listening: %v", err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("error accepting: %v", err)
			continue
		}
		rpc.ServeConn(conn)
	}
}
