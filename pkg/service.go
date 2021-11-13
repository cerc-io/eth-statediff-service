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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	sd "github.com/ethereum/go-ethereum/statediff"
	"github.com/ethereum/go-ethereum/statediff/indexer/interfaces"
	sdtypes "github.com/ethereum/go-ethereum/statediff/types"
	"github.com/sirupsen/logrus"

	"github.com/vulcanize/eth-statediff-service/pkg/prom"
)

const defaultQueueSize = 1024

// StateDiffService is the state-diffing service interface
type StateDiffService interface {
	// Lifecycle Start() and Stop()
	node.Lifecycle
	// APIs and Protocols() interface for node service registration
	APIs() []rpc.API
	Protocols() []p2p.Protocol
	// Loop is the main event loop for processing state diffs
	Loop(wg *sync.WaitGroup) error
	// Run is a one-off command to run on a predefined set of ranges
	Run(ranges []RangeRequest) error
	// StateDiffAt method to get state diff object at specific block
	StateDiffAt(blockNumber uint64, params sd.Params) (*sd.Payload, error)
	// StateDiffFor method to get state diff object at specific block
	StateDiffFor(blockHash common.Hash, params sd.Params) (*sd.Payload, error)
	// StateTrieAt method to get state trie object at specific block
	StateTrieAt(blockNumber uint64, params sd.Params) (*sd.Payload, error)
	// WriteStateDiffAt method to write state diff object directly to DB
	WriteStateDiffAt(blockNumber uint64, params sd.Params) error
	// WriteStateDiffFor method to get state trie object at specific block
	WriteStateDiffFor(blockHash common.Hash, params sd.Params) error
	// WriteStateDiffsInRange method to wrtie state diff objects within the range directly to the DB
	WriteStateDiffsInRange(start, stop uint64, params sd.Params) error
}

// Service is the underlying struct for the state diffing service
type Service struct {
	// Used to build the state diff objects
	Builder Builder
	// Used to read data from leveldb
	lvlDBReader Reader
	// Used to signal shutdown of the service
	quitChan chan struct{}
	// Interface for publishing statediffs as PG-IPLD objects
	indexer interfaces.StateDiffIndexer
	// range queue
	queue chan RangeRequest
	// number of ranges we can work over concurrently
	workers uint
	// ranges configured locally
	preruns []RangeRequest
}

// NewStateDiffService creates a new Service
func NewStateDiffService(lvlDBReader Reader, indexer interfaces.StateDiffIndexer, conf Config) (*Service, error) {
	b, err := NewBuilder(lvlDBReader.StateDB(), conf.TrieWorkers)
	if err != nil {
		return nil, err
	}
	if conf.WorkerQueueSize == 0 {
		conf.WorkerQueueSize = defaultQueueSize
	}
	return &Service{
		lvlDBReader: lvlDBReader,
		Builder:     b,
		indexer:     indexer,
		workers:     conf.ServiceWorkers,
		queue:       make(chan RangeRequest, conf.WorkerQueueSize),
		preruns:     conf.PreRuns,
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

// Run does a one-off processing run on the provided RangeRequests + any pre-runs, exiting afterwards
func (sds *Service) Run(rngs []RangeRequest) error {
	for _, preRun := range sds.preruns {
		logrus.Infof("processing prerun range (%d, %d)", preRun.Start, preRun.Stop)
		for i := preRun.Start; i <= preRun.Stop; i++ {
			if err := sds.WriteStateDiffAt(i, preRun.Params); err != nil {
				return fmt.Errorf("error writing statediff at height %d in range (%d, %d) : %v", i, preRun.Start, preRun.Stop, err)
			}
		}
	}
	sds.preruns = nil
	for _, rng := range rngs {
		logrus.Infof("processing prerun range (%d, %d)", rng.Start, rng.Stop)
		for i := rng.Start; i <= rng.Stop; i++ {
			if err := sds.WriteStateDiffAt(i, rng.Params); err != nil {
				return fmt.Errorf("error writing statediff at height %d in range (%d, %d) : %v", i, rng.Start, rng.Stop, err)
			}
		}
	}
	return nil
}

// Loop is an empty service loop for awaiting rpc requests
func (sds *Service) Loop(wg *sync.WaitGroup) error {
	if sds.quitChan != nil {
		return fmt.Errorf("service loop is already running")
	}

	sds.quitChan = make(chan struct{})
	for i := 0; i < int(sds.workers); i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case blockRange := <-sds.queue:
					logrus.Infof("service worker %d received range (%d, %d) off of work queue, beginning processing", id, blockRange.Start, blockRange.Stop)
					prom.DecQueuedRanges()
					for j := blockRange.Start; j <= blockRange.Stop; j++ {
						if err := sds.WriteStateDiffAt(j, blockRange.Params); err != nil {
							logrus.Errorf("service worker %d error writing statediff at height %d in range (%d, %d) : %v", id, j, blockRange.Start, blockRange.Stop, err)
						}
						select {
						case <-sds.quitChan:
							logrus.Infof("closing service worker %d\n"+
								"working in range (%d, %d)\n"+
								"last processed height: %d", id, blockRange.Start, blockRange.Stop, j)
							return
						default:
							logrus.Infof("service worker %d finished processing statediff height %d in range (%d, %d)", id, j, blockRange.Start, blockRange.Stop)
						}
					}
					logrus.Infof("service worker %d finished processing range (%d, %d)", id, blockRange.Start, blockRange.Stop)
				case <-sds.quitChan:
					logrus.Infof("closing the statediff service loop worker %d", id)
					return
				}
			}
		}(i)
	}
	for _, preRun := range sds.preruns {
		if err := sds.WriteStateDiffsInRange(preRun.Start, preRun.Stop, preRun.Params); err != nil {
			close(sds.quitChan)
			return err
		}
	}
	return nil
}

// StateDiffAt returns a state diff object payload at the specific blockheight
// This operation cannot be performed back past the point of db pruning; it requires an archival node for historical data
func (sds *Service) StateDiffAt(blockNumber uint64, params sd.Params) (*sd.Payload, error) {
	currentBlock, err := sds.lvlDBReader.GetBlockByNumber(blockNumber)
	if err != nil {
		return nil, err
	}
	logrus.Infof("sending state diff at block %d", blockNumber)
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
	logrus.Infof("sending state diff at block %s", blockHash.Hex())
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
	logrus.Infof("sending state trie at block %d", blockNumber)
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
	return sds.Loop(new(sync.WaitGroup))
}

// Stop is used to close down the service
func (sds *Service) Stop() error {
	logrus.Info("stopping statediff service")
	close(sds.quitChan)
	return nil
}

// WriteStateDiffAt writes a state diff at the specific blockheight directly to the database
// This operation cannot be performed back past the point of db pruning; it requires an archival node
// for historical data
func (sds *Service) WriteStateDiffAt(blockNumber uint64, params sd.Params) error {
	logrus.Infof("Writing state diff at block %d", blockNumber)
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
	logrus.Infof("Writing state diff for block %s", blockHash.Hex())
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
	err = sds.Builder.WriteStateDiffObject(sdtypes.StateRoots{
		NewStateRoot: block.Root(),
		OldStateRoot: parentRoot,
	}, params, output, codeOutput)
	prom.SetTimeMetric(prom.T_STATE_PROCESSING, time.Now().Sub(t))
	t = time.Now()
	err = tx.Submit(err)
	prom.SetLastProcessedHeight(height)
	prom.SetTimeMetric(prom.T_POSTGRES_TX_COMMIT, time.Now().Sub(t))
	return err
}

// WriteStateDiffsInRange adds a RangeRequest to the work queue
func (sds *Service) WriteStateDiffsInRange(start, stop uint64, params sd.Params) error {
	if stop < start {
		return fmt.Errorf("invalid block range (%d, %d): stop height must be greater or equal to start height", start, stop)
	}
	blocked := time.NewTimer(30 * time.Second)
	select {
	case sds.queue <- RangeRequest{Start: start, Stop: stop, Params: params}:
		prom.IncQueuedRanges()
		logrus.Infof("added range (%d, %d) to the worker queue", start, stop)
		return nil
	case <-blocked.C:
		return fmt.Errorf("unable to add range (%d, %d) to the worker queue", start, stop)
	}
}
