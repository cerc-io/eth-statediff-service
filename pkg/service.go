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

	"github.com/cerc-io/plugeth-statediff"
	"github.com/cerc-io/plugeth-statediff/adapt"
	"github.com/cerc-io/plugeth-statediff/indexer/interfaces"
	sdtypes "github.com/cerc-io/plugeth-statediff/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"

	"github.com/cerc-io/eth-statediff-service/pkg/prom"
)

const defaultQueueSize = 1024

// Service is the underlying struct for the state diffing service
type Service struct {
	// Used to build the state diff objects
	builder statediff.Builder
	// Used to read data from LevelDB
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
func NewStateDiffService(lvlDBReader Reader, indexer interfaces.StateDiffIndexer, conf ServiceConfig) *Service {
	builder := statediff.NewBuilder(adapt.GethStateView(lvlDBReader.StateDB()))
	builder.SetSubtrieWorkers(conf.TrieWorkers)
	if conf.WorkerQueueSize == 0 {
		conf.WorkerQueueSize = defaultQueueSize
	}
	return &Service{
		lvlDBReader: lvlDBReader,
		builder:     builder,
		indexer:     indexer,
		workers:     conf.ServiceWorkers,
		queue:       make(chan RangeRequest, conf.WorkerQueueSize),
		preruns:     conf.PreRuns,
	}
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

func segmentRange(workers, start, stop uint64, params statediff.Params) []RangeRequest {
	segmentSize := ((stop - start) + 1) / workers
	remainder := ((stop - start) + 1) % workers
	numOfSegments := workers
	if remainder > 0 {
		numOfSegments++
	}
	segments := make([]RangeRequest, numOfSegments)
	for i := range segments {
		end := start + segmentSize - 1
		if end > stop {
			end = stop
		}
		segments[i] = RangeRequest{start, end, params}
		start = end + 1
	}
	return segments
}

// Run does a one-off processing run on the provided RangeRequests + any pre-runs, exiting afterwards
func (sds *Service) Run(rngs []RangeRequest, parallel bool) error {
	for _, preRun := range sds.preruns {
		// if the rangeSize is smaller than the number of workers
		// make sure we do synchronous processing to avoid quantization issues
		rangeSize := (preRun.Stop - preRun.Start) + 1
		numWorkers := uint64(sds.workers)
		if rangeSize < numWorkers {
			parallel = false
		}
		if parallel {
			logrus.Infof("parallel processing prerun range (%d, %d) (%d blocks) divided into %d sized chunks with %d workers", preRun.Start, preRun.Stop,
				rangeSize, rangeSize/numWorkers, numWorkers)
			workChan := make(chan RangeRequest)
			quitChan := make(chan struct{})
			// spin up numWorkers number of worker goroutines
			wg := new(sync.WaitGroup)
			for i := 0; i < int(numWorkers); i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for {
						select {
						case workerSegment := <-workChan:
							for j := workerSegment.Start; j <= workerSegment.Stop; j++ {
								if err := sds.WriteStateDiffAt(j, workerSegment.Params); err != nil {
									logrus.Errorf("error writing statediff at height %d in range (%d, %d) : %v", id, workerSegment.Start, workerSegment.Stop, err)
								}
							}
							logrus.Infof("prerun worker %d finished processing range (%d, %d)", id, workerSegment.Start, workerSegment.Stop)
						case <-quitChan:
							return
						}
					}
				}(i)
			}
			// break range up into segments
			segments := segmentRange(numWorkers, preRun.Start, preRun.Stop, preRun.Params)
			// send the segments to the work channel
			for _, segment := range segments {
				workChan <- segment
			}
			close(quitChan)
			wg.Wait()
		} else {
			logrus.Infof("sequential processing prerun range (%d, %d)", preRun.Start, preRun.Stop)
			for i := preRun.Start; i <= preRun.Stop; i++ {
				if err := sds.WriteStateDiffAt(i, preRun.Params); err != nil {
					return fmt.Errorf("error writing statediff at height %d in range (%d, %d) : %v", i, preRun.Start, preRun.Stop, err)
				}
			}
		}
	}
	sds.preruns = nil
	// At present this code is never called so we have not written the parallel version:
	for _, rng := range rngs {
		logrus.Infof("processing requested range (%d, %d)", rng.Start, rng.Stop)
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
					log := logrus.WithField("range", blockRange).WithField("worker", id)
					log.Debug("processing range")
					prom.DecQueuedRanges()
					for j := blockRange.Start; j <= blockRange.Stop; j++ {
						if err := sds.WriteStateDiffAt(j, blockRange.Params); err != nil {
							log.Errorf("error writing statediff at block %d: %v", j, err)
						}
						select {
						case <-sds.quitChan:
							log.Infof("closing service worker (last processed block: %d)", j)
							return
						default:
							log.Infof("Finished processing block %d", j)
						}
					}
					log.Debugf("Finished processing range")
				case <-sds.quitChan:
					logrus.Debugf("closing the statediff service loop worker %d", id)
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
func (sds *Service) StateDiffAt(blockNumber uint64, params statediff.Params) (*statediff.Payload, error) {
	currentBlock, err := sds.lvlDBReader.GetBlockByNumber(blockNumber)
	if err != nil {
		return nil, err
	}
	logrus.Infof("sending state diff at block %d", blockNumber)

	// compute leaf paths of watched addresses in the params
	params.ComputeWatchedAddressesLeafPaths()

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
func (sds *Service) StateDiffFor(blockHash common.Hash, params statediff.Params) (*statediff.Payload, error) {
	currentBlock, err := sds.lvlDBReader.GetBlockByHash(blockHash)
	if err != nil {
		return nil, err
	}
	logrus.Infof("sending state diff at block %s", blockHash)

	// compute leaf paths of watched addresses in the params
	params.ComputeWatchedAddressesLeafPaths()

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
func (sds *Service) processStateDiff(currentBlock *types.Block, parentRoot common.Hash, params statediff.Params) (*statediff.Payload, error) {
	stateDiff, err := sds.builder.BuildStateDiffObject(statediff.Args{
		BlockHash:    currentBlock.Hash(),
		BlockNumber:  currentBlock.Number(),
		OldStateRoot: parentRoot,
		NewStateRoot: currentBlock.Root(),
	}, params)
	if err != nil {
		return nil, err
	}
	stateDiffRlp, err := rlp.EncodeToBytes(&stateDiff)
	if err != nil {
		return nil, err
	}
	logrus.Infof("state diff object at block %d is %d bytes in length", currentBlock.Number().Uint64(), len(stateDiffRlp))
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
func (sds *Service) WriteStateDiffAt(blockNumber uint64, params statediff.Params) error {
	logrus.Infof("Writing state diff at block %d", blockNumber)
	t := time.Now()
	currentBlock, err := sds.lvlDBReader.GetBlockByNumber(blockNumber)
	if err != nil {
		return err
	}

	// compute leaf paths of watched addresses in the params
	params.ComputeWatchedAddressesLeafPaths()

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
func (sds *Service) WriteStateDiffFor(blockHash common.Hash, params statediff.Params) error {
	logrus.Infof("Writing state diff for block %s", blockHash)
	t := time.Now()
	currentBlock, err := sds.lvlDBReader.GetBlockByHash(blockHash)
	if err != nil {
		return err
	}

	// compute leaf paths of watched addresses in the params
	params.ComputeWatchedAddressesLeafPaths()

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
func (sds *Service) writeStateDiff(block *types.Block, parentRoot common.Hash, params statediff.Params, t time.Time) error {
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
	output := func(node sdtypes.StateLeafNode) error {
		return sds.indexer.PushStateNode(tx, node, block.Hash().String())
	}
	codeOutput := func(c sdtypes.IPLD) error {
		return sds.indexer.PushIPLD(tx, c)
	}
	prom.SetTimeMetric(prom.T_BLOCK_PROCESSING, time.Now().Sub(t))
	t = time.Now()
	err = sds.builder.WriteStateDiff(statediff.Args{
		NewStateRoot: block.Root(),
		OldStateRoot: parentRoot,
		BlockNumber:  block.Number(),
		BlockHash:    block.Hash(),
	}, params, output, codeOutput)
	prom.SetTimeMetric(prom.T_STATE_PROCESSING, time.Now().Sub(t))
	t = time.Now()
	err = tx.Submit()
	prom.SetLastProcessedHeight(height)
	prom.SetTimeMetric(prom.T_POSTGRES_TX_COMMIT, time.Now().Sub(t))
	return err
}

// WriteStateDiffsInRange adds a RangeRequest to the work queue
func (sds *Service) WriteStateDiffsInRange(start, stop uint64, params statediff.Params) error {
	if stop < start {
		return fmt.Errorf("invalid block range (%d, %d): stop height must be greater or equal to start height", start, stop)
	}
	blocked := time.NewTimer(30 * time.Second)
	select {
	case sds.queue <- RangeRequest{Start: start, Stop: stop, Params: params}:
		prom.IncQueuedRanges()
		logrus.Infof("Added range (%d, %d) to the worker queue", start, stop)
		return nil
	case <-blocked.C:
		return fmt.Errorf("unable to add range (%d, %d) to the worker queue", start, stop)
	}
}
