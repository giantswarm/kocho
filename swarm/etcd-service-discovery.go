package swarm

import (
	"io/ioutil"
	"net/http"

	"github.com/giantswarm/kocho/ssh"
	"github.com/giantswarm/kocho/swarm/types"

	"github.com/juju/errgo"
)

// RemoveInstanceFromDiscovery removes an instance from etcd discovery.
func RemoveInstanceFromDiscovery(i swarmtypes.Instance) error {
	etcdMemberName, err := ssh.GetEtcd2MemberName(i.PublicIPAddress)
	if err != nil {
		return errgo.Mask(err)
	}
	// Not a quorum member
	if etcdMemberName == "" {
		return nil
	}

	discoveryUrl, err := ssh.GetEtcdDiscoveryUrl(i.PublicIPAddress)
	if err != nil {
		return errgo.Mask(err)
	}

	machineUrl := discoveryUrl + "/" + etcdMemberName
	req, err := http.NewRequest("DELETE", machineUrl, nil)
	if err != nil {
		return errgo.Mask(err)
	}

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return errgo.Mask(err)
	}
	return nil
}

func getNewDiscoveryUrl() (string, error) {
	resp, err := http.Get(discoveryService)
	if err != nil {
		return "", errgo.Mask(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errgo.Mask(err)
	}

	return string(body), nil
}
