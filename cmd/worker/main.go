package main

import (
	"log"
	"net/http"
	"net/rpc"

	"github.com/johananl/simple-cm/worker"
)

func main() {
	// Initialize RPC server
	w := worker.Worker{ModuleDir: "./modules"}
	rpc.Register(&w)
	rpc.HandleHTTP()

	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		log.Fatal(err)
	}
}
