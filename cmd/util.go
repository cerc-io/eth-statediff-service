package cmd

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/statediff"
	ind "github.com/ethereum/go-ethereum/statediff/indexer"
	"github.com/ethereum/go-ethereum/statediff/indexer/node"
	"github.com/ethereum/go-ethereum/statediff/indexer/shared"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/spf13/viper"

	sd "github.com/cerc-io/eth-statediff-service/pkg"
	"github.com/cerc-io/eth-statediff-service/pkg/prom"
)

type blockRange [2]uint64

func createStateDiffService(lvlDBReader sd.Reader, chainConf *params.ChainConfig, nodeInfo node.Info) (sd.StateDiffService, error) {
	// create statediff service
	logWithCommand.Info("Setting up database")
	conf, err := getConfig(nodeInfo)
	if err != nil {
		logWithCommand.Fatal(err)
	}

	logWithCommand.Info("Creating statediff indexer")
	db, indexer, err := ind.NewStateDiffIndexer(context.Background(), chainConf, nodeInfo, conf)
	if err != nil {
		logWithCommand.Fatal(err)
	}
	if conf.Type() == shared.POSTGRES && viper.GetBool("prom.dbStats") {
		prom.RegisterDBCollector(viper.GetString("database.name"), db)
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
	preRunParams := statediff.Params{
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

func instantiateLevelDBReader() (sd.Reader, *params.ChainConfig, node.Info) {
	// load some necessary params
	logWithCommand.Info("Loading statediff service parameters")
	mode := viper.GetString("leveldb.mode")
	path := viper.GetString("leveldb.path")
	ancientPath := viper.GetString("leveldb.ancient")
	url := viper.GetString("leveldb.url")

	if mode == "local" {
		if path == "" || ancientPath == "" {
			logWithCommand.Fatal("Require a valid eth LevelDB primary datastore path and ancient datastore path")
		}
	} else if mode == "remote" {
		if url == "" {
			logWithCommand.Fatal("Require a valid RPC url for accessing LevelDB")
		}
	} else {
		logWithCommand.Fatal("Invalid mode provided for LevelDB access")
	}

	nodeInfo := getEthNodeInfo()

	var chainConf *params.ChainConfig
	var err error
	chainConfigPath := viper.GetString("ethereum.chainConfig")

	if chainConfigPath != "" {
		chainConf, err = statediff.LoadConfig(chainConfigPath)
	} else {
		chainConf, err = statediff.ChainConfig(nodeInfo.ChainID)
	}

	if err != nil {
		logWithCommand.Fatalf("Unable to instantiate chain config: %s", err.Error())
	}

	// create LevelDB reader
	logWithCommand.Info("Creating LevelDB reader")
	readerConf := sd.LvLDBReaderConfig{
		TrieConfig: &trie.Config{
			Cache:     viper.GetInt("cache.trie"),
			Journal:   "",
			Preimages: false,
		},
		ChainConfig: chainConf,
		Mode:        mode,
		Path:        path,
		AncientPath: ancientPath,
		Url:         url,
		DBCacheSize: viper.GetInt("cache.database"),
	}
	reader, err := sd.NewLvlDBReader(readerConf)
	if err != nil {
		logWithCommand.Fatalf("Unable to instantiate levelDB reader: %s", err.Error())
	}
	return reader, chainConf, nodeInfo
}
