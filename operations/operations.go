package operations

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
)

// Host is a remote host against which Operations can be executed. The host should be reachable at
// Hostname over SSH using user User with the private SSH key named Key.
type Host struct {
	Hostname string
	User     string
	// NOTE: Private SSH keys are transmitted from master to worker unencrypted over the network.
	// This is highly unsecure and should not be used as-is in production. Possible solutions:
	// - Encrypt the communication between master and worker.
	// - Store the keys in a secure, reference the key name from master and have worker pull it.
	KeyName string
}

type Operation struct {
	Description string
	ScriptName  string
	Attributes  map[string]string
}

// Script return the script which needs to be run in order to execute an Operation.
// TODO Improve handling of module dir path
func (o *Operation) Script(moduleDir string) (string, error) {
	log.Printf("Reading script at %s", o.ScriptName)
	// Template script with attributes
	tmpl, err := template.ParseFiles(fmt.Sprintf("%s/%s", moduleDir, o.ScriptName))
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
