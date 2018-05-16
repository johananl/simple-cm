package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/rpc"

	"github.com/johananl/simple-cm/worker"
)

func main() {
	modulesDir := flag.String("modules-dir", "/etc/simple-cm/modules", "Directory to look for modules in")
	port := flag.String("port", "8888", "TCP port to listen on")
	flag.Parse()

	// Initialize RPC server
	w := worker.Worker{ModulesDir: *modulesDir}
	rpc.Register(&w)
	rpc.HandleHTTP()

	err := http.ListenAndServe(fmt.Sprintf(":%s", *port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
