module github.com/vulcanize/eth-statediff-service

go 1.13

require (
	github.com/ethereum/go-ethereum v1.9.24
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/vulcanize/go-eth-state-node-iterator v0.0.1-alpha
)

replace github.com/ethereum/go-ethereum v1.9.24 => github.com/vulcanize/go-ethereum v1.9.24-statediff-0.0.11
