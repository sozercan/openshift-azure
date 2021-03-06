// Package api defines the external API for the plugin.
package api

import (
	"context"

	"github.com/openshift/openshift-azure/pkg/api/v1"
)

type Plugin interface {
	ValidateExternal(oc *v1.OpenShiftCluster) []error

	// ValidateInternal exists (a) to be able to place validation logic in a
	// single place in the event of multiple external API versions, and (b) to
	// be able to compare a new API manifest against a pre-existing API manifest
	// (for update, upgrade, etc.)

	// TODO: confirm with MSFT that they can pass in `old` at the time
	// ValidateInternal is called and that it makes sense to do this.
	ValidateInternal(new, old *ContainerService) []error

	GenerateConfig(cs *ContainerService, configBytes []byte) ([]byte, error)

	GenerateARM(cs *ContainerService, configBytes []byte) ([]byte, error)

	HealthCheck(ctx context.Context, cs *ContainerService, configBytes []byte) error
}
