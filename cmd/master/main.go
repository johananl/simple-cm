package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sync"

	ops "github.com/johananl/simple-cm/operations"
	"github.com/johananl/simple-cm/worker"

	"github.com/gocql/gocql"
)

// Formats a script's output for visual clarity.
func formatScriptOutput(s string) string {
	return "===================================================================\n" +
		s + "\n" +
		"===================================================================\n"
}

func main() {
	// Read SSH private key
	key, err := ioutil.ReadFile(os.Getenv("SSH_KEY"))
	if err != nil {
		log.Fatalf("error reading SSH key: %v", err)
	}

	// Connect to DB
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "simplecm"
	session, _ := cluster.CreateSession()
	defer session.Close()

	// Get all hosts
	var hosts []ops.Host
	var hostname, user string
	q := `SELECT hostname, user FROM hosts`
	iter := session.Query(q).Iter()
	for iter.Scan(&hostname, &user) {
		hosts = append(hosts, ops.Host{
			Hostname: hostname,
			User:     user,
			Key:      []byte(key),
		})
	}

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
			var operations []ops.Operation
			var description, scriptName string
			var attributes map[string]string
			q := `SELECT description, script_name, attributes FROM operations where hostname = ?`
			iter := session.Query(q, h.Hostname).Iter()
			for iter.Scan(&description, &scriptName, &attributes) {
				o := ops.Operation{
					Description: description,
					ScriptName:  scriptName,
					Attributes:  attributes,
				}
				operations = append(operations, o)
			}

			in := worker.ExecuteInput{
				Host:       h,
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
