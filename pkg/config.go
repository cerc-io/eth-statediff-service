package statediff

// ServiceConfig holds config params for the statediffing service
type ServiceConfig struct {
	ServiceWorkers  uint
	TrieWorkers     uint
	WorkerQueueSize uint
	PreRuns         []RangeRequest
}
