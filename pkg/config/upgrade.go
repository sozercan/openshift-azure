package config

import (
	acsapi "github.com/openshift/openshift-azure/pkg/api"
)

const (
	versionLatest = 1
)

func Upgrade(cs *acsapi.ContainerService, c *Config) error {
	return nil
}
