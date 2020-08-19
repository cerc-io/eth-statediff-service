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

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/statediff"
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
	// APIs(), Protocols(), Start() and Stop()
	node.Service
	// Main event loop for processing state diffs
	Loop(wg *sync.WaitGroup)
	// Method to get state diff object at specific block
	StateDiffAt(blockNumber uint64, params statediff.Params) (*statediff.Payload, error)
	// Method to get state trie object at specific block
	StateTrieAt(blockNumber uint64, params statediff.Params) (*statediff.Payload, error)
}

// Service is the underlying struct for the state diffing service
type Service struct {
	// Used to build the state diff objects
	Builder statediff.Builder
	// Used to read data from leveldb
	lvlDBReader lvlDBReader
	// Used to signal shutdown of the service
	QuitChan chan bool
}

// NewStateDiffService creates a new statediff.Service
func NewStateDiffService(lvlDBReader lvlDBReader) (*Service, error) {
	return &Service{
		lvlDBReader: lvlDBReader,
		Builder:     statediff.NewBuilder(lvlDBReader.StateDB()),
		QuitChan:    make(chan bool),
	}, nil
}

// Protocols exports the services p2p protocols, this service has none
func (sds *Service) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}
}

// APIs returns the RPC descriptors the statediff.Service offers
func (sds *Service) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: statediff.APIName,
			Version:   statediff.APIVersion,
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
			log.Info("closing the statediff service loop")
			wg.Done()
			return
		}
	}
}

// StateDiffAt returns a state diff object payload at the specific blockheight
// This operation cannot be performed back past the point of db pruning; it requires an archival node for historical data
func (sds *Service) StateDiffAt(blockNumber uint64, params statediff.Params) (*statediff.Payload, error) {
	currentBlock, err := sds.lvlDBReader.GetBlockByNumber(blockNumber)
	if err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("sending state diff at block %d", blockNumber))
	if blockNumber == 0 {
		return sds.processStateDiff(currentBlock, common.Hash{}, params)
	}
	parentBlock, err := sds.lvlDBReader.GetBlockByHash(currentBlock.ParentHash())
	if err != nil {
		return nil, err
	}
	return sds.processStateDiff(currentBlock, parentBlock.Root(), params)
}

// processStateDiff method builds the state diff payload from the current block, parent state root, and provided params
func (sds *Service) processStateDiff(currentBlock *types.Block, parentRoot common.Hash, params statediff.Params) (*statediff.Payload, error) {
	stateDiff, err := sds.Builder.BuildStateDiffObject(statediff.Args{
		NewStateRoot: currentBlock.Root(),
		OldStateRoot: parentRoot,
		BlockHash:    currentBlock.Hash(),
		BlockNumber:  currentBlock.Number(),
	}, params)
	if err != nil {
		return nil, err
	}
	stateDiffRlp, err := rlp.EncodeToBytes(stateDiff)
	if err != nil {
		return nil, err
	}
	return sds.newPayload(stateDiffRlp, currentBlock, params)
}

func (sds *Service) newPayload(stateObject []byte, block *types.Block, params statediff.Params) (*statediff.Payload, error) {
	payload := &statediff.Payload{
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
func (sds *Service) StateTrieAt(blockNumber uint64, params statediff.Params) (*statediff.Payload, error) {
	currentBlock, err := sds.lvlDBReader.GetBlockByNumber(blockNumber)
	if err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("sending state trie at block %d", blockNumber))
	return sds.processStateTrie(currentBlock, params)
}

func (sds *Service) processStateTrie(block *types.Block, params statediff.Params) (*statediff.Payload, error) {
	stateNodes, err := sds.Builder.BuildStateTrieObject(block)
	if err != nil {
		return nil, err
	}
	stateTrieRlp, err := rlp.EncodeToBytes(stateNodes)
	if err != nil {
		return nil, err
	}
	return sds.newPayload(stateTrieRlp, block, params)
}

// Start is used to begin the service
func (sds *Service) Start(*p2p.Server) error {
	log.Info("starting statediff service")
	go sds.Loop(new(sync.WaitGroup))
	return nil
}

// Stop is used to close down the service
func (sds *Service) Stop() error {
	log.Info("stopping statediff service")
	close(sds.QuitChan)
	return nil
}
