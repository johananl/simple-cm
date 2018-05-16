package main

import (
	"fmt"
	"log"
	"net/rpc"
	"sync"

	"github.com/johananl/simple-cm/master"
	ops "github.com/johananl/simple-cm/operations"
	"github.com/johananl/simple-cm/worker"
)

// Formats a script's output for visual clarity.
func formatScriptOutput(s string) string {
	return "===================================================================\n" +
		s + "\n" +
		"===================================================================\n"
}

func main() {
	// Init master
	m := master.Master{SSHKeysPath: "./ssh_keys"}

	// Connect to DB
	session, err := m.ConnectToDB([]string{"127.0.0.1"}, "simplecm")
	if err != nil {
		log.Fatalf("could not connect to DB: %v", err)
	}

	// Read hosts from DB
	hosts := m.GetAllHosts(session)
	log.Printf("%d hosts retrieved from DB", len(hosts))

	// Connect to workers
	// TODO Read wiring params from environment
	client, err := rpc.DialHTTP("tcp", "localhost:8888")
	if err != nil {
		log.Fatalf("error dialing: %v", err)
	}

	var wg sync.WaitGroup
	for _, h := range hosts {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Get operations for host
			operations := m.GetOperations(session, h.Hostname)

			key, err := m.SSHKey(h.KeyName)
			if err != nil {
				log.Printf("error reading SSH key for host %v: %v", h.Hostname, err)
				// TODO Handle failure indications for all operaions
				return
			}

			in := worker.ExecuteInput{
				Hostname:   h.Hostname,
				User:       h.User,
				Key:        key,
				Operations: operations,
			}
			var out worker.ExecuteOutput
			err = client.Call("Worker.Execute", in, &out)
			if err != nil {
				log.Printf("error executing operations: %v", err)
			}

			// Analyze results
			var good, bad []ops.OperationResult
			for _, i := range out.Results {
				if i.Successful {
					good = append(good, i)
				} else {
					bad = append(bad, i)
				}
			}

			// TODO Set colors for success / fail
			if len(good) > 0 {
				log.Println("Completed operations:")
				for _, i := range good {
					fmt.Println("* ", i.Operation.Description)
					if i.StdOut != "" {
						fmt.Printf("stdout:\n%v", formatScriptOutput(i.StdOut))
					}
					if i.StdErr != "" {
						fmt.Printf("stderr:\n%v", formatScriptOutput(i.StdErr))
					}
				}
			}

			if len(bad) > 0 {
				log.Println("Failed operations:")
				for _, i := range bad {
					fmt.Println("* ", i.Operation.Description)
					if i.StdOut != "" {
						fmt.Printf("stdout:\n%v", formatScriptOutput(i.StdOut))
					}
					if i.StdErr != "" {
						fmt.Printf("stderr:\n%v", formatScriptOutput(i.StdErr))
					}
				}
			}
		}()
		wg.Wait()
	}
}
