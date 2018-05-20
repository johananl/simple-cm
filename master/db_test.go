package master

import (
	"log"
	"testing"
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
	q := `create table hosts(hostname text, user text, key_name text, password text, primary key(hostname));`
	if err := session.Query(q).Exec(); err != nil {
		t.Fatalf("Error creating hosts table: %v", err)
	}
	q = `insert into hosts (hostname, user, key_name, password) values ('testhost', 'testuser', '', 'testpass');`
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
