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
		s +
		"===================================================================\n"
}

func main() {
	sshKeysPath := flag.String("ssh-keys-dir", "/etc/simple-cm/keys", "Directory to look for SSH keys in")
	dbHostsFlag := flag.String("db-hosts", "127.0.0.1", "A comma-separated list of DB nodes to connect to")
	dbKeyspace := flag.String("db-keyspace", "simplecm", "Cassandra keyspace to use")
	workersFlag := flag.String("workers", "127.0.0.1:8888", "A comma-separated list of workers to connect to, in a <host>:<port> format")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

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
			continue
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
				log.Printf("[%s] Could not get operations from DB: %v", h.Hostname, err)
				return
			}

			log.Printf("[%s] Retrieved %d operations", h.Hostname, len(operations))

			// Read SSH key only if configured
			key := ""
			if h.KeyName != "" {
				key, err = m.SSHKey(h.KeyName)
				if err != nil {
					log.Printf("[%s] Error reading SSH key: %v", h.Hostname, err)
					// Not returning here because we might still be able to log in with a password.
				}
			}

			// Execute operations
			in := worker.ExecuteInput{
				Hostname:   h.Hostname,
				User:       h.User,
				Key:        key,
				Password:   h.Password,
				Operations: operations,
			}
			var out worker.ExecuteOutput

			client, err := m.SelectWorker()
			if err != nil {
				log.Printf("[%s] Could not select worker: %v", h.Hostname, err)
				return
			}

			// TODO Call RPC asynchronously?
			err = client.Call("Worker.Execute", in, &out)
			if err != nil {
				log.Printf("[%s] Error executing operations: %v", h.Hostname, err)
				return
			}

			// Store results in DB
			err = m.StoreResults(session, runID, h.Hostname, out.Results)
			if err != nil {
				log.Printf("[%s] Could not store results in DB: %v", h.Hostname, err)
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
				s := fmt.Sprintf("[%s] Completed operations:\n", h.Hostname)
				for _, i := range good {
					s = s + fmt.Sprintf("* %s\n", i.Operation.Description)
					if i.StdOut != "" {
						s = s + fmt.Sprintf("stdout:\n%v", formatScriptOutput(i.StdOut))
					}
					if i.StdErr != "" {
						s = s + fmt.Sprintf("stderr:\n%v", formatScriptOutput(i.StdErr))
					}
				}
				log.Print(s)
			}

			if len(bad) > 0 {
				s := fmt.Sprintf("[%s] Failed operations:\n", h.Hostname)
				for _, i := range bad {
					s = s + fmt.Sprintf("* %s\n", i.Operation.Description)
					if i.StdOut != "" {
						s = s + fmt.Sprintf("stdout:\n%v", formatScriptOutput(i.StdOut))
					}
					if i.StdErr != "" {
						s = s + fmt.Sprintf("stderr:\n%v", formatScriptOutput(i.StdErr))
					}
				}
				log.Print(s)
			}
		}()
		wg.Wait()
	}
}
