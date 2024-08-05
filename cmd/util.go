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
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/ethereum/go-ethereum/triedb/hashdb"
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

func createReader() (pkg.Reader, *params.ChainConfig, node.Info) {
	// load some necessary params
	logWithCommand.Debug("Loading statediff service parameters")
	path := viper.GetString("ethdb.path")
	ancientPath := viper.GetString("ethdb.ancient")
	url := viper.GetString("ethdb.url")

	if path == "" {
		logWithCommand.Fatal("Require a valid Ethereum chain data path")
	}
	if ancientPath == "" {
		ancientPath = path + "/ancient"
	}

	nodeInfo := getEthNodeInfo()

	chainConfigPath := viper.GetString("ethereum.chainConfig")
	chainConf, err := utils.LoadConfig(chainConfigPath)
	if err != nil {
		logWithCommand.Fatalf("Unable to instantiate chain config: %s", err)
	}

	logWithCommand.Debug("Creating DB reader")
	readerConf := pkg.EthDBReaderConfig{
		TrieConfig: &triedb.Config{
			Preimages: false,
			IsVerkle:  false,
			HashDB: &hashdb.Config{
				CleanCacheSize: viper.GetInt("cache.trie"),
			},
		},
		ChainConfig: chainConf,
		Path:        path,
		AncientPath: ancientPath,
		Url:         url,
		DBCacheSize: viper.GetInt("cache.database"),
	}
	reader, err := pkg.NewEthDBReader(readerConf)
	if err != nil {
		logWithCommand.Fatalf("Unable to instantiate DB reader: %s", err)
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
		logWithCommand.Fatal("Latest header found in DB has a nil block height")
	}
	logWithCommand.
		WithField("height", header.Number).
		WithField("hash", header.Hash()).
		Info("Latest block found in DB")
}
