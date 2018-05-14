package worker

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/ssh"
)

// A Worker executes operations.
type Worker struct{}

// Host is a remote host against which Operations can be executed. The host should be reachable at
// Hostname over SSH using user User with private SSH key Key (Key contains the actual contents).
// TODO Store key contents here instead of path. The current situation creates a coupling between
// master and worker because of the file path.
type Host struct {
	Hostname string
	User     string
	Key      []byte
}

// Operation is an interface representing a generic operation.
type Operation interface {
	Desc() string
	Script() string
}

// OperationResult represents the result of an Operation.
type OperationResult struct {
	Operation Operation
	StdOut    string
	StdErr    string
	Error     error
}

// FileExistsOperation ensures the file at Path exists.
type FileExistsOperation struct {
	Description string
	Path        string
}

// Desc returns the operation's description.
func (o FileExistsOperation) Desc() string {
	return o.Description
}

// Script returns the operation's script which can then be executed on remote hosts.
func (o FileExistsOperation) Script() string {
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
func (o FileContainsOperation) Desc() string {
	return o.Description
}

// Script returns the operation's script which can then be executed on remote hosts.
func (o FileContainsOperation) Script() string {
	s := `#!/bin/bash

if ! grep -q %s %s; then
	echo "%s" >> %s
fi`
	return fmt.Sprintf(s, o.Text, o.Path, o.Text, o.Path)
}

// ExecuteInput represents the input to the Execute function. It should contain a Host and
// a slice of Operations.
type ExecuteInput struct {
	Host       Host
	Operations []Operation
}

// ExecuteOutput represents the output returned by the Execute function. The output contains a
// slice of OperationResults.
type ExecuteOutput struct {
	Results []OperationResult
}

// Execute executes one or more Operations on a remote host.
func (w *Worker) Execute(in *ExecuteInput, out *ExecuteOutput) error {
	// Initialize SSH connection to remote host
	k, err := publicKey(in.Host.Key)
	if err != nil {
		return fmt.Errorf("could not parse SSH key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: in.Host.User,
		Auth: []ssh.AuthMethod{k},
		// The following line prevents the need to manually approve each remote host as a known
		// host. This, however, poses a security risk and a better mechanism should probably be
		// used in production.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", in.Host.Hostname), config)
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}

	// Execute operations
	var failures error
	var results []OperationResult

	for _, o := range in.Operations {
		stdOut, stdErr, err := w.executeOperation(client, in.Host, o)
		if err != nil {
			log.Printf("Execution failed: %v", err)
			if *stdOut != "" {
				log.Printf("stdout: %s", *stdOut)
			}
			if *stdErr != "" {
				log.Printf("stderr: %s", *stdErr)
			}
			failures = errors.New("one or more operations failed")
		}

		r := OperationResult{Operation: o, StdOut: *stdOut, StdErr: *stdErr, Error: err}
		results = append(results, r)
	}
	out.Results = results

	return failures
}

// Executes one Operation on a remote host. The function sends back OperationResults or an error.
func (w *Worker) executeOperation(c *ssh.Client, h Host, o Operation) (*string, *string, error) {
	log.Printf("[%s] Executing operation %s", h.Hostname, o.Desc())
	// Initialize session (this needs to be done per operation).
	sess, err := c.NewSession()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %v", err)
	}
	defer sess.Close()

	// stdOut, stdErr, err := w.RunScript(session, o.Script())
	var stdOut, stdErr bytes.Buffer

	sess.Stdout = &stdOut
	sess.Stderr = &stdErr
	script := o.Script()

	log.Printf(
		"Running the following script:\n"+
			"===================================================================\n"+
			"%s\n"+
			"===================================================================",
		script,
	)
	err = sess.Run(script)

	stdOutStr := string(stdOut.Bytes())
	stdErrStr := string(stdErr.Bytes())

	return &stdOutStr, &stdErrStr, err
}

// Parses a private key and returns an ssh.AuthMethod.
func publicKey(b []byte) (ssh.AuthMethod, error) {
	key, err := ssh.ParsePrivateKey(b)
	if err != nil {
		return nil, fmt.Errorf("error parsing key: %v", err)
	}
	return ssh.PublicKeys(key), nil
}
