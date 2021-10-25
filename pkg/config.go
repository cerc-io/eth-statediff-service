package statediff

// Config holds config params for the statediffing service
type Config struct {
	ServiceWorkers  uint
	TrieWorkers     uint
	WorkerQueueSize uint
	PreRuns         []RangeRequest
}
