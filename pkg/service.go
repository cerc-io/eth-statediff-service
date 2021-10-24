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
	"bytes"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	sd "github.com/ethereum/go-ethereum/statediff"
	sdtypes "github.com/ethereum/go-ethereum/statediff/types"
	"github.com/sirupsen/logrus"
	"github.com/vulcanize/eth-statediff-service/pkg/prom"

	ind "github.com/ethereum/go-ethereum/statediff/indexer"
)

// lvlDBReader are the db interfaces required by the statediffing service
type lvlDBReader interface {
	GetBlockByHash(hash common.Hash) (*types.Block, error)
	GetBlockByNumber(number uint64) (*types.Block, error)
	GetReceiptsByHash(hash common.Hash) (types.Receipts, error)
	GetTdByHash(hash common.Hash) (*big.Int, error)
	StateDB() state.Database
}

// IService is the state-diffing service interface
type IService interface {
	// Start() and Stop()
	node.Lifecycle
	// For node service registration
	APIs() []rpc.API
	Protocols() []p2p.Protocol
	// Main event loop for processing state diffs
	Loop(wg *sync.WaitGroup)
	// Method to get state diff object at specific block
	StateDiffAt(blockNumber uint64, params sd.Params) (*sd.Payload, error)
	// Method to get state diff object at specific block
	StateDiffFor(blockHash common.Hash, params sd.Params) (*sd.Payload, error)
	// Method to get state trie object at specific block
	StateTrieAt(blockNumber uint64, params sd.Params) (*sd.Payload, error)
	// Method to write state diff object directly to DB
	WriteStateDiffAt(blockNumber uint64, params sd.Params) error
	// Method to get state trie object at specific block
	WriteStateDiffFor(blockHash common.Hash, params sd.Params) error
}

// Service is the underlying struct for the state diffing service
type Service struct {
	// Used to build the state diff objects
	Builder Builder
	// Used to read data from leveldb
	lvlDBReader lvlDBReader
	// Used to signal shutdown of the service
	QuitChan chan bool
	// Interface for publishing statediffs as PG-IPLD objects
	indexer ind.Indexer
}

// NewStateDiffService creates a new Service
func NewStateDiffService(lvlDBReader lvlDBReader, indexer ind.Indexer, workers uint) (*Service, error) {
	builder, err := NewBuilder(lvlDBReader.StateDB(), workers)
	if err != nil {
		return nil, err
	}
	return &Service{
		lvlDBReader: lvlDBReader,
		Builder:     builder,
		QuitChan:    make(chan bool),
		indexer:     indexer,
	}, nil
}

// Protocols exports the services p2p protocols, this service has none
func (sds *Service) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}
}

// APIs returns the RPC descriptors the Service offers
func (sds *Service) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: APIName,
			Version:   APIVersion,
			Service:   NewPublicStateDiffAPI(sds),
			Public:    true,
		},
	}
}

// Loop is an empty service loop for awaiting rpc requests
func (sds *Service) Loop(wg *sync.WaitGroup) {
	wg.Add(1)
	for {
		select {
		case <-sds.QuitChan:
			logrus.Info("closing the statediff service loop")
			wg.Done()
			return
		}
	}
}

// StateDiffAt returns a state diff object payload at the specific blockheight
// This operation cannot be performed back past the point of db pruning; it requires an archival node for historical data
func (sds *Service) StateDiffAt(blockNumber uint64, params sd.Params) (*sd.Payload, error) {
	currentBlock, err := sds.lvlDBReader.GetBlockByNumber(blockNumber)
	if err != nil {
		return nil, err
	}
	logrus.Info(fmt.Sprintf("sending state diff at block %d", blockNumber))
	if blockNumber == 0 {
		return sds.processStateDiff(currentBlock, common.Hash{}, params)
	}
	parentBlock, err := sds.lvlDBReader.GetBlockByHash(currentBlock.ParentHash())
	if err != nil {
		return nil, err
	}
	return sds.processStateDiff(currentBlock, parentBlock.Root(), params)
}

// StateDiffFor returns a state diff object payload for the specific blockhash
// This operation cannot be performed back past the point of db pruning; it requires an archival node for historical data
func (sds *Service) StateDiffFor(blockHash common.Hash, params sd.Params) (*sd.Payload, error) {
	currentBlock, err := sds.lvlDBReader.GetBlockByHash(blockHash)
	if err != nil {
		return nil, err
	}
	logrus.Info(fmt.Sprintf("sending state diff at block %s", blockHash.Hex()))
	if currentBlock.NumberU64() == 0 {
		return sds.processStateDiff(currentBlock, common.Hash{}, params)
	}
	parentBlock, err := sds.lvlDBReader.GetBlockByHash(currentBlock.ParentHash())
	if err != nil {
		return nil, err
	}
	return sds.processStateDiff(currentBlock, parentBlock.Root(), params)
}

// processStateDiff method builds the state diff payload from the current block, parent state root, and provided params
func (sds *Service) processStateDiff(currentBlock *types.Block, parentRoot common.Hash, params sd.Params) (*sd.Payload, error) {
	stateDiff, err := sds.Builder.BuildStateDiffObject(sd.Args{
		BlockHash:    currentBlock.Hash(),
		BlockNumber:  currentBlock.Number(),
		OldStateRoot: parentRoot,
		NewStateRoot: currentBlock.Root(),
	}, params)
	if err != nil {
		return nil, err
	}
	stateDiffRlp, err := rlp.EncodeToBytes(stateDiff)
	if err != nil {
		return nil, err
	}
	logrus.Infof("state diff object at block %d is %d bytes in length", currentBlock.Number().Uint64(), len(stateDiffRlp))
	return sds.newPayload(stateDiffRlp, currentBlock, params)
}

func (sds *Service) newPayload(stateObject []byte, block *types.Block, params sd.Params) (*sd.Payload, error) {
	payload := &sd.Payload{
		StateObjectRlp: stateObject,
	}
	if params.IncludeBlock {
		blockBuff := new(bytes.Buffer)
		if err := block.EncodeRLP(blockBuff); err != nil {
			return nil, err
		}
		payload.BlockRlp = blockBuff.Bytes()
	}
	if params.IncludeTD {
		var err error
		payload.TotalDifficulty, err = sds.lvlDBReader.GetTdByHash(block.Hash())
		if err != nil {
			return nil, err
		}
	}
	if params.IncludeReceipts {
		receiptBuff := new(bytes.Buffer)
		receipts, err := sds.lvlDBReader.GetReceiptsByHash(block.Hash())
		if err != nil {
			return nil, err
		}
		if err := rlp.Encode(receiptBuff, receipts); err != nil {
			return nil, err
		}
		payload.ReceiptsRlp = receiptBuff.Bytes()
	}
	return payload, nil
}

// StateTrieAt returns a state trie object payload at the specified blockheight
// This operation cannot be performed back past the point of db pruning; it requires an archival node for historical data
func (sds *Service) StateTrieAt(blockNumber uint64, params sd.Params) (*sd.Payload, error) {
	currentBlock, err := sds.lvlDBReader.GetBlockByNumber(blockNumber)
	if err != nil {
		return nil, err
	}
	logrus.Info(fmt.Sprintf("sending state trie at block %d", blockNumber))
	return sds.processStateTrie(currentBlock, params)
}

func (sds *Service) processStateTrie(block *types.Block, params sd.Params) (*sd.Payload, error) {
	stateNodes, err := sds.Builder.BuildStateTrieObject(block)
	if err != nil {
		return nil, err
	}
	stateTrieRlp, err := rlp.EncodeToBytes(stateNodes)
	if err != nil {
		return nil, err
	}
	logrus.Infof("state trie object at block %d is %d bytes in length", block.Number().Uint64(), len(stateTrieRlp))
	return sds.newPayload(stateTrieRlp, block, params)
}

// Start is used to begin the service
func (sds *Service) Start() error {
	logrus.Info("starting statediff service")
	go sds.Loop(new(sync.WaitGroup))
	return nil
}

// Stop is used to close down the service
func (sds *Service) Stop() error {
	logrus.Info("stopping statediff service")
	close(sds.QuitChan)
	return nil
}

// WriteStateDiffAt writes a state diff at the specific blockheight directly to the database
// This operation cannot be performed back past the point of db pruning; it requires an archival node
// for historical data
func (sds *Service) WriteStateDiffAt(blockNumber uint64, params sd.Params) error {
	logrus.Info(fmt.Sprintf("Writing state diff at block %d", blockNumber))
	t := time.Now()
	currentBlock, err := sds.lvlDBReader.GetBlockByNumber(blockNumber)
	if err != nil {
		return err
	}
	parentRoot := common.Hash{}
	if blockNumber != 0 {
		parentBlock, err := sds.lvlDBReader.GetBlockByHash(currentBlock.ParentHash())
		if err != nil {
			return err
		}
		parentRoot = parentBlock.Root()
	}
	return sds.writeStateDiff(currentBlock, parentRoot, params, t)
}

// WriteStateDiffFor writes a state diff for the specific blockHash directly to the database
// This operation cannot be performed back past the point of db pruning; it requires an archival node
// for historical data
func (sds *Service) WriteStateDiffFor(blockHash common.Hash, params sd.Params) error {
	logrus.Info(fmt.Sprintf("Writing state diff for block %s", blockHash.Hex()))
	t := time.Now()
	currentBlock, err := sds.lvlDBReader.GetBlockByHash(blockHash)
	if err != nil {
		return err
	}
	parentRoot := common.Hash{}
	if currentBlock.NumberU64() != 0 {
		parentBlock, err := sds.lvlDBReader.GetBlockByHash(currentBlock.ParentHash())
		if err != nil {
			return err
		}
		parentRoot = parentBlock.Root()
	}
	return sds.writeStateDiff(currentBlock, parentRoot, params, t)
}

// Writes a state diff from the current block, parent state root, and provided params
func (sds *Service) writeStateDiff(block *types.Block, parentRoot common.Hash, params sd.Params, t time.Time) error {
	var totalDifficulty *big.Int
	var receipts types.Receipts
	var err error
	if params.IncludeTD {
		totalDifficulty, err = sds.lvlDBReader.GetTdByHash(block.Hash())
	}
	if err != nil {
		return err
	}
	if params.IncludeReceipts {
		receipts, err = sds.lvlDBReader.GetReceiptsByHash(block.Hash())
	}
	if err != nil {
		return err
	}
	height := block.Number().Int64()
	prom.SetLastLoadedHeight(height)
	prom.SetTimeMetric(prom.T_BLOCK_LOAD, time.Now().Sub(t))
	t = time.Now()
	tx, err := sds.indexer.PushBlock(block, receipts, totalDifficulty)
	if err != nil {
		return err
	}
	// defer handling of commit/rollback for any return case
	output := func(node sdtypes.StateNode) error {
		return sds.indexer.PushStateNode(tx, node)
	}
	codeOutput := func(c sdtypes.CodeAndCodeHash) error {
		return sds.indexer.PushCodeAndCodeHash(tx, c)
	}
	prom.SetTimeMetric(prom.T_BLOCK_PROCESSING, time.Now().Sub(t))
	t = time.Now()
	err = sds.Builder.WriteStateDiffObject(sd.StateRoots{
		NewStateRoot: block.Root(),
		OldStateRoot: parentRoot,
	}, params, output, codeOutput)
	prom.SetTimeMetric(prom.T_STATE_PROCESSING, time.Now().Sub(t))
	t = time.Now()
	err = tx.Close(err)
	prom.SetLastProcessedHeight(height)
	prom.SetTimeMetric(prom.T_POSTGRES_TX_COMMIT, time.Now().Sub(t))
	return err
}
