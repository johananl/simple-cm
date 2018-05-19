package master

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/rpc"
	"sync"
	"time"

	"github.com/gocql/gocql"
	ops "github.com/johananl/simple-cm/operations"
)

// A Master coordinates Operations among Workers.
type Master struct {
	SSHKeysDir     string
	Workers        []*rpc.Client
	LastUsedWorker int
	lock           sync.RWMutex
}

// ConnectToDB connects to the given DB and returns a *gocql.Session.
func (m *Master) ConnectToDB(hosts []string, keyspace string) (*gocql.Session, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = keyspace
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("error creating DB session: %v", err)
	}

	return session, nil
}

// SSHKey gets the name of an SSH private key and returns its contents.
func (m *Master) SSHKey(key string) (string, error) {
	s, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", m.SSHKeysDir, key))
	if err != nil {
		log.Fatalf("error reading SSH key: %v", err)
	}

	return string(s), nil
}

// GetAllHosts gets all the hosts from the DB and returns a slice of Hosts.
func (m *Master) GetHosts(session *gocql.Session) ([]ops.Host, error) {
	var hosts []ops.Host
	var hostname, user, keyName, password string
	q := `SELECT hostname, user, key_name, password FROM hosts`
	iter := session.Query(q).Iter()
	for iter.Scan(&hostname, &user, &keyName, &password) {
		hosts = append(hosts, ops.Host{
			Hostname: hostname,
			User:     user,
			KeyName:  keyName,
			Password: password,
		})
	}
	if err := iter.Close(); err != nil {
		return []ops.Host{}, fmt.Errorf("error getting hosts from DB: %v", err)
	}

	return hosts, nil
}

// GetOperations gets all operations for the given host from the DB and returns them in a slice.
func (m *Master) GetOperations(session *gocql.Session, hostname string) ([]ops.Operation, error) {
	var operations []ops.Operation
	var description, scriptName string
	var attributes map[string]string
	q := `SELECT description, script_name, attributes FROM operations where hostname = ?`
	iter := session.Query(q, hostname).Iter()
	for iter.Scan(&description, &scriptName, &attributes) {
		o := ops.Operation{
			Description: description,
			ScriptName:  scriptName,
			Attributes:  attributes,
		}
		operations = append(operations, o)
	}
	if err := iter.Close(); err != nil {
		return []ops.Operation{}, fmt.Errorf("error getting operations from DB: %v", err)
	}

	return operations, nil
}

// SelectWorker returns workers using a simple round-robin algorithm.
//
// NOTE: More sophisticated algorithms could of course be used to select workers. Round-robin is a
// very simple one. A "policy" argument could be added to this method where the caller could
// specify which policy they want to use for distributing the load across multiple workers.
func (m *Master) SelectWorker() (*rpc.Client, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if len(m.Workers) == 0 {
		return nil, errors.New("no workers connected")
	}

	if m.LastUsedWorker == len(m.Workers)-1 {
		// Last used worker is the last one in the slice - start over from index 0
		log.Println("Selected worker 0")
		m.LastUsedWorker = 0
		return m.Workers[0], nil
	}

	selected := m.LastUsedWorker + 1
	log.Printf("Selected worker %d", selected)

	m.LastUsedWorker++

	return m.Workers[selected], nil
}

// StoreRun stores a new run in the DB.
func (m *Master) StoreRun(session *gocql.Session, id gocql.UUID) error {
	log.Printf("Saving new run %s to DB", id.String())
	q := `INSERT INTO runs (id, create_time) values (?, ?)`
	if err := session.Query(q, id, time.Now()).Exec(); err != nil {
		return fmt.Errorf("error storing run in DB: %v", err)
	}
	return nil
}

// StoreResults stores the results of a run in the DB.
// TODO Store stdout and stderr in DB.
func (m *Master) StoreResults(session *gocql.Session, runID gocql.UUID, hostname string, results []ops.OperationResult) error {
	log.Printf("Saving %d results for host %s to DB", len(results), hostname)
	for _, r := range results {
		// Insert result atomically to two tables
		b := session.NewBatch(gocql.UnloggedBatch)

		now := time.Now()

		q1 := `INSERT INTO results_by_run_id (id, run_id, hostname, ts, script_name, successful)
			values (uuid(), ?, ?, ?, ?, ?)`
		b.Query(q1, runID, hostname, now, r.Operation.ScriptName, r.Successful)

		q2 := `INSERT INTO results_by_run_id_and_hostname
			(id, run_id, hostname, ts, script_name, successful)
			values (uuid(), ?, ?, ?, ?, ?)`
		b.Query(q2, runID, hostname, now, r.Operation.ScriptName, r.Successful)

		if err := session.ExecuteBatch(b); err != nil {
			return fmt.Errorf("error storing results in DB: %v", err)
		}
	}
	return nil
}
