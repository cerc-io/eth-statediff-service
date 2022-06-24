// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Contains a batch of utility type declarations used by the tests. As the node
// operates on unique types, a lot of them are needed to check various features.

package statediff

import (
	"fmt"
	"math/bits"
	"sync"

	"github.com/ethereum/go-ethereum/core/state"
	sd "github.com/ethereum/go-ethereum/statediff"
	sdtypes "github.com/ethereum/go-ethereum/statediff/types"
	"github.com/sirupsen/logrus"

	iter "github.com/vulcanize/go-eth-state-node-iterator"
)

type builder struct {
	sd.StateDiffBuilder
	numWorkers uint
}

// NewBuilder is used to create a statediff builder
func NewBuilder(stateCache state.Database, workers uint) (sd.Builder, error) {
	if workers == 0 {
		workers = 1
	}
	if bits.OnesCount(workers) != 1 {
		return nil, fmt.Errorf("workers must be a power of 2")
	}
	return &builder{
		StateDiffBuilder: sd.StateDiffBuilder{
			StateCache: stateCache,
		},
		numWorkers: workers,
	}, nil
}

// BuildStateDiffObject builds a statediff object from two blocks and the provided parameters
func (sdb *builder) BuildStateDiffObject(args sd.Args, params sd.Params) (sdtypes.StateObject, error) {
	var stateNodes []sdtypes.StateNode
	var codeAndCodeHashes []sdtypes.CodeAndCodeHash
	err := sdb.WriteStateDiffObject(
		sdtypes.StateRoots{OldStateRoot: args.OldStateRoot, NewStateRoot: args.NewStateRoot},
		params, sd.StateNodeAppender(&stateNodes), sd.CodeMappingAppender(&codeAndCodeHashes))
	if err != nil {
		return sdtypes.StateObject{}, err
	}
	return sdtypes.StateObject{
		BlockHash:         args.BlockHash,
		BlockNumber:       args.BlockNumber,
		Nodes:             stateNodes,
		CodeAndCodeHashes: codeAndCodeHashes,
	}, nil
}

// WriteStateDiffObject writes a statediff object to output callback
func (sdb *builder) WriteStateDiffObject(args sdtypes.StateRoots, params sd.Params, output sdtypes.StateNodeSink, codeOutput sdtypes.CodeSink) error {
	// Load tries for old and new states
	oldTrie, err := sdb.StateCache.OpenTrie(args.OldStateRoot)
	if err != nil {
		return fmt.Errorf("error creating trie for oldStateRoot: %v", err)
	}
	newTrie, err := sdb.StateCache.OpenTrie(args.NewStateRoot)
	if err != nil {
		return fmt.Errorf("error creating trie for newStateRoot: %v", err)
	}

	// Split old and new tries into corresponding subtrie iterators
	oldIters1 := iter.SubtrieIterators(oldTrie, sdb.numWorkers)
	oldIters2 := iter.SubtrieIterators(oldTrie, sdb.numWorkers)
	newIters1 := iter.SubtrieIterators(newTrie, sdb.numWorkers)
	newIters2 := iter.SubtrieIterators(newTrie, sdb.numWorkers)

	// Create iterators ahead of time to avoid race condition in state.Trie access
	// We do two state iterations per subtrie: one for new/updated nodes,
	// one for deleted/updated nodes; prepare 2 iterator instances for each task
	var iterPairs [][]sd.IterPair
	for i := uint(0); i < sdb.numWorkers; i++ {
		iterPairs = append(iterPairs, []sd.IterPair{
			{Older: oldIters1[i], Newer: newIters1[i]},
			{Older: oldIters2[i], Newer: newIters2[i]},
		})
	}

	// Dispatch workers to process trie data; sync and collect results here via channels
	nodeChan := make(chan sdtypes.StateNode)
	codeChan := make(chan sdtypes.CodeAndCodeHash)

	go func() {
		nodeSender := func(node sdtypes.StateNode) error { nodeChan <- node; return nil }
		codeSender := func(code sdtypes.CodeAndCodeHash) error { codeChan <- code; return nil }
		var wg sync.WaitGroup

		for w := uint(0); w < sdb.numWorkers; w++ {
			wg.Add(1)
			go func(worker uint) {
				defer wg.Done()
				var err error
				if !params.IntermediateStateNodes || len(params.WatchedAddresses) > 0 {
					// if we are watching only specific accounts then we are only diffing leaf nodes
					err = sdb.BuildStateDiffWithoutIntermediateStateNodes(iterPairs[worker], params, nodeSender, codeSender)
				} else {
					err = sdb.BuildStateDiffWithIntermediateStateNodes(iterPairs[worker], params, nodeSender, codeSender)
				}

				if err != nil {
					logrus.Errorf("buildStateDiff error for worker %d, params %+v", worker, params)
				}
			}(w)
		}
		wg.Wait()
		close(nodeChan)
		close(codeChan)
	}()

	for nodeChan != nil || codeChan != nil {
		select {
		case node, more := <-nodeChan:
			if more {
				if err := output(node); err != nil {
					return err
				}
			} else {
				nodeChan = nil
			}
		case codeAndCodeHash, more := <-codeChan:
			if more {
				if err := codeOutput(codeAndCodeHash); err != nil {
					return err
				}
			} else {
				codeChan = nil
			}
		}
	}

	return nil
}
