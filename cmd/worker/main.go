package main

import (
	"log"
	"net/http"
	"net/rpc"

	"github.com/johananl/simple-cm/worker"
)

func main() {
	// Register types to allow gob serialization
	// gob.Register(ops.FileExistsOperation{})
	// gob.Register(ops.FileContainsOperation{})
	// gob.Register(ssh.ExitError{})

	// Initialize RPC server
	w := new(worker.Worker)
	rpc.Register(w)
	rpc.HandleHTTP()

	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		log.Fatal(err)
	}
}
