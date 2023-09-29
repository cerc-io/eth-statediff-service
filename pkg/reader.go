// Copyright © 2020 Vulcanize, Inc
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package statediff

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/cerc-io/leveldb-ethdb-rpc/pkg/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie"
)

// Reader interface required by the statediffing service
type Reader interface {
	GetBlockByHash(hash common.Hash) (*types.Block, error)
	GetBlockByNumber(number uint64) (*types.Block, error)
	GetReceiptsByHash(hash common.Hash) (types.Receipts, error)
	GetTdByHash(hash common.Hash) (*big.Int, error)
	StateDB() state.Database
	GetLatestHeader() (*types.Header, error)
}

// LvlDBReader exposes the necessary Reader methods on lvldb
type LvlDBReader struct {
	ethDB       ethdb.Database
	stateDB     state.Database
	chainConfig *params.ChainConfig
}

// LvLDBReaderConfig struct for initializing a LvlDBReader
type LvLDBReaderConfig struct {
	TrieConfig             *trie.Config
	ChainConfig            *params.ChainConfig
	Mode                   string
	Path, AncientPath, Url string
	DBCacheSize            int
}

// NewLvlDBReader creates a new Reader using LevelDB
func NewLvlDBReader(conf LvLDBReaderConfig) (*LvlDBReader, error) {
	var edb ethdb.Database
	var err error

	switch conf.Mode {
	case "local":
		edb, err = rawdb.NewLevelDBDatabase(conf.Path, conf.DBCacheSize, 256, "eth-statediff-service", true)
		if err != nil {
			return nil, err
		}

		edb, err = rawdb.NewDatabaseWithFreezer(edb, conf.AncientPath, "eth-statediff-service", true)
		if err != nil {
			return nil, err
		}
	case "remote":
		edb, err = client.NewDatabaseClient(conf.Url)
		if err != nil {
			return nil, err
		}
	}

	return &LvlDBReader{
		ethDB:       edb,
		stateDB:     state.NewDatabaseWithConfig(edb, conf.TrieConfig),
		chainConfig: conf.ChainConfig,
	}, nil
}

// GetBlockByHash gets block by hash
func (ldr *LvlDBReader) GetBlockByHash(hash common.Hash) (*types.Block, error) {
	height := rawdb.ReadHeaderNumber(ldr.ethDB, hash)
	if height == nil {
		return nil, fmt.Errorf("unable to read header height for header hash %s", hash)
	}
	block := rawdb.ReadBlock(ldr.ethDB, hash, *height)
	if block == nil {
		return nil, fmt.Errorf("unable to read block at height %d hash %s", *height, hash)
	}
	return block, nil
}

func (ldr *LvlDBReader) GetBlockByNumber(number uint64) (*types.Block, error) {
	hash := rawdb.ReadCanonicalHash(ldr.ethDB, number)
	block := rawdb.ReadBlock(ldr.ethDB, hash, number)
	if block == nil {
		return nil, fmt.Errorf("unable to read block at height %d hash %s", number, hash)
	}
	return block, nil
}

// GetReceiptsByHash gets receipt by hash
func (ldr *LvlDBReader) GetReceiptsByHash(hash common.Hash) (types.Receipts, error) {
	number := rawdb.ReadHeaderNumber(ldr.ethDB, hash)
	if number == nil {
		return nil, fmt.Errorf("unable to read header height for header hash %s", hash)
	}
	receipts := rawdb.ReadReceipts(ldr.ethDB, hash, *number, ldr.chainConfig)
	if receipts == nil {
		return nil, fmt.Errorf("unable to read receipts at height %d hash %s", number, hash)
	}
	return receipts, nil
}

// GetTdByHash gets td by hash
func (ldr *LvlDBReader) GetTdByHash(hash common.Hash) (*big.Int, error) {
	number := rawdb.ReadHeaderNumber(ldr.ethDB, hash)
	if number == nil {
		return nil, fmt.Errorf("unable to read header height for header hash %s", hash)
	}
	td := rawdb.ReadTd(ldr.ethDB, hash, *number)
	if td == nil {
		return nil, fmt.Errorf("unable to read total difficulty at height %d hash %s", number, hash)
	}
	return td, nil
}

// StateDB returns the underlying statedb
func (ldr *LvlDBReader) StateDB() state.Database {
	return ldr.stateDB
}

// GetLatestHeader gets the latest header from the levelDB
func (ldr *LvlDBReader) GetLatestHeader() (*types.Header, error) {
	header := rawdb.ReadHeadHeader(ldr.ethDB)
	if header == nil {
		return nil, errors.New("unable to read head header")
	}
	return header, nil
}
