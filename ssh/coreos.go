package ssh

import (
	"github.com/juju/errgo"
)

// GetMachineID connects to the given host and returns the content of /etc/machine-id.
func GetMachineID(host string) (string, error) {
	id, err := RunRemoteCommand(host, "cat /etc/machine-id")
	return id, errgo.Mask(err)
}
