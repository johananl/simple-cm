package main

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"strings"
	"sync"

	"github.com/gocql/gocql"

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
	sshKeysPath := flag.String("ssh-keys-dir", "/etc/simple-cm/keys", "Directory to look for SSH keys in")
	dbHostsFlag := flag.String("db-hosts", "127.0.0.1", "A comma-separated list of DB nodes to connect to")
	dbKeyspace := flag.String("db-keyspace", "simplecm", "Cassandra keyspace to use")
	workersFlag := flag.String("workers", "127.0.0.1:8888", "A comma-separated list of workers to connect to, in a <host>:<port> format")
	flag.Parse()

	// Init master
	m := master.Master{SSHKeysDir: *sshKeysPath}

	// Connect to DB
	dbHosts := strings.Split(*dbHostsFlag, ",")
	log.Printf("Connecting to DB hosts %s", dbHosts)
	session, err := m.ConnectToDB(dbHosts, *dbKeyspace)
	if err != nil {
		log.Fatalf("Could not connect to DB: %v", err)
	}
	defer session.Close()

	// Read hosts from DB
	hosts, err := m.GetHosts(session)
	if err != nil {
		log.Fatalf("Could not get hosts from DB: %v", err)
	}

	log.Printf("%d hosts retrieved from DB", len(hosts))

	// Connect to workers
	workers := strings.Split(*workersFlag, ",")
	log.Printf("Connecting to workers %s", workers)
	for _, w := range workers {
		c, err := rpc.DialHTTP("tcp", w)
		if err != nil {
			log.Printf("Error dialing worker %v: %v", w, err)
		}
		m.Workers = append(m.Workers, c)
	}

	// Store new run in DB
	runID := gocql.TimeUUID()
	err = m.StoreRun(session, runID)
	if err != nil {
		log.Fatalf("Could not store run in DB: %v", err)
	}

	var wg sync.WaitGroup
	for _, h := range hosts {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Get operations for host
			operations, err := m.GetOperations(session, h.Hostname)
			if err != nil {
				log.Fatalf("Could not get operations for host %s from DB: %v", h.Hostname, err)
			}

			log.Printf("Retrieved %d operations for host %s", len(operations), h.Hostname)

			// Read SSH key only if configured
			var key string
			if h.KeyName != "" {
				key, err = m.SSHKey(h.KeyName)
				if err != nil {
					log.Printf("Error reading SSH key for host %v: %v", h.Hostname, err)
					// TODO Handle failure indications for all operaions
				}
			} else {
				key = ""
			}

			in := worker.ExecuteInput{
				Hostname:   h.Hostname,
				User:       h.User,
				Key:        key,
				Password:   h.Password,
				Operations: operations,
			}
			var out worker.ExecuteOutput

			client := m.SelectWorker()

			err = client.Call("Worker.Execute", in, &out)
			if err != nil {
				log.Printf("Error executing operations: %v", err)
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
