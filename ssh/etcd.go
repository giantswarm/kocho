package ssh

import (
	"strings"

	"github.com/juju/errgo"
)

// GetEtcdDiscoveryUrl connects to the given host and extracts the used etcd
// discovery URL.
func GetEtcdDiscoveryUrl(host string) (string, error) {
	uuid, err := RunRemoteCommand(host, "systemctl cat etcd2 | grep ETCD_DISCOVERY |grep -oE 'http.*/[a-z0-9A-Z]+'")
	return uuid, errgo.Mask(err)
}

// RemoveFromEtcd connects to the given host and removes it from the etcd discovery.
//
// Utilizes GetMachineID()
func RemoveFromEtcd(host string) error {
	machineId, err := GetMachineID(host)
	if err != nil {
		return errgo.Mask(err)
	}

	// Dirty: Remove the machine from the etcd cluster by running a HTTP DELETE
	// request from the machine we want to kill
	cmd := []string{
		"curl",
		"--silent",
		"--location",
		"--request DELETE",
		"http://127.0.0.1:2380/v2/admin/machines/" + machineId,
	}
	if _, err = RunRemoteCommand(host, strings.Join(cmd, " ")); err != nil {
		return errgo.Mask(err)
	}
	return nil
}

// StopEtcd connects to the given host and stops the etcd2 daemon running.
func StopEtcd(host string) error {
	cmd := []string{
		"sudo",
		"systemctl",
		"stop",
		"etcd",
		"&&",
		"sudo",
		"systemctl",
		"disable",
		"etcd",
	}

	if _, err := RunRemoteCommand(host, strings.Join(cmd, " ")); err != nil {
		return errgo.Mask(err)
	}

	return nil
}
