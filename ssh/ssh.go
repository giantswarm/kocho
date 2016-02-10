package ssh

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/juju/errgo"
)

// Executor is responsible for executing the given command on the given host and returning its output.
type Executor interface {
	RunRemoteCommand(host, command string) (string, error)
}

// SSHShellExecutor provides an Executor using os.Exec to utilize the command line tool ssh.
type SSHShellExecutor struct {
	Binary   string
	Username string
}

var (
	// DefaultExecutor is an SSHShellExecutor with some default values.
	DefaultExecutor = &SSHShellExecutor{
		Username: "core",
		Binary:   "ssh",
	}
)

// RunRemoteCommand connects to the given host and runs the command, returning any output.
func (ssh *SSHShellExecutor) RunRemoteCommand(host, command string) (string, error) {
	cmd := exec.Command(ssh.Binary, "-o", "StrictHostKeyChecking=no", ssh.Username+"@"+host, command)
	var stdOut, stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	if err := cmd.Run(); err != nil {
		return "", errgo.NoteMask(err, stdErr.String())
	}

	out := stdOut.String()
	out = strings.TrimSuffix(out, "\n")
	return out, nil
}

// RunRemoteCommand executes the given command using the DefaultExecutor.
func RunRemoteCommand(host, command string) (string, error) {
	return DefaultExecutor.RunRemoteCommand(host, command)
}
