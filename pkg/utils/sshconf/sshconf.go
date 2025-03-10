package sshconf

import (
	"slices"

	"github.com/mikkeloscar/sshconfig"
)

type SSHConfig struct {
	hosts []SSHEndpoint
}

type SSHEndpoint struct {
	host *sshconfig.SSHHost
}

func (e *SSHEndpoint) GetUser(defaultUser string) string {
	if e.host.User != "" {
		return e.host.User
	}

	return defaultUser
}

func NewSSHConfig(configFile string) *SSHConfig {
	hosts, _ := sshconfig.Parse(configFile)
	endpoints := make([]SSHEndpoint, len(hosts))
	for i, host := range hosts {
		endpoints[i] = SSHEndpoint{host: host}
	}
	return &SSHConfig{endpoints}
}

func (s *SSHConfig) GetHost(hostname string) *SSHEndpoint {
	for _, item := range s.hosts {
		if slices.Contains(item.host.Host, hostname) {
			return &item
		}
	}

	return nil
}
