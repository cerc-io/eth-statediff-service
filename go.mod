module github.com/vulcanize/eth-statediff-service

go 1.13

require (
	github.com/ethereum/go-ethereum v1.10.9
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	github.com/vulcanize/go-eth-state-node-iterator v0.0.1-alpha
)

replace github.com/ethereum/go-ethereum v1.10.9 => github.com/vulcanize/go-ethereum v1.10.9-statediff-0.0.27

replace github.com/vulcanize/go-eth-state-node-iterator v0.0.1-alpha => ../../vulcanize/go-eth-state-node-iterator
