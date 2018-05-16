package worker

import (
	"bytes"
	"fmt"
	"log"
	"time"

	ops "github.com/johananl/simple-cm/operations"
	"golang.org/x/crypto/ssh"
)

// A Worker executes operations.
type Worker struct {
	ModulesDir string
}

// ExecuteInput represents the input to the Execute function. It the hostname, username and private
// SSH key which should be used to connect to the remote host, as well as one or more operations to
// be executed on it.
type ExecuteInput struct {
	Hostname   string
	User       string
	Key        string
	Operations []ops.Operation
}

// ExecuteOutput represents the output returned by the Execute function. The output contains a
// slice of OperationResults.
type ExecuteOutput struct {
	Results []ops.OperationResult
}

// Execute executes one or more Operations on a remote host.
func (w *Worker) Execute(in *ExecuteInput, out *ExecuteOutput) error {
	// Initialize SSH connection to remote host
	k, err := parseKey([]byte(in.Key))
	if err != nil {
		return fmt.Errorf("could not parse SSH key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: in.User,
		Auth: []ssh.AuthMethod{k},
		// The following line prevents the need to manually approve each remote host as a known
		// host. This, however, poses a security risk and a better mechanism should probably be
		// used in production.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", in.Hostname), config)
	if err != nil {
		return fmt.Errorf("failed to dial: %v", err)
	}

	// Execute operations
	var results []ops.OperationResult

	for _, o := range in.Operations {
		stdOut, stdErr, err := w.executeOperation(client, in.Hostname, o)

		r := ops.OperationResult{Operation: o, StdOut: stdOut, StdErr: stdErr}
		if err != nil {
			log.Printf("Execution failed: %v", err)
			if stdOut != "" {
				log.Printf("stdout: %s", stdOut)
			}
			if stdErr != "" {
				log.Printf("stderr: %s", stdErr)
			}
		} else {
			r.Successful = true
		}
		results = append(results, r)
	}
	out.Results = results

	return nil
}

// Executes one Operation on a remote host. The function sends back OperationResults or an error.
func (w *Worker) executeOperation(c *ssh.Client, host string, o ops.Operation) (string, string, error) {
	log.Printf("[%s] Executing operation %s", host, o.Description)
	// Initialize session (this needs to be done per operation).
	sess, err := c.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("failed to create session: %v", err)
	}
	defer sess.Close()

	var stdOut, stdErr bytes.Buffer

	sess.Stdout = &stdOut
	sess.Stderr = &stdErr
	script, err := o.Script(w.ModulesDir)
	if err != nil {
		return "", "", err
	}

	log.Printf("Running the following script:\n%v", formatScriptOutput(script))
	err = sess.Run(script)

	stdOutStr := string(stdOut.Bytes())
	stdErrStr := string(stdErr.Bytes())

	return stdOutStr, stdErrStr, err
}

// Parses a private key and returns an ssh.AuthMethod.
func parseKey(b []byte) (ssh.AuthMethod, error) {
	key, err := ssh.ParsePrivateKey(b)
	if err != nil {
		return nil, fmt.Errorf("error parsing key: %v", err)
	}
	return ssh.PublicKeys(key), nil
}

// Formats a script's output for visual clarity.
func formatScriptOutput(s string) string {
	return "===================================================================\n" +
		s + "\n" +
		"===================================================================\n"
}
