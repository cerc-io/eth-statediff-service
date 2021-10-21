package cmd

import (
	ind "github.com/ethereum/go-ethereum/statediff/indexer"
	"github.com/ethereum/go-ethereum/statediff/indexer/node"
	"github.com/ethereum/go-ethereum/statediff/indexer/postgres"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/spf13/viper"

	sd "github.com/vulcanize/eth-statediff-service/pkg"
	"github.com/vulcanize/eth-statediff-service/pkg/prom"
)

func createStateDiffService() (sd.IService, error) {
	logWithCommand.Info("Loading statediff service parameters")
	path := viper.GetString("leveldb.path")
	ancientPath := viper.GetString("leveldb.ancient")
	if path == "" || ancientPath == "" {
		logWithCommand.Fatal("require a valid eth leveldb primary datastore path and ancient datastore path")
	}

	nodeInfo := GetEthNodeInfo()
	config, err := chainConfig(nodeInfo.ChainID)
	if err != nil {
		logWithCommand.Fatal(err)
	}

	// create leveldb reader
	logWithCommand.Info("Creating leveldb reader")
	conf := sd.ReaderConfig{
		TrieConfig: &trie.Config{
			Cache:     viper.GetInt("cache.trie"),
			Journal:   "",
			Preimages: false,
		},
		ChainConfig: config,
		Path:        path,
		AncientPath: ancientPath,
		DBCacheSize: viper.GetInt("cache.database"),
	}
	lvlDBReader, err := sd.NewLvlDBReader(conf)
	if err != nil {
		logWithCommand.Fatal(err)
	}

	// create statediff service
	logWithCommand.Info("Creating statediff service")
	db, err := setupPostgres(nodeInfo)
	if err != nil {
		logWithCommand.Fatal(err)
	}
	indexer, err := ind.NewStateDiffIndexer(config, db)
	if err != nil {
		logWithCommand.Fatal(err)
	}
	return sd.NewStateDiffService(lvlDBReader, indexer, viper.GetUint("statediff.workers"))
}

func setupPostgres(nodeInfo node.Info) (*postgres.DB, error) {
	params := GetDBParams()
	db, err := postgres.NewDB(postgres.DbConnectionString(params), GetDBConfig(), nodeInfo)
	if err != nil {
		return nil, err
	}
	prom.RegisterDBCollector(params.Name, db.DB)
	return db, nil
}
