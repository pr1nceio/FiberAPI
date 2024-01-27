package structs

type WispNodesResponse struct {
	Object string                 `json:"object"`
	Data   []WispObject           `json:"data"`
	Meta   map[string]interface{} `json:"meta"`
}

type WispPagination struct {
	Total       int `json:"total"`
	Count       int `json:"count"`
	PerPage     int `json:"per_page"`
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
}

type WispObject struct {
	Object     string   `json:"object"`
	Attributes WispNode `json:"attributes"`
}

type WispNode struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Connection  struct {
		FQDN    string      `json:"fqdn"`
		Display interface{} `json:"display"`
	} `json:"connection"`
	Ports  map[string]int `json:"ports"`
	Limits struct {
		CPU                int `json:"cpu"` // Percentage
		CPUOverallocate    int `json:"cpu_overallocate"`
		Memory             int `json:"memory"`
		MemoryOverallocate int `json:"memory_overallocate"`
		Disk               int `json:"disk"`
		DiskOverallocate   int `json:"disk_overallocate"`
	} `json:"limits"`
	Public          bool `json:"public"`
	MaintenanceMode bool `json:"maintenance_mode"`
	UploadSizeMB    int  `json:"upload_size"`
	ServersCount    int  `json:"servers_count"`
	MemoryUsageMB   int  `json:"memory_usage"`
	DiskUsageMB     int  `json:"disk_usage"`
	CPUUsage        int  `json:"cpu_usage"`
}
