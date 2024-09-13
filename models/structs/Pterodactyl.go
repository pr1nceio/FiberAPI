package structs

import (
	gator "github.com/m41denx/alligator"
)

type PterodactylNodeExtended struct {
	*gator.Node
	ServersCount  int64 `json:"servers_count"`
	MemoryUsageMB int64 `json:"memory_usage"`
	DiskUsageMB   int64 `json:"disk_usage"`
	CPUUsage      int64 `json:"cpu_usage"`
}
