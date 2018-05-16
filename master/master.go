package master

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/rpc"

	"github.com/gocql/gocql"
	ops "github.com/johananl/simple-cm/operations"
)

// A Master coordinates Operations among Workers.
type Master struct {
	SSHKeysDir string
	Workers    []*rpc.Client
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
func (m *Master) GetAllHosts(session *gocql.Session) []ops.Host {
	var hosts []ops.Host
	var hostname, user, keyName string
	q := `SELECT hostname, user, key_name FROM hosts`
	iter := session.Query(q).Iter()
	for iter.Scan(&hostname, &user, &keyName) {
		hosts = append(hosts, ops.Host{
			Hostname: hostname,
			User:     user,
			KeyName:  keyName,
		})
	}

	return hosts
}

// GetOperations gets all operations for the given host from the DB and returns them in a slice.
func (m *Master) GetOperations(session *gocql.Session, hostname string) []ops.Operation {
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

	return operations
}

// SelectWorker returns the best worker to send Operations to at the moment.
func (m *Master) SelectWorker() *rpc.Client {
	// TODO Implement selection logic
	return m.Workers[0]
}
