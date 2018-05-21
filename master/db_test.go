// +build integration

package master

import (
	"log"
	"reflect"
	"testing"
	"time"

	ops "github.com/johananl/simple-cm/operations"

	"github.com/gocql/gocql"
)

var m Master
var keyspace string
var dbHosts []string

func init() {
	m = Master{}
	keyspace = "simplecm"
	dbHosts = []string{"127.0.0.1"}
}

func TestConnectToDB(t *testing.T) {
	// m := Master{}

	_, err := m.ConnectToDB(dbHosts, keyspace)
	if err != nil {
		t.Fatalf("Error connecting to test DB: %v", err)
	}
}

func TestGetHosts(t *testing.T) {
	session, err := m.ConnectToDB(dbHosts, keyspace)
	if err != nil {
		t.Fatalf("Error connecting to test DB: %v", err)
	}

	// Insert dummy hosts to DB
	q := `create table hosts(hostname text, user text, key_name text, password text,
		primary key(hostname));`
	if err := session.Query(q).Exec(); err != nil {
		t.Fatalf("Error creating table: %v", err)
	}
	q = `insert into hosts (hostname, user, key_name, password)
		values ('testhost', 'testuser', '','testpass');`
	if err := session.Query(q).Exec(); err != nil {
		t.Fatalf("Error inserting dummy hosts: %v", err)
	}

	// Run test
	hosts, err := m.GetHosts(session)
	if err != nil {
		t.Fatalf("Error getting hosts: %v", err)
	}

	// Verify
	if len(hosts) != 1 {
		log.Fatalf("Wrong number of hosts returned: got %d want %d", len(hosts), 1)
	}

	if hosts[0].Hostname != "testhost" {
		t.Fatalf("Wrong hostname retrieved: got %s want %s", hosts[0].Hostname, "testhost")
	}
	if hosts[0].User != "testuser" {
		t.Fatalf("Wrong user retrieved: got %s want %s", hosts[0].User, "testuser")
	}
	if hosts[0].Password != "testpass" {
		t.Fatalf("Wrong password retrieved: got %s want %s", hosts[0].Password, "testpass")
	}
}

func TestGetOperations(t *testing.T) {
	session, err := m.ConnectToDB(dbHosts, keyspace)
	if err != nil {
		t.Fatalf("Error connecting to test DB: %v", err)
	}

	// Insert dummy operations to DB
	q := `create table operations(id UUID, hostname text, description text, script_name text,
		attributes map<text, text>, primary key(hostname, id));`
	if err := session.Query(q).Exec(); err != nil {
		t.Fatalf("Error creating table: %v", err)
	}
	q = `insert into operations (id, hostname, description, script_name, attributes)
		values (uuid(), 'host1', 'verify_test_file_exists', 'file_exists',
		{'path': '/etc/passwd'});`
	if err := session.Query(q).Exec(); err != nil {
		t.Fatalf("Error inserting dummy operations: %v", err)
	}

	// Run test
	ops, err := m.GetOperations(session, "host1")
	if err != nil {
		t.Fatalf("Error getting hosts: %v", err)
	}

	// Verify
	if len(ops) != 1 {
		log.Fatalf("Wrong number of operations returned: got %d want %d", len(ops), 1)
	}

	if ops[0].Description != "verify_test_file_exists" {
		t.Fatalf("Wrong description: got %s want %s", ops[0].Description,
			"verify_test_file_exists")
	}
	if ops[0].ScriptName != "file_exists" {
		t.Fatalf("Wrong script name: got %s want %s", ops[0].ScriptName, "file_exists")
	}
	eq := reflect.DeepEqual(ops[0].Attributes, map[string]string{"path": "/etc/passwd"})
	if !eq {
		t.Fatalf("Wrong attributes: got %v want %v", ops[0].Attributes,
			map[string]string{"path": "/etc/passwd"})
	}
}

func TestStoreRun(t *testing.T) {
	session, err := m.ConnectToDB(dbHosts, keyspace)
	if err != nil {
		t.Fatalf("Error connecting to test DB: %v", err)
	}

	// Create table
	q := `create table runs(id UUID, create_time timestamp, primary key(id, create_time));`
	if err := session.Query(q).Exec(); err != nil {
		t.Fatalf("Error creating table: %v", err)
	}

	// Run test
	id := gocql.TimeUUID()
	ts := time.Now()
	err = m.StoreRun(session, id, ts)
	if err != nil {
		t.Fatalf("Error storing run: %v", err)
	}

	// Verify
	var idOut gocql.UUID
	var createTime time.Time
	q = `select id, create_time from runs where id = ? LIMIT 1`
	if err := session.Query(q, id).Consistency(gocql.One).Scan(&idOut, &createTime); err != nil {
		log.Fatalf("Error getting run from DB: %v", err)
	}
	if idOut != id {
		log.Fatalf("Wrong ID retrieved: got %v want %v", idOut, id)
	}
	// TODO Fix timezone conversion problem. There is a mismatch between how the timestamp is
	// represented in the DB and in the code.
	// if createTime != ts {
	// 	log.Fatalf("Wrong creation time received: got %v want %v", createTime, ts)
	// }
}

func TestStoreResults(t *testing.T) {
	session, err := m.ConnectToDB(dbHosts, keyspace)
	if err != nil {
		t.Fatalf("Error connecting to test DB: %v", err)
	}

	// Create tables
	q := `create table results_by_run_id(id UUID, run_id UUID, hostname text, ts timestamp,
		script_name text, successful boolean, primary key(run_id, id));`
	if err := session.Query(q).Exec(); err != nil {
		t.Fatalf("Error creating table: %v", err)
	}

	q = `create table results_by_run_id_and_hostname(id UUID, run_id UUID, hostname text,
		ts timestamp, script_name text, successful boolean, primary key(run_id, hostname, id));`
	if err = session.Query(q).Exec(); err != nil {
		t.Fatalf("Error creating table: %v", err)
	}

	// Run test
	runID := gocql.TimeUUID()
	hostname := "testhost"
	results := []ops.OperationResult{
		ops.OperationResult{
			Operation: ops.Operation{
				Description: "test_op",
				ScriptName:  "fake",
				Attributes:  map[string]string{"fakekey": "fakevalue"},
			},
			StdOut:     "",
			StdErr:     "",
			Successful: true,
		},
	}
	err = m.StoreResults(session, runID, hostname, results)
	if err != nil {
		t.Fatalf("Error storing run: %v", err)
	}

	// Verify
	var id, runIDOut gocql.UUID
	var hostnameOut string
	var ts time.Time
	var scriptName string
	var successful bool
	q = `select id, run_id, hostname, ts, script_name, successful from results_by_run_id
		where run_id = ? LIMIT 1`
	if err := session.Query(q, runID).Scan(&id, &runIDOut, &hostnameOut, &ts, &scriptName,
		&successful); err != nil {
		log.Fatalf("Error getting run from DB: %v", err)
	}
	if runIDOut != runID {
		t.Fatalf("Wrong run ID: got %v want %v", runIDOut, runID)
	}
	if hostnameOut != hostname {
		t.Fatalf("Wrong hostname: got %s want %s", hostnameOut, hostname)
	}
	if scriptName != "fake" {
		t.Fatalf("Wrong script name: got %s want fake", scriptName)
	}
	if !successful {
		t.Fatalf("Result should have been successful but is not")
	}
}
