package master

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gocql/gocql"
)

// A Master coordinates Operations among Workers.
type Master struct {
	SSHKeysPath string
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
	s, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", m.SSHKeysPath, key))
	if err != nil {
		log.Fatalf("error reading SSH key: %v", err)
	}

	return string(s), nil
}

// // GetAllHosts gets all the hosts from the DB and returns a slice of Hosts.
// func (m *Master) GetAllHosts(sess *gocql.Session) []ops.Host {
// 	var hosts []ops.Host
// 	var hostname, user string
// 	q := `SELECT hostname, user FROM hosts`
// 	iter := sess.Query(q).Iter()
// 	for iter.Scan(&hostname, &user) {
// 		hosts = append(hosts, ops.Host{
// 			Hostname: hostname,
// 			User:     user,
// 			Key:      []byte(key),
// 		})
// 	}
// }
