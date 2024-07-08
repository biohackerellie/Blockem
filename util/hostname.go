package util

import (
	"os"
	"strings"
)

var (
	hostname    string
	hostnameErr error
)

const hostnameFile string = "/etc/hostname"

func init() {
	getHostname(hostnameFile)
}

func Hostname() (string, error) {
	return hostname, hostnameErr
}

func HostnameString() string {
	return hostname
}

func getHostname(location string) {
	hostname, hostnameErr = os.Hostname()

	if hn, err := os.ReadFile(location); err == nil {
		hostname = strings.TrimSpace(string(hn))
		hostnameErr = nil

		return
	}
}
