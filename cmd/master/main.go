package main

import (
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"net/rpc"

	"github.com/johananl/simple-cm/worker"
)

func main() {
	// Register types to allow gob serialization
	gob.Register(worker.FileExistsOperation{})
	gob.Register(worker.FileContainsOperation{})

	// TODO Read wiring params from environment
	client, err := rpc.DialHTTP("tcp", "localhost:8888")
	if err != nil {
		log.Fatalf("error dialing: %v", err)
	}

	// Read SSH private key
	buffer, err := ioutil.ReadFile("./private_key")
	if err != nil {
		log.Fatalf("error reading SSH key: %v", err)
	}

	h := worker.Host{
		Hostname: "172.28.128.3",
		User:     "vagrant",
		Key:      buffer,
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
		log.Printf("error executing operations: %v", err)
	}

	// Analyze results
	var good, bad []worker.OperationResult
	for _, i := range out.Results {
		if i.Successful {
			good = append(good, i)
		} else {
			bad = append(bad, i)
		}
	}

	if len(good) > 0 {
		log.Println("Completed operations:")
		for _, i := range good {
			fmt.Println(i.Operation.Desc())
			if i.StdOut != "" {
				fmt.Printf("stdout:\n"+
					"===================================================================\n"+
					"%s\n"+
					"===================================================================",
					i.StdOut,
				)
			}
			if i.StdErr != "" {
				fmt.Printf("stderr:\n"+
					"===================================================================\n"+
					"%s\n"+
					"===================================================================\n",
					i.StdErr,
				)
			}
		}
	}

	if len(bad) > 0 {
		log.Println("Failed operations:")
		for _, i := range bad {
			fmt.Println(i.Operation.Desc())
			if i.StdOut != "" {
				fmt.Printf("stdout:\n"+
					"===================================================================\n"+
					"%s\n"+
					"===================================================================\n",
					i.StdOut,
				)
			}
			if i.StdErr != "" {
				fmt.Printf("stderr:\n"+
					"===================================================================\n"+
					"%s\n"+
					"===================================================================\n",
					i.StdErr,
				)
			}
		}
	}
}
