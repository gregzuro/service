package common

import (
	"fmt"
	"io/ioutil"
	"net"
	"regexp"
	"runtime"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// GetEntityID constructs and returns an identifier based on the switch's environment
func (args *CommandArgs) GetEntityID() (string, error) {
	// TODO: if we do anything else with proc, consider replacing this with
	// a call to https://github.com/shirou/gopsutil
	// or https://github.com/mitchellh/go-ps
	if runtime.GOOS != "linux" {
		return args.ShortName, nil
	}
	cgroup, err := ioutil.ReadFile("/proc/1/cgroup")
	if err != nil {
		log.WithFields(log.Fields{
			"context": "ReadFile",
		}).Info(err)
		return args.ShortName, nil
	}
	r := regexp.MustCompile(`(?m)^[0-9]+:cpu(,cpuacct)?:(.*)`)
	m := r.FindSubmatch(cgroup)
	if m == nil {
		err = fmt.Errorf("No matching cgroup line. Change the regexp?")
		log.WithFields(log.Fields{
			"context": "FindSubmatch",
		}).Error(err)
		return args.ShortName, err
	}

	cpuCgroup := strings.Split(string(m[2]), "/")
	var entityID string
	switch cpuCgroup[1] {
	case "docker":
		// if so, use the container id
		entityID = args.ShortName + ":" + cpuCgroup[2][0:12]

	case "user":

		entityID = args.ShortName + ":"

		// otherwise use a MAC address
		ifs, err := net.Interfaces()
		if err != nil {
			log.WithFields(log.Fields{
				"context": "net.Interface()"}).Error(err)
			return entityID, err
		}
		for _, v := range ifs {
			h := v.HardwareAddr.String()
			if len(h) != 0 {
				if v.Name == "eth0" {
					entityID = entityID + h
					break
				}
			}
		}

	default:
		entityID = entityID + ":" + "dunno"
	}

	// are we running on a phone / device?

	// are we running in a regular os environment on a server?

	return entityID, nil
}
