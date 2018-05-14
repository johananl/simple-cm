package main

import (
	"encoding/gob"
	"log"
	"net/http"
	"net/rpc"

	"golang.org/x/crypto/ssh"

	"github.com/johananl/simple-cm/worker"
)

func main() {
	// Register types to allow gob serialization
	gob.Register(worker.FileExistsOperation{})
	gob.Register(worker.FileContainsOperation{})
	gob.Register(ssh.ExitError{})

	// Initialize RPC server
	w := new(worker.Worker)
	rpc.Register(w)
	rpc.HandleHTTP()

	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		log.Fatal(err)
	}
}
