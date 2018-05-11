package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"

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
	return fmt.Sprintf("[ -f %s ]", o.Path)
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
	return fmt.Sprintf("grep -q %s %s", o.Text, o.Path)
}

// Execute executes an operation.
func (w *Worker) Execute(h Host, o Operation) (*OperationResult, error) {
	log.Printf("Executing operation %s on host %s", o.Desc(), h.Hostname)

	config := &ssh.ClientConfig{
		User: h.User,
		Auth: []ssh.AuthMethod{PublicKeyFile(h.KeyPath)},
		// The following line prevents the need to manually approve each remote host as a known
		// host. This, however, poses a security risk and a better mechanism should probably be
		// used in production.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", h.Hostname), config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()

	var stdOut, stdErr bytes.Buffer
	session.Stdout = &stdOut
	session.Stderr = &stdErr

	script := o.Script()

	err = session.Run(script)
	if err != nil {
		return nil, err
	}

	r := OperationResult{
		StdOut: string(stdOut.Bytes()),
		StdErr: string(stdErr.Bytes()),
	}

	return &r, nil
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

	res, err := w.Execute(h, &feo)
	if err != nil {
		log.Fatalf("Error while executing operation: %s", err)
	}

	log.Printf("Operation completed with exit code %d", res.ExitCode)

	res, err = w.Execute(h, &fco)
	if err != nil {
		log.Fatalf("Error while executing operation: %s", err)
	}

	log.Printf("Operation completed with exit code %d", res.ExitCode)
}
