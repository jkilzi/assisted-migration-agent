package models

import (
	"time"

	api "github.com/kubev2v/migration-planner/api/v1alpha1"
)

type InfrastructureData struct {
	Datastores            []api.Datastore
	Networks              []api.Network
	HostPowerStates       map[string]int
	Hosts                 *[]api.Host
	HostsPerCluster       []int
	ClustersPerDatacenter []int
	TotalHosts            int
	TotalClusters         int
	TotalDatacenters      int
	VmsPerCluster         []int
}

// Inventory represents inventory data stored in the database.
type Inventory struct {
	Data      []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}
