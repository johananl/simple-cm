package operations

import "fmt"

// Host is a remote host against which Operations can be executed. The host should be reachable at
// Hostname over SSH using user User with private SSH key Key (Key contains the actual contents).
type Host struct {
	Hostname string
	User     string
	// NOTE: Private SSH keys are transmitted from master to worker unencrypted over the network.
	// This is highly unsecure and should not be used as-is in production. Possible solutions:
	// - Encrypt the communication between master and worker.
	// - Store the keys in a secure, reference the key name from master and have worker pull it.
	Key []byte
}

// Operation is an interface representing a generic operation.
type Operation interface {
	Desc() string
	Script() string
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

// OperationResult represents the result of an Operation.
type OperationResult struct {
	Operation  Operation
	StdOut     string
	StdErr     string
	Successful bool
}
