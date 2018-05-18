package master

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/rpc"
	"sync"

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
func (m *Master) GetAllHosts(session *gocql.Session) ([]ops.Host, error) {
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
	q := `SELECT description, script_name, attributes FROM operations_by_hostname where hostname = ?`
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
func (m *Master) SelectWorker() *rpc.Client {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.LastUsedWorker == len(m.Workers)-1 {
		// Last used worker is the last one in the slice - start over from index 0
		log.Println("Selected worker 0")
		m.LastUsedWorker = 0
		return m.Workers[0]
	}

	current := m.LastUsedWorker + 1
	log.Printf("Selected worker %d", current)

	m.LastUsedWorker++

	return m.Workers[current]
}
