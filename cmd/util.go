package cmd

import (
	"context"

	statediff "github.com/cerc-io/plugeth-statediff"
	"github.com/cerc-io/plugeth-statediff/indexer"
	"github.com/cerc-io/plugeth-statediff/indexer/node"
	"github.com/cerc-io/plugeth-statediff/indexer/shared"
	"github.com/cerc-io/plugeth-statediff/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/spf13/viper"

	pkg "github.com/cerc-io/eth-statediff-service/pkg"
	"github.com/cerc-io/eth-statediff-service/pkg/prom"
)

type blockRange [2]uint64

func createStateDiffService(lvlDBReader pkg.Reader, chainConf *params.ChainConfig, nodeInfo node.Info) (*pkg.Service, error) {
	// create statediff service
	logWithCommand.Debug("Setting up database")
	conf, err := getConfig(nodeInfo)
	if err != nil {
		logWithCommand.Fatal(err)
	}

	logWithCommand.Debug("Creating statediff indexer")
	db, indexer, err := indexer.NewStateDiffIndexer(context.Background(), chainConf, nodeInfo, conf, true)
	if err != nil {
		logWithCommand.Fatal(err)
	}
	if conf.Type() == shared.POSTGRES && viper.GetBool("prom.dbStats") {
		prom.RegisterDBCollector(viper.GetString("database.name"), db)
	}

	logWithCommand.Debug("Creating statediff service")
	sdConf := pkg.ServiceConfig{
		ServiceWorkers:  viper.GetUint("statediff.serviceWorkers"),
		TrieWorkers:     viper.GetUint("statediff.trieWorkers"),
		WorkerQueueSize: viper.GetUint("statediff.workerQueueSize"),
		PreRuns:         setupPreRunRanges(),
	}
	return pkg.NewStateDiffService(lvlDBReader, indexer, sdConf), nil
}

func setupPreRunRanges() []pkg.RangeRequest {
	if !viper.GetBool("statediff.prerun") {
		return nil
	}
	preRunParams := statediff.Params{
		IncludeBlock:    viper.GetBool("prerun.params.includeBlock"),
		IncludeReceipts: viper.GetBool("prerun.params.includeReceipts"),
		IncludeTD:       viper.GetBool("prerun.params.includeTD"),
		IncludeCode:     viper.GetBool("prerun.params.includeCode"),
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
	blockRanges := make([]pkg.RangeRequest, len(rawRanges))
	for i, rawRange := range rawRanges {
		blockRanges[i] = pkg.RangeRequest{
			Start:  rawRange[0],
			Stop:   rawRange[1],
			Params: preRunParams,
		}
	}
	if viper.IsSet("prerun.start") && viper.IsSet("prerun.stop") {
		hardStart := viper.GetInt("prerun.start")
		hardStop := viper.GetInt("prerun.stop")
		blockRanges = append(blockRanges, pkg.RangeRequest{
			Start:  uint64(hardStart),
			Stop:   uint64(hardStop),
			Params: preRunParams,
		})
	}

	return blockRanges
}

func instantiateLevelDBReader() (pkg.Reader, *params.ChainConfig, node.Info) {
	// load some necessary params
	logWithCommand.Debug("Loading statediff service parameters")
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

	chainConfigPath := viper.GetString("ethereum.chainConfig")
	chainConf, err := utils.LoadConfig(chainConfigPath)
	if err != nil {
		logWithCommand.Fatalf("Unable to instantiate chain config: %s", err)
	}

	// create LevelDB reader
	logWithCommand.Debug("Creating LevelDB reader")
	readerConf := pkg.LvLDBReaderConfig{
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
	reader, err := pkg.NewLvlDBReader(readerConf)
	if err != nil {
		logWithCommand.Fatalf("Unable to instantiate levelDB reader: %s", err)
	}
	return reader, chainConf, nodeInfo
}

// report latest block info
func reportLatestBlock(reader pkg.Reader) {
	header, err := reader.GetLatestHeader()
	if err != nil {
		logWithCommand.Fatalf("Unable to determine latest header height and hash: %s", err.Error())
	}
	if header.Number == nil {
		logWithCommand.Fatal("Latest header found in levelDB has a nil block height")
	}
	logWithCommand.
		WithField("height", header.Number).
		WithField("hash", header.Hash()).
		Info("Latest block found in levelDB")
}
