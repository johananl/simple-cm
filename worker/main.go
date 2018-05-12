package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"golang.org/x/crypto/ssh"
)

// A Worker executes operations.
type Worker struct{}

// Host is a remote host against which Operations can be executed. The host should be reachable at
// Hostname over SSH using user User with SSH key at path KeyPath.
type Host struct {
	Hostname string
	User     string
	KeyPath  string
}

// Operation is an interface representing a generic operation.
type Operation interface {
	Desc() string
	Script() string
}

// OperationResult represents the result of an Operation.
type OperationResult struct {
	StdOut   string
	StdErr   string
	ExitCode int
}

// FileExistsOperation ensures the file at Path exists.
type FileExistsOperation struct {
	Description string
	Path        string
}

// Desc returns the operation's description.
func (o *FileExistsOperation) Desc() string {
	return o.Description
}

// Script returns the operation's script which can then be executed on remote hosts.
func (o *FileExistsOperation) Script() string {
	s := `#!/bin/bash

if [ ! -f %s ]; then
	touch %s
fi`
	return fmt.Sprintf(s, o.Path, o.Path)
}

// FileContainsOperation ensures the file at Path contains the text Text.
type FileContainsOperation struct {
	Description string
	Path        string
	Text        string
}

// Desc returns the operation's description.
func (o *FileContainsOperation) Desc() string {
	return o.Description
}

// Script returns the operation's script which can then be executed on remote hosts.
func (o *FileContainsOperation) Script() string {
	s := `#!/bin/bash

if ! grep -q %s %s; then
	echo "%s" >> %s
fi`
	return fmt.Sprintf(s, o.Text, o.Path, o.Text, o.Path)
}

// RunScript runs a script on a remote machine using the provided SSH Session and returns the
// stdout and stderr of the remote command as well as an error.
func (w *Worker) RunScript(sess *ssh.Session, script string) (*string, *string, error) {
	var stdOut, stdErr bytes.Buffer

	sess.Stdout = &stdOut
	sess.Stderr = &stdErr

	log.Printf(
		"Running the following script:\n"+
			"===================================================================\n"+
			"%s\n"+
			"===================================================================",
		script,
	)
	err := sess.Run(script)

	stdOutStr := string(stdOut.Bytes())
	stdErrStr := string(stdErr.Bytes())

	return &stdOutStr, &stdErrStr, err
}

// Execute executes one or more Operations on a remote host. The function sends back
// OperationResults and errors.
// TODO Break this down and make this testable.
func (w *Worker) Execute(c *ssh.Client, h Host, operations []Operation) (chan *OperationResult, chan error, chan bool) {
	log.Printf("Executing %d operation(s) on host %s", len(operations), h.Hostname)

	resChan := make(chan *OperationResult)
	errChan := make(chan error)
	done := make(chan bool)

	go func() {
		for _, o := range operations {
			log.Printf("[%s] Executing operation %s", h.Hostname, o.Desc())
			// Initialize session (this needs to be done per operation).
			// TODO Execute each operation in a separate function so that defer runs immediately
			// at the end of an operation.
			session, err := c.NewSession()
			if err != nil {
				errChan <- fmt.Errorf("failed to create session: %v", err)
			}
			defer session.Close()

			stdOut, stdErr, err := w.RunScript(session, o.Script())
			if err != nil {
				// Remote command returned an error
				// TODO Separate command errors from SSH errors
				errChan <- err
				continue
			}

			r := OperationResult{
				StdOut: *stdOut,
				StdErr: *stdErr,
			}

			resChan <- &r
		}

		done <- true
	}()

	return resChan, errChan, done
}

// PublicKeyFile reads a private SSH key from a file and returns an ssh.AuthMethod.
func PublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func main() {
	w := Worker{}
	h := Host{
		Hostname: "172.28.128.3",
		User:     "vagrant",
		KeyPath:  "./private_key",
	}

	feo := FileExistsOperation{
		Description: "check_test_file_exists",
		Path:        "/tmp/test.txt",
	}

	fco := FileContainsOperation{
		Description: "check_test_file_contains_hello",
		Path:        "/tmp/test.txt",
		Text:        "hello",
	}

	// Initialize SSH connection to remote host
	config := &ssh.ClientConfig{
		User: h.User,
		Auth: []ssh.AuthMethod{PublicKeyFile(h.KeyPath)},
		// The following line prevents the need to manually approve each remote host as a known
		// host. This, however, poses a security risk and a better mechanism should probably be
		// used in production.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", h.Hostname), config)
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}

	res, errChan, done := w.Execute(client, h, []Operation{&feo, &fco})
	for {
		select {
		case <-res:
		case e := <-errChan:
			log.Printf("Operation returned an error: %v", e)
		case <-done:
			log.Println("Done")
			return
		}
	}
}
