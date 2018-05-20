package master

import (
	"testing"
)

const (
	maxDBConnectionAttempts  = 2
	dbReconnectSleepInterval = 3
)

func TestConnectToDB(t *testing.T) {
	m := Master{}

	_, err := m.ConnectToDB([]string{"127.0.0.1"}, "simplecm")
	if err != nil {
		t.Fatalf("Error connecting to test DB: %v", err)
	}
}

// Starts a DB container to be used by the tests
// func setupDB() (string, error) {
// 	// Start test DB
// 	log.Printf("Starting DB container")
// 	cmd := exec.Command("docker", "run", "-d", "--rm", "-p", "9042:9042", "cassandra")
// 	time.Sleep(1 * time.Second)

// 	var stdOut, stdErr bytes.Buffer

// 	cmd.Stdout = &stdOut
// 	cmd.Stderr = &stdErr
// 	err := cmd.Run()
// 	if err != nil {
// 		return "", fmt.Errorf("error starting test DB: %v\nstdout: %s\nstderr: %s", err, string(stdOut.Bytes()), string(stdErr.Bytes()))
// 	}

// 	cID := string(stdOut.Bytes())
// 	log.Printf("Test DB running in container %s", cID)

// 	// Verify DB connectivity
// 	m := Master{}

// 	for i := 0; i < maxDBConnectionAttempts; i++ {
// 		_, err = m.ConnectToDB([]string{"127.0.0.1"}, "simplecm")
// 		if err != nil {
// 			log.Printf("DB connection attempt %d failed - retrying...", i+1)
// 			time.Sleep(dbReconnectSleepInterval * time.Second)
// 			continue
// 		}
// 		break
// 	}

// 	if err != nil {
// 		return cID, fmt.Errorf("error connecting to test DB: %v", err)
// 	}

// 	return cID, nil
// }

// Stops the DB container used by the tests
// func teardownDB(containerID string) error {
// 	// Stop test DB
// 	log.Printf("Stopping DB container %s", containerID)
// 	cmd := exec.Command("docker", "stop", containerID)

// 	var stdOut, stdErr bytes.Buffer
// 	cmd.Stdout = &stdOut
// 	cmd.Stderr = &stdErr
// 	err := cmd.Run()
// 	if err != nil {
// 		return fmt.Errorf("error stopping test DB: %v\nstdout: %s\nstderr: %s", err, string(stdOut.Bytes()), string(stdErr.Bytes()))
// 	}

// 	return nil
// }

// func TestMain(m *testing.M) {
// 	// Setup
// 	cID, err := setupDB()
// 	if err != nil {
// 		log.Printf("Error while setting up test DB: %v", err)
// 		if cID != "" {
// 			err = teardownDB(cID)
// 			if err != nil {
// 				log.Printf("Could not stop test DB: %v", err)
// 			}
// 		}
// 		os.Exit(1)
// 	}

// 	// Run tests
// 	retCode := m.Run()

// 	// Teardown
// 	err = teardownDB(cID)
// 	if err != nil {
// 		log.Printf("Could not stop test DB: %v", err)
// 	}

// 	os.Exit(retCode)
// }
