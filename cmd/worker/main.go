package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	server := http.Server{Addr: fmt.Sprintf(":%s", *port)}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()
	log.Printf("Listening for connections on :%s", *port)

	<-stop
	log.Println("Shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	log.Printf("Graceful shutdown complete")
}
