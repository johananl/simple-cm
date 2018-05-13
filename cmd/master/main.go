package main

import (
	"log"
	"net/rpc"

	"github.com/johananl/simple-cm/worker"
)

func main() {
	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		log.Fatalf("error dialing: %v", err)
	}

	h := worker.Host{
		Hostname: "172.28.128.3",
		User:     "vagrant",
		KeyPath:  "~/tmp/private_key",
	}
	o := []worker.Operation{
		worker.FileExistsOperation{
			Description: "verify_test_file_exists",
			Path:        "/tmp/test.txt",
		},
		worker.FileContainsOperation{
			Description: "verify_test_file_contains_hello",
			Path:        "/tmp/test.txt",
			Text:        "hello",
		},
	}

	in := worker.ExecuteInput{
		Host:       h,
		Operations: o,
	}
	var out worker.ExecuteOutput
	err = client.Call("Worker.Execute", in, &out)
	if err != nil {
		log.Fatalf("error executing operations: %v", err)
	}

	// Analyze results
	var good, bad []worker.OperationResult
	for _, i := range out.Results {
		if i.Error != nil {
			bad = append(bad, i)
		} else {
			good = append(good, i)
		}
	}

	if len(good) > 0 {
		log.Println("Completed operations:")
		for _, i := range good {
			log.Println(i.Operation)
		}
	}

	if len(bad) > 0 {
		log.Println("Failed operations:")
		for _, i := range bad {
			log.Println(i.Operation)
		}
	}
}
