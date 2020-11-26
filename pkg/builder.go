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
	"bytes"
	"fmt"
	"math/bits"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"

	sd "github.com/ethereum/go-ethereum/statediff"
	sdtypes "github.com/ethereum/go-ethereum/statediff/types"
	iter "github.com/vulcanize/go-eth-state-node-iterator"
)

var (
	nullHashBytes     = common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000000")
	emptyNode, _      = rlp.EncodeToBytes([]byte{})
	emptyContractRoot = crypto.Keccak256Hash(emptyNode)
	nullCodeHash      = crypto.Keccak256Hash([]byte{}).Bytes()
)

// Builder interface exposes the method for building a state diff between two blocks
type Builder interface {
	BuildStateDiffObject(args sd.Args, params sd.Params) (sd.StateObject, error)
	BuildStateTrieObject(current *types.Block) (sd.StateObject, error)
	WriteStateDiffObject(args sd.StateRoots, params sd.Params, output sdtypes.StateNodeSink, codeOutput sdtypes.CodeSink) error
}

type builder struct {
	stateCache state.Database
	numWorkers uint
}

type iterPair struct {
	older, newer trie.NodeIterator
}

func resolveNode(it trie.NodeIterator, trieDB *trie.Database) (sdtypes.StateNode, []interface{}, error) {
	nodePath := make([]byte, len(it.Path()))
	copy(nodePath, it.Path())
	node, err := trieDB.Node(it.Hash())
	if err != nil {
		return sdtypes.StateNode{}, nil, err
	}
	var nodeElements []interface{}
	if err := rlp.DecodeBytes(node, &nodeElements); err != nil {
		return sdtypes.StateNode{}, nil, err
	}
	ty, err := sd.CheckKeyType(nodeElements)
	if err != nil {
		return sdtypes.StateNode{}, nil, err
	}
	return sdtypes.StateNode{
		NodeType:  ty,
		Path:      nodePath,
		NodeValue: node,
	}, nodeElements, nil
}

// convenience
func stateNodeAppender(nodes *[]sdtypes.StateNode) sdtypes.StateNodeSink {
	return func(node sdtypes.StateNode) error {
		*nodes = append(*nodes, node)
		return nil
	}
}
func storageNodeAppender(nodes *[]sdtypes.StorageNode) sdtypes.StorageNodeSink {
	return func(node sdtypes.StorageNode) error {
		*nodes = append(*nodes, node)
		return nil
	}
}
func codeMappingAppender(data *[]sdtypes.CodeAndCodeHash) sdtypes.CodeSink {
	return func(c sdtypes.CodeAndCodeHash) error {
		*data = append(*data, c)
		return nil
	}
}

// NewBuilder is used to create a statediff builder
func NewBuilder(stateCache state.Database, workers uint) (Builder, error) {
	if workers == 0 {
		workers = 1
	}
	if bits.OnesCount(workers) != 1 {
		return nil, fmt.Errorf("workers must be a power of 2")
	}
	return &builder{
		stateCache: stateCache, // state cache is safe for concurrent reads
		numWorkers: workers,
	}, nil
}

// BuildStateTrieObject builds a state trie object from the provided block
func (sdb *builder) BuildStateTrieObject(current *types.Block) (sd.StateObject, error) {
	currentTrie, err := sdb.stateCache.OpenTrie(current.Root())
	if err != nil {
		return sd.StateObject{}, fmt.Errorf("error creating trie for block %d: %v", current.Number(), err)
	}
	it := currentTrie.NodeIterator([]byte{})
	stateNodes, codeAndCodeHashes, err := sdb.buildStateTrie(it)
	if err != nil {
		return sd.StateObject{}, fmt.Errorf("error collecting state nodes for block %d: %v", current.Number(), err)
	}
	return sd.StateObject{
		BlockNumber:       current.Number(),
		BlockHash:         current.Hash(),
		Nodes:             stateNodes,
		CodeAndCodeHashes: codeAndCodeHashes,
	}, nil
}

func (sdb *builder) buildStateTrie(it trie.NodeIterator) ([]sdtypes.StateNode, []sdtypes.CodeAndCodeHash, error) {
	stateNodes := make([]sdtypes.StateNode, 0)
	codeAndCodeHashes := make([]sdtypes.CodeAndCodeHash, 0)
	for it.Next(true) {
		// skip value nodes
		if it.Leaf() || bytes.Equal(nullHashBytes, it.Hash().Bytes()) {
			continue
		}
		node, nodeElements, err := resolveNode(it, sdb.stateCache.TrieDB())
		if err != nil {
			return nil, nil, err
		}
		switch node.NodeType {
		case sdtypes.Leaf:
			var account state.Account
			if err := rlp.DecodeBytes(nodeElements[1].([]byte), &account); err != nil {
				return nil, nil, fmt.Errorf("error decoding account for leaf node at path %x nerror: %v", node.Path, err)
			}
			partialPath := trie.CompactToHex(nodeElements[0].([]byte))
			valueNodePath := append(node.Path, partialPath...)
			encodedPath := trie.HexToCompact(valueNodePath)
			leafKey := encodedPath[1:]
			node.LeafKey = leafKey
			if !bytes.Equal(account.CodeHash, nullCodeHash) {
				var storageNodes []sdtypes.StorageNode
				err := sdb.buildStorageNodesEventual(account.Root, nil, true, storageNodeAppender(&storageNodes))
				if err != nil {
					return nil, nil, fmt.Errorf("failed building eventual storage diffs for account %+v\r\nerror: %v", account, err)
				}
				node.StorageNodes = storageNodes
				// emit codehash => code mappings for cod
				codeHash := common.BytesToHash(account.CodeHash)
				code, err := sdb.stateCache.ContractCode(common.Hash{}, codeHash)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to retrieve code for codehash %s\r\n error: %v", codeHash.String(), err)
				}
				codeAndCodeHashes = append(codeAndCodeHashes, sdtypes.CodeAndCodeHash{
					Hash: codeHash,
					Code: code,
				})
			}
			stateNodes = append(stateNodes, node)
		case sdtypes.Extension, sdtypes.Branch:
			stateNodes = append(stateNodes, node)
		default:
			return nil, nil, fmt.Errorf("unexpected node type %s", node.NodeType)
		}
	}
	return stateNodes, codeAndCodeHashes, it.Error()
}

// BuildStateDiffObject builds a statediff object from two blocks and the provided parameters
func (sdb *builder) BuildStateDiffObject(args sd.Args, params sd.Params) (sd.StateObject, error) {
	var stateNodes []sdtypes.StateNode
	var codeAndCodeHashes []sdtypes.CodeAndCodeHash
	err := sdb.WriteStateDiffObject(
		sd.StateRoots{OldStateRoot: args.OldStateRoot, NewStateRoot: args.NewStateRoot},
		params, stateNodeAppender(&stateNodes), codeMappingAppender(&codeAndCodeHashes))
	if err != nil {
		return sd.StateObject{}, err
	}
	return sd.StateObject{
		BlockHash:         args.BlockHash,
		BlockNumber:       args.BlockNumber,
		Nodes:             stateNodes,
		CodeAndCodeHashes: codeAndCodeHashes,
	}, nil
}

// Writes a statediff object to output callback
func (sdb *builder) WriteStateDiffObject(args sd.StateRoots, params sd.Params, output sdtypes.StateNodeSink, codeOutput sdtypes.CodeSink) error {
	if len(params.WatchedAddresses) > 0 {
		// if we are watching only specific accounts then we are only diffing leaf nodes
		log.Info("Ignoring intermediate state nodes because WatchedAddresses was passed")
		params.IntermediateStateNodes = false
	}

	// Load tries for old and new states
	oldTrie, err := sdb.stateCache.OpenTrie(args.OldStateRoot)
	if err != nil {
		return fmt.Errorf("error creating trie for old state root: %v", err)
	}
	newTrie, err := sdb.stateCache.OpenTrie(args.NewStateRoot)
	if err != nil {
		return fmt.Errorf("error creating trie for new state root: %v", err)
	}

	// Split old and new tries into corresponding subtrie iterators
	oldIterFac := iter.NewSubtrieIteratorFactory(oldTrie, sdb.numWorkers)
	newIterFac := iter.NewSubtrieIteratorFactory(newTrie, sdb.numWorkers)

	// Create iterators ahead of time to avoid race condition in state.Trie access
	// We do two state iterations per subtrie: one for new/updated nodes,
	// one for deleted/updated nodes; prepare 2 iterator instances for each task
	var iterPairs [][]iterPair
	for i := uint(0); i < sdb.numWorkers; i++ {
		iterPairs = append(iterPairs, []iterPair{
			iterPair{older: oldIterFac.IteratorAt(i), newer: newIterFac.IteratorAt(i)},
			iterPair{older: oldIterFac.IteratorAt(i), newer: newIterFac.IteratorAt(i)},
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
				sdb.buildStateDiff(iterPairs[worker], params, nodeSender, codeSender)
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

func (sdb *builder) buildStateDiff(args []iterPair, params sd.Params, output sdtypes.StateNodeSink, codeOutput sdtypes.CodeSink) error {
	// collect a slice of all the intermediate nodes that were touched and exist at B
	// a map of their leafkey to all the accounts that were touched and exist at B
	// and a slice of all the paths for the nodes in both of the above sets
	var diffAccountsAtB AccountMap
	var diffPathsAtB map[string]bool
	var err error
	if params.IntermediateStateNodes {
		diffAccountsAtB, diffPathsAtB, err = sdb.createdAndUpdatedStateWithIntermediateNodes(args[0], output)
	} else {
		diffAccountsAtB, diffPathsAtB, err = sdb.createdAndUpdatedState(args[0], params.WatchedAddresses)
	}

	if err != nil {
		return fmt.Errorf("error collecting createdAndUpdatedNodes: %v", err)
	}

	// collect a slice of all the nodes that existed at a path in A that doesn't exist in B
	// a map of their leafkey to all the accounts that were touched and exist at A
	diffAccountsAtA, err := sdb.deletedOrUpdatedState(args[1], diffPathsAtB, output)
	if err != nil {
		return fmt.Errorf("error collecting deletedOrUpdatedNodes: %v", err)
	}

	// collect and sort the leafkey keys for both account mappings into a slice
	createKeys := sortKeys(diffAccountsAtB)
	deleteKeys := sortKeys(diffAccountsAtA)

	// and then find the intersection of these keys
	// these are the leafkeys for the accounts which exist at both A and B but are different
	// this also mutates the passed in createKeys and deleteKeys, removing the intersection keys
	// and leaving the truly created or deleted keys in place
	updatedKeys := findIntersection(createKeys, deleteKeys)

	// build the diff nodes for the updated accounts using the mappings at both A and B as directed by the keys found as the intersection of the two
	err = sdb.buildAccountUpdates(
		diffAccountsAtB, diffAccountsAtA, updatedKeys,
		params.WatchedStorageSlots, params.IntermediateStorageNodes, output)
	if err != nil {
		return fmt.Errorf("error building diff for updated accounts: %v", err)
	}
	// build the diff nodes for created accounts
	err = sdb.buildAccountCreations(diffAccountsAtB, params.WatchedStorageSlots, params.IntermediateStorageNodes, output, codeOutput)
	if err != nil {
		return fmt.Errorf("error building diff for created accounts: %v", err)
	}
	return nil
}

// createdAndUpdatedState returns
// a mapping of their leafkeys to all the accounts that exist in a different state at B than A
// and a slice of the paths for all of the nodes included in both
func (sdb *builder) createdAndUpdatedState(iters iterPair, watchedAddresses []common.Address) (AccountMap, map[string]bool, error) {
	diffPathsAtB := make(map[string]bool)
	diffAcountsAtB := make(AccountMap)
	it, _ := trie.NewDifferenceIterator(iters.older, iters.newer)
	for it.Next(true) {
		if it.Leaf() || bytes.Equal(nullHashBytes, it.Hash().Bytes()) {
			continue
		}
		node, nodeElements, err := resolveNode(it, sdb.stateCache.TrieDB())
		if err != nil {
			return nil, nil, err
		}
		switch node.NodeType {
		case sdtypes.Leaf:
			// created vs updated is important for leaf nodes since we need to diff their storage
			// so we need to map all changed accounts at B to their leafkey, since account can change pathes but not leafkey
			var account state.Account
			if err := rlp.DecodeBytes(nodeElements[1].([]byte), &account); err != nil {
				return nil, nil, fmt.Errorf("error decoding account for leaf node at path %x nerror: %v", node.Path, err)
			}
			partialPath := trie.CompactToHex(nodeElements[0].([]byte))
			valueNodePath := append(node.Path, partialPath...)
			encodedPath := trie.HexToCompact(valueNodePath)
			leafKey := encodedPath[1:]
			if isWatchedAddress(watchedAddresses, leafKey) {
				diffAcountsAtB[common.Bytes2Hex(leafKey)] = accountWrapper{
					NodeType:  node.NodeType,
					Path:      node.Path,
					NodeValue: node.NodeValue,
					LeafKey:   leafKey,
					Account:   &account,
				}
			}
		}
		// add both intermediate and leaf node paths to the list of diffPathsAtB
		diffPathsAtB[common.Bytes2Hex(node.Path)] = true
	}
	return diffAcountsAtB, diffPathsAtB, it.Error()
}

// createdAndUpdatedStateWithIntermediateNodes returns
// a slice of all the intermediate nodes that exist in a different state at B than A
// a mapping of their leafkeys to all the accounts that exist in a different state at B than A
// and a slice of the paths for all of the nodes included in both
func (sdb *builder) createdAndUpdatedStateWithIntermediateNodes(iters iterPair, output sdtypes.StateNodeSink) (AccountMap, map[string]bool, error) {
	diffPathsAtB := make(map[string]bool)
	diffAcountsAtB := make(AccountMap)
	it, _ := trie.NewDifferenceIterator(iters.older, iters.newer)
	for it.Next(true) {
		if it.Leaf() || bytes.Equal(nullHashBytes, it.Hash().Bytes()) {
			continue
		}
		node, nodeElements, err := resolveNode(it, sdb.stateCache.TrieDB())
		if err != nil {
			return nil, nil, err
		}
		switch node.NodeType {
		case sdtypes.Leaf:
			// created vs updated is important for leaf nodes since we need to diff their storage
			// so we need to map all changed accounts at B to their leafkey, since account can change paths but not leafkey
			var account state.Account
			if err := rlp.DecodeBytes(nodeElements[1].([]byte), &account); err != nil {
				return nil, nil, fmt.Errorf("error decoding account for leaf node at path %x nerror: %v", node.Path, err)
			}
			partialPath := trie.CompactToHex(nodeElements[0].([]byte))
			valueNodePath := append(node.Path, partialPath...)
			encodedPath := trie.HexToCompact(valueNodePath)
			leafKey := encodedPath[1:]
			diffAcountsAtB[common.Bytes2Hex(leafKey)] = accountWrapper{
				NodeType:  node.NodeType,
				Path:      node.Path,
				NodeValue: node.NodeValue,
				LeafKey:   leafKey,
				Account:   &account,
			}
		case sdtypes.Extension, sdtypes.Branch:
			// create a diff for any intermediate node that has changed at b
			// created vs updated makes no difference for intermediate nodes since we do not need to diff storage
			if err := output(sdtypes.StateNode{
				NodeType:  node.NodeType,
				Path:      node.Path,
				NodeValue: node.NodeValue,
			}); err != nil {
				return nil, nil, err
			}
		default:
			return nil, nil, fmt.Errorf("unexpected node type %s", node.NodeType)
		}
		// add both intermediate and leaf node paths to the list of diffPathsAtB
		diffPathsAtB[common.Bytes2Hex(node.Path)] = true
	}
	return diffAcountsAtB, diffPathsAtB, it.Error()
}

// deletedOrUpdatedState returns a slice of all the paths that are emptied at B
// and a mapping of their leafkeys to all the accounts that exist in a different state at A than B
func (sdb *builder) deletedOrUpdatedState(iters iterPair, diffPathsAtB map[string]bool, output sdtypes.StateNodeSink) (AccountMap, error) {
	diffAccountAtA := make(AccountMap)
	it, _ := trie.NewDifferenceIterator(iters.newer, iters.older)
	for it.Next(true) {
		if it.Leaf() || bytes.Equal(nullHashBytes, it.Hash().Bytes()) {
			continue
		}
		node, nodeElements, err := resolveNode(it, sdb.stateCache.TrieDB())
		if err != nil {
			return nil, err
		}
		// if this node's path did not show up in diffPathsAtB
		// that means the node at this path was deleted (or moved) in B
		// emit an empty "removed" diff to signify as such
		if _, ok := diffPathsAtB[common.Bytes2Hex(node.Path)]; !ok {
			if err := output(sdtypes.StateNode{
				Path:      node.Path,
				NodeValue: []byte{},
				NodeType:  sdtypes.Removed,
			}); err != nil {
				return nil, err
			}
		}
		switch node.NodeType {
		case sdtypes.Leaf:
			// map all different accounts at A to their leafkey
			var account state.Account
			if err := rlp.DecodeBytes(nodeElements[1].([]byte), &account); err != nil {
				return nil, fmt.Errorf("error decoding account for leaf node at path %x nerror: %v", node.Path, err)
			}
			partialPath := trie.CompactToHex(nodeElements[0].([]byte))
			valueNodePath := append(node.Path, partialPath...)
			encodedPath := trie.HexToCompact(valueNodePath)
			leafKey := encodedPath[1:]
			diffAccountAtA[common.Bytes2Hex(leafKey)] = accountWrapper{
				NodeType:  node.NodeType,
				Path:      node.Path,
				NodeValue: node.NodeValue,
				LeafKey:   leafKey,
				Account:   &account,
			}
		case sdtypes.Extension, sdtypes.Branch:
			// fall through, we did everything we need to do with these node types
		default:
			return nil, fmt.Errorf("unexpected node type %s", node.NodeType)
		}
	}
	return diffAccountAtA, it.Error()
}

// buildAccountUpdates uses the account diffs maps for A => B and B => A and the known intersection of their leafkeys
// to generate the statediff node objects for all of the accounts that existed at both A and B but in different states
// needs to be called before building account creations and deletions as this mutates
// those account maps to remove the accounts which were updated
func (sdb *builder) buildAccountUpdates(creations, deletions AccountMap, updatedKeys []string, watchedStorageKeys []common.Hash, intermediateStorageNodes bool, output sdtypes.StateNodeSink) error {
	var err error
	for _, key := range updatedKeys {
		createdAcc := creations[key]
		deletedAcc := deletions[key]
		var storageDiffs []sdtypes.StorageNode
		if deletedAcc.Account != nil && createdAcc.Account != nil {
			oldSR := deletedAcc.Account.Root
			newSR := createdAcc.Account.Root
			err = sdb.buildStorageNodesIncremental(oldSR, newSR, watchedStorageKeys, intermediateStorageNodes, storageNodeAppender(&storageDiffs))
			if err != nil {
				return fmt.Errorf("failed building incremental storage diffs for account with leafkey %s\r\nerror: %v", key, err)
			}
		}
		if err = output(sdtypes.StateNode{
			NodeType:     createdAcc.NodeType,
			Path:         createdAcc.Path,
			NodeValue:    createdAcc.NodeValue,
			LeafKey:      createdAcc.LeafKey,
			StorageNodes: storageDiffs,
		}); err != nil {
			return err
		}
		delete(creations, key)
		delete(deletions, key)
	}

	return nil
}

// buildAccountCreations returns the statediff node objects for all the accounts that exist at B but not at A
// it also returns the code and codehash for created contract accounts
func (sdb *builder) buildAccountCreations(accounts AccountMap, watchedStorageKeys []common.Hash, intermediateStorageNodes bool, output sdtypes.StateNodeSink, codeOutput sdtypes.CodeSink) error {
	for _, val := range accounts {
		diff := sdtypes.StateNode{
			NodeType:  val.NodeType,
			Path:      val.Path,
			LeafKey:   val.LeafKey,
			NodeValue: val.NodeValue,
		}
		if !bytes.Equal(val.Account.CodeHash, nullCodeHash) {
			// For contract creations, any storage node contained is a diff
			var storageDiffs []sdtypes.StorageNode
			err := sdb.buildStorageNodesEventual(val.Account.Root, watchedStorageKeys, intermediateStorageNodes, storageNodeAppender(&storageDiffs))
			if err != nil {
				return fmt.Errorf("failed building eventual storage diffs for node %x\r\nerror: %v", val.Path, err)
			}
			diff.StorageNodes = storageDiffs
			// emit codehash => code mappings for code
			codeHash := common.BytesToHash(val.Account.CodeHash)
			code, err := sdb.stateCache.ContractCode(common.Hash{}, codeHash)
			if err != nil {
				return fmt.Errorf("failed to retrieve code for codehash %s\r\n error: %v", codeHash.String(), err)
			}
			if err := codeOutput(sdtypes.CodeAndCodeHash{
				Hash: codeHash,
				Code: code,
			}); err != nil {
				return err
			}
		}
		if err := output(diff); err != nil {
			return err
		}
	}

	return nil
}

// buildStorageNodesEventual builds the storage diff node objects for a created account
// i.e. it returns all the storage nodes at this state, since there is no previous state
func (sdb *builder) buildStorageNodesEventual(sr common.Hash, watchedStorageKeys []common.Hash, intermediateNodes bool, output sdtypes.StorageNodeSink) error {
	if bytes.Equal(sr.Bytes(), emptyContractRoot.Bytes()) {
		return nil
	}
	log.Debug("Storage Root For Eventual Diff", "root", sr.Hex())
	sTrie, err := sdb.stateCache.OpenTrie(sr)
	if err != nil {
		log.Info("error in build storage diff eventual", "error", err)
		return err
	}
	it := sTrie.NodeIterator(make([]byte, 0))
	err = sdb.buildStorageNodesFromTrie(it, watchedStorageKeys, intermediateNodes, output)
	if err != nil {
		return err
	}
	return nil
}

// buildStorageNodesFromTrie returns all the storage diff node objects in the provided node iterator
// if any storage keys are provided it will only return those leaf nodes
// including intermediate nodes can be turned on or off
func (sdb *builder) buildStorageNodesFromTrie(it trie.NodeIterator, watchedStorageKeys []common.Hash, intermediateNodes bool, output sdtypes.StorageNodeSink) error {
	for it.Next(true) {
		if it.Leaf() || bytes.Equal(nullHashBytes, it.Hash().Bytes()) {
			continue
		}
		node, nodeElements, err := resolveNode(it, sdb.stateCache.TrieDB())
		if err != nil {
			return err
		}
		switch node.NodeType {
		case sdtypes.Leaf:
			partialPath := trie.CompactToHex(nodeElements[0].([]byte))
			valueNodePath := append(node.Path, partialPath...)
			encodedPath := trie.HexToCompact(valueNodePath)
			leafKey := encodedPath[1:]
			if isWatchedStorageKey(watchedStorageKeys, leafKey) {
				if err := output(sdtypes.StorageNode{
					NodeType:  node.NodeType,
					Path:      node.Path,
					NodeValue: node.NodeValue,
					LeafKey:   leafKey,
				}); err != nil {
					return err
				}
			}
		case sdtypes.Extension, sdtypes.Branch:
			if intermediateNodes {
				if err := output(sdtypes.StorageNode{
					NodeType:  node.NodeType,
					Path:      node.Path,
					NodeValue: node.NodeValue,
				}); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unexpected node type %s", node.NodeType)
		}
	}
	return it.Error()
}

// buildStorageNodesIncremental builds the storage diff node objects for all nodes that exist in a different state at B than A
func (sdb *builder) buildStorageNodesIncremental(oldSR common.Hash, newSR common.Hash, watchedStorageKeys []common.Hash, intermediateNodes bool, output sdtypes.StorageNodeSink) error {
	if bytes.Equal(newSR.Bytes(), oldSR.Bytes()) {
		return nil
	}
	log.Debug("Storage Roots for Incremental Diff", "old", oldSR.Hex(), "new", newSR.Hex())
	oldTrie, err := sdb.stateCache.OpenTrie(oldSR)
	if err != nil {
		return err
	}
	newTrie, err := sdb.stateCache.OpenTrie(newSR)
	if err != nil {
		return err
	}

	diffPathsAtB, err := sdb.createdAndUpdatedStorage(oldTrie.NodeIterator([]byte{}), newTrie.NodeIterator([]byte{}), watchedStorageKeys, intermediateNodes, output)
	if err != nil {
		return err
	}
	err = sdb.deletedOrUpdatedStorage(oldTrie.NodeIterator([]byte{}), newTrie.NodeIterator([]byte{}), diffPathsAtB, watchedStorageKeys, intermediateNodes, output)
	if err != nil {
		return err
	}
	return nil
}

func (sdb *builder) createdAndUpdatedStorage(a, b trie.NodeIterator, watchedKeys []common.Hash, intermediateNodes bool, output sdtypes.StorageNodeSink) (map[string]bool, error) {
	diffPathsAtB := make(map[string]bool)
	it, _ := trie.NewDifferenceIterator(a, b)
	for it.Next(true) {
		if it.Leaf() || bytes.Equal(nullHashBytes, it.Hash().Bytes()) {
			continue
		}
		node, nodeElements, err := resolveNode(it, sdb.stateCache.TrieDB())
		if err != nil {
			return nil, err
		}
		switch node.NodeType {
		case sdtypes.Leaf:
			partialPath := trie.CompactToHex(nodeElements[0].([]byte))
			valueNodePath := append(node.Path, partialPath...)
			encodedPath := trie.HexToCompact(valueNodePath)
			leafKey := encodedPath[1:]
			if isWatchedStorageKey(watchedKeys, leafKey) {
				if err := output(sdtypes.StorageNode{
					NodeType:  node.NodeType,
					Path:      node.Path,
					NodeValue: node.NodeValue,
					LeafKey:   leafKey,
				}); err != nil {
					return nil, err
				}
			}
		case sdtypes.Extension, sdtypes.Branch:
			if intermediateNodes {
				if err := output(sdtypes.StorageNode{
					NodeType:  node.NodeType,
					Path:      node.Path,
					NodeValue: node.NodeValue,
				}); err != nil {
					return nil, err
				}
			}
		default:
			return nil, fmt.Errorf("unexpected node type %s", node.NodeType)
		}
		diffPathsAtB[common.Bytes2Hex(node.Path)] = true
	}
	return diffPathsAtB, it.Error()
}

func (sdb *builder) deletedOrUpdatedStorage(a, b trie.NodeIterator, diffPathsAtB map[string]bool, watchedKeys []common.Hash, intermediateNodes bool, output sdtypes.StorageNodeSink) error {
	it, _ := trie.NewDifferenceIterator(b, a)
	for it.Next(true) {
		if it.Leaf() || bytes.Equal(nullHashBytes, it.Hash().Bytes()) {
			continue
		}
		node, nodeElements, err := resolveNode(it, sdb.stateCache.TrieDB())
		if err != nil {
			return err
		}
		// if this node path showed up in diffPathsAtB
		// that means this node was updated at B and we already have the updated diff for it
		// otherwise that means this node was deleted in B and we need to add a "removed" diff to represent that event
		if _, ok := diffPathsAtB[common.Bytes2Hex(node.Path)]; ok {
			continue
		}
		switch node.NodeType {
		case sdtypes.Leaf:
			partialPath := trie.CompactToHex(nodeElements[0].([]byte))
			valueNodePath := append(node.Path, partialPath...)
			encodedPath := trie.HexToCompact(valueNodePath)
			leafKey := encodedPath[1:]
			if isWatchedStorageKey(watchedKeys, leafKey) {
				if err := output(sdtypes.StorageNode{
					NodeType:  sdtypes.Removed,
					Path:      node.Path,
					NodeValue: []byte{},
				}); err != nil {
					return err
				}
			}
		case sdtypes.Extension, sdtypes.Branch:
			if intermediateNodes {
				if err := output(sdtypes.StorageNode{
					NodeType:  sdtypes.Removed,
					Path:      node.Path,
					NodeValue: []byte{},
				}); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unexpected node type %s", node.NodeType)
		}
	}
	return it.Error()
}

// isWatchedAddress is used to check if a state account corresponds to one of the addresses the builder is configured to watch
func isWatchedAddress(watchedAddresses []common.Address, stateLeafKey []byte) bool {
	// If we aren't watching any specific addresses, we are watching everything
	if len(watchedAddresses) == 0 {
		return true
	}
	for _, addr := range watchedAddresses {
		addrHashKey := crypto.Keccak256(addr.Bytes())
		if bytes.Equal(addrHashKey, stateLeafKey) {
			return true
		}
	}
	return false
}

// isWatchedStorageKey is used to check if a storage leaf corresponds to one of the storage slots the builder is configured to watch
func isWatchedStorageKey(watchedKeys []common.Hash, storageLeafKey []byte) bool {
	// If we aren't watching any specific addresses, we are watching everything
	if len(watchedKeys) == 0 {
		return true
	}
	for _, hashKey := range watchedKeys {
		if bytes.Equal(hashKey.Bytes(), storageLeafKey) {
			return true
		}
	}
	return false
}
