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

const (
	moduleDir = "modules"
)

// Formats a script's output for visual clarity.
func formatScriptOutput(s string) string {
	return "===================================================================\n" +
		s + "\n" +
		"===================================================================\n"
}

// Allows resolving a string denoting an operation's type to the correct Go type.
// var factory = map[string]interface{}{
// 	"FileExistsOperation":   ops.FileContainsOperation{},
// 	"FileContainsOperation": ops.FileContainsOperation{},
// }

func main() {
	// Register types to allow gob serialization
	// gob.Register(ops.FileExistsOperation{})
	// gob.Register(ops.FileContainsOperation{})

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
			// operations := []ops.Operation{
			// 	ops.FileExistsOperation{
			// 		Description: "verify_test_file_exists",
			// 		Path:        "/tmp/test.txt",
			// 	},
			// 	ops.FileContainsOperation{
			// 		Description: "verify_test_file_contains_hello",
			// 		Path:        "/tmp/test.txt",
			// 		Text:        "hello",
			// 	},
			// }
			operations := []ops.Operation{
				ops.Operation{
					Description: "verify_test_file_exists",
					Module:      fmt.Sprintf("%s/%s", moduleDir, "file_exists"),
					Attributes: map[string]string{
						"Path": "/tmp/test.txt",
					},
				},
				ops.Operation{
					Description: "verify_test_file_contains_hello",
					Module:      fmt.Sprintf("%s/%s", moduleDir, "file_contains"),
					Attributes: map[string]string{
						"Path": "/tmp/test.txt",
						"Text": "hello",
					},
				},
			}

			// // Get operations for host
			// var operations []ops.Operation
			// var hostname, opType string
			// var attributes map[string]string
			// q := `SELECT hostname, op_type, attributes FROM operations`
			// iter := session.Query(q).Iter()
			// for iter.Scan(&hostname, &opType, &attributes) {

			// }

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
