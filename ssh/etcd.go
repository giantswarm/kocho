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

// GetEtcd2MemberName connects to the given host and returns the name of the given host in the etcd2 quorum.
// This assumes the member has "name=<machine-id>" set.
func GetEtcd2MemberName(host string) (string, error) {
	cmd := "etcdctl member list | fgrep \"name=$(cat /etc/machine-id)\" | cut -d: -f1"
	if name, err := RunRemoteCommand(host, cmd); err != nil {
		return "", errgo.Mask(err)
	} else {
		return name, nil
	}
}

// RemoveFromEtcd connects to the given host and removes it from the etcd discovery.
//
// Utilizes GetMachineID()
func RemoveFromEtcd(host string) error {
	// Dirty: Remove the machine from the etcd cluster by running a HTTP DELETE
	// request from the machine we want to kill
	cmd := []string{
		"etcdctl",
		"member",
		"remove",
		"$(etcdctl member list | fgrep \"name=$(cat /etc/machine-id)\" | cut -d: -f1)",
	}
	if _, err := RunRemoteCommand(host, strings.Join(cmd, " ")); err != nil {
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
