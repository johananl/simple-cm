package main

import (
	"fmt"
	"log"

	"github.com/johananl/simple-cm/worker"
)

func main() {
	w := worker.Worker{}
	h := worker.Host{
		Hostname: "172.28.128.3",
		User:     "vagrant",
		KeyPath:  "./private_key",
	}

	feo := worker.FileExistsOperation{
		Description: "check_test_file_exists",
		Path:        "/tmp/test.txt",
	}

	fco := worker.FileContainsOperation{
		Description: "check_test_file_contains_hello",
		Path:        "/tmp/test.txt",
		Text:        "hello",
	}

	in := worker.ExecuteInput{
		Host:       h,
		Operations: []worker.Operation{feo, fco},
	}

	out := worker.ExecuteOutput{}

	err := w.Execute(&in, &out)
	if err != nil {
		log.Printf("Errors during execution")
	}

	log.Println("Completed operations:")
	for _, r := range out.Results {
		if r.Error == nil {
			fmt.Println(r.Operation.Desc())
		}
	}

	log.Println("Failed operations:")
	for _, r := range out.Results {
		if r.Error != nil {
			fmt.Println(r.Operation.Desc())
		}
	}
}
