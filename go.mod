module github.com/vulcanize/eth-statediff-service

go 1.16

require (
	github.com/ethereum/go-ethereum v1.10.17
	github.com/jmoiron/sqlx v1.2.0
	github.com/prometheus/client_golang v1.4.0
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.3.0
	github.com/spf13/viper v1.10.1
	github.com/vulcanize/go-eth-state-node-iterator v1.0.0
	github.com/vulcanize/leveldb-ethdb-rpc v0.0.0-20220509104510-09fcf2aa603d
)

replace github.com/ethereum/go-ethereum v1.10.17 => github.com/vulcanize/go-ethereum v1.10.17-statediff-3.2.0
