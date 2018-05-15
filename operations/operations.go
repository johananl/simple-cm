package operations

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
)

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

type Operation struct {
	Description string
	Module      string
	Attributes  map[string]string
}

// Script return the script which needs to be run in order to execute an Operation.
func (o *Operation) Script() (string, error) {
	log.Printf("Reading script at %s", o.Module)
	// Template script with attributes
	tmpl, err := template.ParseFiles(o.Module)
	if err != nil {
		return "", fmt.Errorf("error parsing script template: %v", err)
	}

	b := bytes.Buffer{}
	err = tmpl.Execute(&b, o.Attributes)
	if err != nil {
		return "", fmt.Errorf("error templagint script: %v", err)
	}

	return string(b.Bytes()), nil
}

// OperationResult represents the result of an Operation.
type OperationResult struct {
	Operation  Operation
	StdOut     string
	StdErr     string
	Successful bool
}
