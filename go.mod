module github.com/vulcanize/eth-statediff-service

go 1.13

require (
	github.com/armon/consul-api v0.0.0-20180202201655-eb2c6b5be1b6 // indirect
	github.com/ethereum/go-ethereum v1.10.1
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	github.com/ugorji/go v1.1.4 // indirect
	github.com/vulcanize/go-eth-state-node-iterator v0.0.1-alpha
	github.com/xordataexchange/crypt v0.0.3-0.20170626215501-b2862e3d0a77 // indirect
)

replace github.com/ethereum/go-ethereum v1.10.1 => github.com/vulcanize/go-ethereum v1.10.1-statediff-0.0.16
