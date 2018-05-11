package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
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
	Script() []string
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
func (o *FileExistsOperation) Script() []string {
	return []string{"[", "-f", o.Path, "]"}
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
func (o *FileContainsOperation) Script() []string {
	return []string{"grep", "-q", o.Text, o.Path}
}

// Execute executes an operation.
func (w *Worker) Execute(h Host, o Operation) (*OperationResult, error) {
	log.Printf("Executing operation %s", o.Desc())

	script := o.Script()
	sshCmd := append([]string{"-i", h.KeyPath, fmt.Sprintf("%s@%s", h.User, h.Hostname)}, script...)
	cmd := exec.Command("ssh", sshCmd...)

	var stdOut, stdErr bytes.Buffer

	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	err := cmd.Run()

	if err != nil {
		return nil, err
	}

	r := OperationResult{
		StdOut: string(stdOut.Bytes()),
		StdErr: string(stdErr.Bytes()),
	}

	return &r, nil
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
