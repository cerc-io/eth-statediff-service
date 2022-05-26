module github.com/vulcanize/eth-statediff-service

go 1.16

require (
	github.com/btcsuite/btcd v0.22.1 // indirect
	github.com/ethereum/go-ethereum v1.10.18
	github.com/jmoiron/sqlx v1.2.0
	github.com/prometheus/client_golang v1.4.0
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.3.0
	github.com/spf13/viper v1.10.1
	github.com/vulcanize/go-eth-state-node-iterator v1.0.1
	github.com/vulcanize/leveldb-ethdb-rpc v0.1.1
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
)

replace github.com/ethereum/go-ethereum v1.10.18 => github.com/vulcanize/go-ethereum v1.10.18-statediff-3.2.1
