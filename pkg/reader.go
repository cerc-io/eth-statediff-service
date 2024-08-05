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

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/triedb"
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

// EthDBReader exposes the necessary Reader methods on an ethdb
type EthDBReader struct {
	ethDB       ethdb.Database
	stateDB     state.Database
	chainConfig *params.ChainConfig
}

type EthDBReaderConfig struct {
	TrieConfig             *triedb.Config
	ChainConfig            *params.ChainConfig
	Path, AncientPath, Url string
	DBCacheSize            int
}

// NewEthDBReader creates a new Reader using LevelDB
func NewEthDBReader(conf EthDBReaderConfig) (*EthDBReader, error) {
	opts := rawdb.OpenOptions{
		Directory:         conf.Path,
		AncientsDirectory: conf.AncientPath,
		Namespace:         "eth-statediff-service",
		Cache:             conf.DBCacheSize,
		Handles:           256,
		ReadOnly:          true,
	}
	edb, err := rawdb.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}

	return &EthDBReader{
		ethDB:       edb,
		stateDB:     state.NewDatabaseWithConfig(edb, conf.TrieConfig),
		chainConfig: conf.ChainConfig,
	}, nil
}

// GetBlockByHash gets block by hash
func (ldr *EthDBReader) GetBlockByHash(hash common.Hash) (*types.Block, error) {
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

func (ldr *EthDBReader) GetBlockByNumber(number uint64) (*types.Block, error) {
	hash := rawdb.ReadCanonicalHash(ldr.ethDB, number)
	block := rawdb.ReadBlock(ldr.ethDB, hash, number)
	if block == nil {
		return nil, fmt.Errorf("unable to read block at height %d hash %s", number, hash)
	}
	return block, nil
}

// GetReceiptsByHash gets receipt by hash
func (ldr *EthDBReader) GetReceiptsByHash(hash common.Hash) (types.Receipts, error) {
	number := rawdb.ReadHeaderNumber(ldr.ethDB, hash)
	if number == nil {
		return nil, fmt.Errorf("unable to read header height for header hash %s", hash)
	}
	header := rawdb.ReadHeader(ldr.ethDB, hash, *number)
	if header == nil {
		return nil, fmt.Errorf("unable to read header for header hash %s", hash)
	}
	receipts := rawdb.ReadReceipts(ldr.ethDB, hash, *number, header.Time, ldr.chainConfig)
	if receipts == nil {
		return nil, fmt.Errorf("unable to read receipts at height %d hash %s", number, hash)
	}
	return receipts, nil
}

// GetTdByHash gets td by hash
func (ldr *EthDBReader) GetTdByHash(hash common.Hash) (*big.Int, error) {
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
func (ldr *EthDBReader) StateDB() state.Database {
	return ldr.stateDB
}

// GetLatestHeader gets the latest header from the levelDB
func (ldr *EthDBReader) GetLatestHeader() (*types.Header, error) {
	header := rawdb.ReadHeadHeader(ldr.ethDB)
	if header == nil {
		return nil, errors.New("unable to read head header")
	}
	return header, nil
}
