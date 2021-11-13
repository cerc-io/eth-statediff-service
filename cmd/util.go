package cmd

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	gethsd "github.com/ethereum/go-ethereum/statediff"
	ind "github.com/ethereum/go-ethereum/statediff/indexer"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/spf13/viper"

	sd "github.com/vulcanize/eth-statediff-service/pkg"
)

type blockRange [2]uint64

func createStateDiffService() (sd.StateDiffService, error) {
	// load some necessary params
	logWithCommand.Info("Loading statediff service parameters")
	path := viper.GetString("leveldb.path")
	ancientPath := viper.GetString("leveldb.ancient")
	if path == "" || ancientPath == "" {
		logWithCommand.Fatal("require a valid eth leveldb primary datastore path and ancient datastore path")
	}

	nodeInfo := getEthNodeInfo()
	chainConf, err := chainConfig(nodeInfo.ChainID)
	if err != nil {
		logWithCommand.Fatal(err)
	}

	// create leveldb reader
	logWithCommand.Info("Creating leveldb reader")
	readerConf := sd.LvLDBReaderConfig{
		TrieConfig: &trie.Config{
			Cache:     viper.GetInt("cache.trie"),
			Journal:   "",
			Preimages: false,
		},
		ChainConfig: chainConf,
		Path:        path,
		AncientPath: ancientPath,
		DBCacheSize: viper.GetInt("cache.database"),
	}
	lvlDBReader, err := sd.NewLvlDBReader(readerConf)
	if err != nil {
		logWithCommand.Fatal(err)
	}

	// create statediff service
	logWithCommand.Info("Setting up database")
	conf, err := getConfig(nodeInfo)
	if err != nil {
		logWithCommand.Fatal(err)
	}
	logWithCommand.Info("Creating statediff indexer")
	indexer, err := ind.NewStateDiffIndexer(context.Background(), chainConf, nodeInfo, conf)
	if err != nil {
		logWithCommand.Fatal(err)
	}
	logWithCommand.Info("Creating statediff service")
	sdConf := sd.Config{
		ServiceWorkers:  viper.GetUint("statediff.serviceWorkers"),
		TrieWorkers:     viper.GetUint("statediff.trieWorkers"),
		WorkerQueueSize: viper.GetUint("statediff.workerQueueSize"),
		PreRuns:         setupPreRunRanges(),
	}
	return sd.NewStateDiffService(lvlDBReader, indexer, sdConf)
}

func setupPreRunRanges() []sd.RangeRequest {
	if !viper.GetBool("statediff.prerun") {
		return nil
	}
	preRunParams := gethsd.Params{
		IntermediateStateNodes:   viper.GetBool("prerun.params.intermediateStateNodes"),
		IntermediateStorageNodes: viper.GetBool("prerun.params.intermediateStorageNodes"),
		IncludeBlock:             viper.GetBool("prerun.params.includeBlock"),
		IncludeReceipts:          viper.GetBool("prerun.params.includeReceipts"),
		IncludeTD:                viper.GetBool("prerun.params.includeTD"),
		IncludeCode:              viper.GetBool("prerun.params.includeCode"),
	}
	var addrStrs []string
	viper.UnmarshalKey("prerun.params.watchedAddresses", &addrStrs)
	addrs := make([]common.Address, len(addrStrs))
	for i, addrStr := range addrStrs {
		addrs[i] = common.HexToAddress(addrStr)
	}
	preRunParams.WatchedAddresses = addrs
	var storageKeyStrs []string
	viper.UnmarshalKey("prerun.params.watchedStorageKeys", &storageKeyStrs)
	keys := make([]common.Hash, len(storageKeyStrs))
	for i, keyStr := range storageKeyStrs {
		keys[i] = common.HexToHash(keyStr)
	}
	preRunParams.WatchedStorageSlots = keys
	var rawRanges []blockRange
	viper.UnmarshalKey("prerun.ranges", &rawRanges)
	blockRanges := make([]sd.RangeRequest, len(rawRanges))
	for i, rawRange := range rawRanges {
		blockRanges[i] = sd.RangeRequest{
			Start:  rawRange[0],
			Stop:   rawRange[1],
			Params: preRunParams,
		}
	}
	if viper.IsSet("prerun.start") && viper.IsSet("prerun.stop") {
		hardStart := viper.GetInt("prerun.start")
		hardStop := viper.GetInt("prerun.stop")
		blockRanges = append(blockRanges, sd.RangeRequest{
			Start:  uint64(hardStart),
			Stop:   uint64(hardStop),
			Params: preRunParams,
		})
	}

	return blockRanges
}

func chainConfig(chainID uint64) (*params.ChainConfig, error) {
	switch chainID {
	case 1:
		return params.MainnetChainConfig, nil
	case 3:
		return params.RopstenChainConfig, nil // Ropsten
	case 4:
		return params.RinkebyChainConfig, nil
	case 5:
		return params.GoerliChainConfig, nil
	default:
		return nil, fmt.Errorf("chain config for chainid %d not available", chainID)
	}
}
