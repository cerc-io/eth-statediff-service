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

package statediff_test

import (
	"bytes"
	"encoding/json"
	"math/big"
	"sort"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	sd "github.com/ethereum/go-ethereum/statediff"
	"github.com/ethereum/go-ethereum/statediff/indexer/ipld"
	"github.com/ethereum/go-ethereum/statediff/indexer/shared"
	"github.com/ethereum/go-ethereum/statediff/test_helpers"
	sdtypes "github.com/ethereum/go-ethereum/statediff/types"

	statediff "github.com/cerc-io/eth-statediff-service/pkg"
)

var (
	contractLeafKey                                        []byte
	emptyDiffs                                             = make([]sdtypes.StateLeafNode, 0)
	emptyStorage                                           = make([]sdtypes.StorageLeafNode, 0)
	block0, block1, block2, block3, block4, block5, block6 *types.Block
	builder                                                sd.Builder
	minerAddress                                           = common.HexToAddress("0x0")
	minerLeafKey                                           = test_helpers.AddressToLeafKey(minerAddress)
	workerCounts                                           = []uint{0, 1, 2, 4, 8}

	slot0 = common.BigToHash(big.NewInt(0))
	slot1 = common.BigToHash(big.NewInt(1))
	slot2 = common.BigToHash(big.NewInt(2))
	slot3 = common.BigToHash(big.NewInt(3))

	slot0StorageKey = crypto.Keccak256Hash(slot0[:])
	slot1StorageKey = crypto.Keccak256Hash(slot1[:])
	slot2StorageKey = crypto.Keccak256Hash(slot2[:])
	slot3StorageKey = crypto.Keccak256Hash(slot3[:])

	slot0StorageValue = common.Hex2Bytes("94703c4b2bd70c169f5717101caee543299fc946c7") // prefixed AccountAddr1
	slot1StorageValue = common.Hex2Bytes("01")
	slot2StorageValue = common.Hex2Bytes("09")
	slot3StorageValue = common.Hex2Bytes("03")

	slot0StorageLeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("390decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563"),
		slot0StorageValue,
	})
	slot1StorageLeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("310e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6"),
		slot1StorageValue,
	})
	slot2StorageLeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("305787fa12a823e0f2b7631cc41b3ba8828b3321ca811111fa75cd3aa3bb5ace"),
		slot2StorageValue,
	})
	slot3StorageLeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("32575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b"),
		slot3StorageValue,
	})
	contractAccountAtBlock2 = &types.StateAccount{
		Nonce:    1,
		Balance:  big.NewInt(0),
		CodeHash: common.HexToHash("0xaaea5efba4fd7b45d7ec03918ac5d8b31aa93b48986af0e6b591f0f087c80127").Bytes(),
		Root:     crypto.Keccak256Hash(block2StorageBranchRootNode),
	}
	contractAccountAtBlock2RLP, _      = rlp.EncodeToBytes(contractAccountAtBlock2)
	contractAccountAtBlock2LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("3114658a74d9cc9f7acf2c5cd696c3494d7c344d78bfec3add0d91ec4e8d1c45"),
		contractAccountAtBlock2RLP,
	})
	contractAccountAtBlock3 = &types.StateAccount{
		Nonce:    1,
		Balance:  big.NewInt(0),
		CodeHash: common.HexToHash("0xaaea5efba4fd7b45d7ec03918ac5d8b31aa93b48986af0e6b591f0f087c80127").Bytes(),
		Root:     crypto.Keccak256Hash(block3StorageBranchRootNode),
	}
	contractAccountAtBlock3RLP, _      = rlp.EncodeToBytes(contractAccountAtBlock3)
	contractAccountAtBlock3LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("3114658a74d9cc9f7acf2c5cd696c3494d7c344d78bfec3add0d91ec4e8d1c45"),
		contractAccountAtBlock3RLP,
	})
	contractAccountAtBlock4 = &types.StateAccount{
		Nonce:    1,
		Balance:  big.NewInt(0),
		CodeHash: common.HexToHash("0xaaea5efba4fd7b45d7ec03918ac5d8b31aa93b48986af0e6b591f0f087c80127").Bytes(),
		Root:     crypto.Keccak256Hash(block4StorageBranchRootNode),
	}
	contractAccountAtBlock4RLP, _      = rlp.EncodeToBytes(contractAccountAtBlock4)
	contractAccountAtBlock4LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("3114658a74d9cc9f7acf2c5cd696c3494d7c344d78bfec3add0d91ec4e8d1c45"),
		contractAccountAtBlock4RLP,
	})
	contractAccountAtBlock5 = &types.StateAccount{
		Nonce:    1,
		Balance:  big.NewInt(0),
		CodeHash: common.HexToHash("0xaaea5efba4fd7b45d7ec03918ac5d8b31aa93b48986af0e6b591f0f087c80127").Bytes(),
		Root:     crypto.Keccak256Hash(block5StorageBranchRootNode),
	}
	contractAccountAtBlock5RLP, _      = rlp.EncodeToBytes(contractAccountAtBlock5)
	contractAccountAtBlock5LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("3114658a74d9cc9f7acf2c5cd696c3494d7c344d78bfec3add0d91ec4e8d1c45"),
		contractAccountAtBlock5RLP,
	})
	minerAccountAtBlock1 = &types.StateAccount{
		Nonce:    0,
		Balance:  big.NewInt(2000002625000000000),
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	minerAccountAtBlock1RLP, _      = rlp.EncodeToBytes(minerAccountAtBlock1)
	minerAccountAtBlock1LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("3380c7b7ae81a58eb98d9c78de4a1fd7fd9535fc953ed2be602daaa41767312a"),
		minerAccountAtBlock1RLP,
	})
	minerAccountAtBlock2 = &types.StateAccount{
		Nonce:    0,
		Balance:  big.NewInt(4000111203461610525),
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	minerAccountAtBlock2RLP, _      = rlp.EncodeToBytes(minerAccountAtBlock2)
	minerAccountAtBlock2LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("3380c7b7ae81a58eb98d9c78de4a1fd7fd9535fc953ed2be602daaa41767312a"),
		minerAccountAtBlock2RLP,
	})

	account1AtBlock1 = &types.StateAccount{
		Nonce:    0,
		Balance:  test_helpers.Block1Account1Balance,
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	account1AtBlock1RLP, _      = rlp.EncodeToBytes(account1AtBlock1)
	account1AtBlock1LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("3926db69aaced518e9b9f0f434a473e7174109c943548bb8f23be41ca76d9ad2"),
		account1AtBlock1RLP,
	})
	account1AtBlock2 = &types.StateAccount{
		Nonce:    2,
		Balance:  big.NewInt(999555797000009000),
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	account1AtBlock2RLP, _      = rlp.EncodeToBytes(account1AtBlock2)
	account1AtBlock2LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("3926db69aaced518e9b9f0f434a473e7174109c943548bb8f23be41ca76d9ad2"),
		account1AtBlock2RLP,
	})
	account1AtBlock5 = &types.StateAccount{
		Nonce:    2,
		Balance:  big.NewInt(2999586469962854280),
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	account1AtBlock5RLP, _      = rlp.EncodeToBytes(account1AtBlock5)
	account1AtBlock5LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("3926db69aaced518e9b9f0f434a473e7174109c943548bb8f23be41ca76d9ad2"),
		account1AtBlock5RLP,
	})
	account1AtBlock6 = &types.StateAccount{
		Nonce:    3,
		Balance:  big.NewInt(2999557977962854280),
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	account1AtBlock6RLP, _      = rlp.EncodeToBytes(account1AtBlock6)
	account1AtBlock6LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("3926db69aaced518e9b9f0f434a473e7174109c943548bb8f23be41ca76d9ad2"),
		account1AtBlock6RLP,
	})
	account2AtBlock2 = &types.StateAccount{
		Nonce:    0,
		Balance:  big.NewInt(1000),
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	account2AtBlock2RLP, _      = rlp.EncodeToBytes(account2AtBlock2)
	account2AtBlock2LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("3957f3e2f04a0764c3a0491b175f69926da61efbcc8f61fa1455fd2d2b4cdd45"),
		account2AtBlock2RLP,
	})
	account2AtBlock3 = &types.StateAccount{
		Nonce:    0,
		Balance:  big.NewInt(2000013574009435976),
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	account2AtBlock3RLP, _      = rlp.EncodeToBytes(account2AtBlock3)
	account2AtBlock3LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("3957f3e2f04a0764c3a0491b175f69926da61efbcc8f61fa1455fd2d2b4cdd45"),
		account2AtBlock3RLP,
	})
	account2AtBlock4 = &types.StateAccount{
		Nonce:    0,
		Balance:  big.NewInt(4000048088163070348),
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	account2AtBlock4RLP, _      = rlp.EncodeToBytes(account2AtBlock4)
	account2AtBlock4LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("3957f3e2f04a0764c3a0491b175f69926da61efbcc8f61fa1455fd2d2b4cdd45"),
		account2AtBlock4RLP,
	})
	account2AtBlock6 = &types.StateAccount{
		Nonce:    0,
		Balance:  big.NewInt(6000063258066544204),
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	account2AtBlock6RLP, _      = rlp.EncodeToBytes(account2AtBlock6)
	account2AtBlock6LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("3957f3e2f04a0764c3a0491b175f69926da61efbcc8f61fa1455fd2d2b4cdd45"),
		account2AtBlock6RLP,
	})
	bankAccountAtBlock0 = &types.StateAccount{
		Nonce:    0,
		Balance:  big.NewInt(test_helpers.TestBankFunds.Int64()),
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	bankAccountAtBlock0RLP, _      = rlp.EncodeToBytes(bankAccountAtBlock0)
	bankAccountAtBlock0LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("2000bf49f440a1cd0527e4d06e2765654c0f56452257516d793a9b8d604dcfdf2a"),
		bankAccountAtBlock0RLP,
	})

	block1BankBalance   = big.NewInt(test_helpers.TestBankFunds.Int64() - test_helpers.BalanceChange10000 - test_helpers.GasFees)
	bankAccountAtBlock1 = &types.StateAccount{
		Nonce:    1,
		Balance:  block1BankBalance,
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	bankAccountAtBlock1RLP, _      = rlp.EncodeToBytes(bankAccountAtBlock1)
	bankAccountAtBlock1LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("30bf49f440a1cd0527e4d06e2765654c0f56452257516d793a9b8d604dcfdf2a"),
		bankAccountAtBlock1RLP,
	})

	block2BankBalance   = block1BankBalance.Int64() - test_helpers.BalanceChange1Ether - test_helpers.GasFees
	bankAccountAtBlock2 = &types.StateAccount{
		Nonce:    2,
		Balance:  big.NewInt(block2BankBalance),
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	bankAccountAtBlock2RLP, _      = rlp.EncodeToBytes(bankAccountAtBlock2)
	bankAccountAtBlock2LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("30bf49f440a1cd0527e4d06e2765654c0f56452257516d793a9b8d604dcfdf2a"),
		bankAccountAtBlock2RLP,
	})
	bankAccountAtBlock3 = &types.StateAccount{
		Nonce:    3,
		Balance:  big.NewInt(999914255999990000),
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	bankAccountAtBlock3RLP, _      = rlp.EncodeToBytes(bankAccountAtBlock3)
	bankAccountAtBlock3LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("30bf49f440a1cd0527e4d06e2765654c0f56452257516d793a9b8d604dcfdf2a"),
		bankAccountAtBlock3RLP,
	})
	bankAccountAtBlock4 = &types.StateAccount{
		Nonce:    6,
		Balance:  big.NewInt(999826859999990000),
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	bankAccountAtBlock4RLP, _      = rlp.EncodeToBytes(&bankAccountAtBlock4)
	bankAccountAtBlock4LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("30bf49f440a1cd0527e4d06e2765654c0f56452257516d793a9b8d604dcfdf2a"),
		bankAccountAtBlock4RLP,
	})
	bankAccountAtBlock5 = &types.StateAccount{
		Nonce:    8,
		Balance:  big.NewInt(999761283999990000),
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	bankAccountAtBlock5RLP, _      = rlp.EncodeToBytes(bankAccountAtBlock5)
	bankAccountAtBlock5LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("30bf49f440a1cd0527e4d06e2765654c0f56452257516d793a9b8d604dcfdf2a"),
		bankAccountAtBlock5RLP,
	})

	block1BranchRootNode, _ = rlp.EncodeToBytes(&[]interface{}{
		crypto.Keccak256(bankAccountAtBlock1LeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(minerAccountAtBlock1LeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(account1AtBlock1LeafNode),
		[]byte{},
		[]byte{},
	})
	block2BranchRootNode, _ = rlp.EncodeToBytes(&[]interface{}{
		crypto.Keccak256(bankAccountAtBlock2LeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(minerAccountAtBlock2LeafNode),
		crypto.Keccak256(contractAccountAtBlock2LeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(account2AtBlock2LeafNode),
		[]byte{},
		crypto.Keccak256(account1AtBlock2LeafNode),
		[]byte{},
		[]byte{},
	})
	block3BranchRootNode, _ = rlp.EncodeToBytes(&[]interface{}{
		crypto.Keccak256(bankAccountAtBlock3LeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(minerAccountAtBlock2LeafNode),
		crypto.Keccak256(contractAccountAtBlock3LeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(account2AtBlock3LeafNode),
		[]byte{},
		crypto.Keccak256(account1AtBlock2LeafNode),
		[]byte{},
		[]byte{},
	})
	block4BranchRootNode, _ = rlp.EncodeToBytes(&[]interface{}{
		crypto.Keccak256(bankAccountAtBlock4LeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(minerAccountAtBlock2LeafNode),
		crypto.Keccak256(contractAccountAtBlock4LeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(account2AtBlock4LeafNode),
		[]byte{},
		crypto.Keccak256(account1AtBlock2LeafNode),
		[]byte{},
		[]byte{},
	})
	block5BranchRootNode, _ = rlp.EncodeToBytes(&[]interface{}{
		crypto.Keccak256(bankAccountAtBlock5LeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(minerAccountAtBlock2LeafNode),
		crypto.Keccak256(contractAccountAtBlock5LeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(account2AtBlock4LeafNode),
		[]byte{},
		crypto.Keccak256(account1AtBlock5LeafNode),
		[]byte{},
		[]byte{},
	})
	block6BranchRootNode, _ = rlp.EncodeToBytes(&[]interface{}{
		crypto.Keccak256(bankAccountAtBlock5LeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(minerAccountAtBlock2LeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(account2AtBlock6LeafNode),
		[]byte{},
		crypto.Keccak256(account1AtBlock6LeafNode),
		[]byte{},
		[]byte{},
	})

	block2StorageBranchRootNode, _ = rlp.EncodeToBytes(&[]interface{}{
		[]byte{},
		[]byte{},
		crypto.Keccak256(slot0StorageLeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(slot1StorageLeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
	})
	block3StorageBranchRootNode, _ = rlp.EncodeToBytes(&[]interface{}{
		[]byte{},
		[]byte{},
		crypto.Keccak256(slot0StorageLeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(slot1StorageLeafNode),
		crypto.Keccak256(slot3StorageLeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
	})
	block4StorageBranchRootNode, _ = rlp.EncodeToBytes(&[]interface{}{
		[]byte{},
		[]byte{},
		crypto.Keccak256(slot0StorageLeafNode),
		[]byte{},
		crypto.Keccak256(slot2StorageLeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
	})
	block5StorageBranchRootNode, _ = rlp.EncodeToBytes(&[]interface{}{
		[]byte{},
		[]byte{},
		crypto.Keccak256(slot0StorageLeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(slot3StorageLeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
	})
)

func TestBuilder(t *testing.T) {
	blocks, chain := test_helpers.MakeChain(3, test_helpers.Genesis, test_helpers.TestChainGen)
	contractLeafKey = test_helpers.AddressToLeafKey(test_helpers.ContractAddr)
	defer chain.Stop()
	block0 = test_helpers.Genesis
	block1 = blocks[0]
	block2 = blocks[1]
	block3 = blocks[2]
	params := sd.Params{}

	var tests = []struct {
		name              string
		startingArguments sd.Args
		expected          *sdtypes.StateObject
	}{
		{
			"testEmptyDiff",
			sd.Args{
				OldStateRoot: block0.Root(),
				NewStateRoot: block0.Root(),
				BlockNumber:  block0.Number(),
				BlockHash:    block0.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block0.Number(),
				BlockHash:   block0.Hash(),
				Nodes:       emptyDiffs,
			},
		},
		{
			"testBlock0",
			//10000 transferred from testBankAddress to account1Addr
			sd.Args{
				OldStateRoot: test_helpers.NullHash,
				NewStateRoot: block0.Root(),
				BlockNumber:  block0.Number(),
				BlockHash:    block0.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block0.Number(),
				BlockHash:   block0.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: bankAccountAtBlock0,
							LeafKey: test_helpers.BankLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock0LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock0LeafNode)).String(),
						Content: bankAccountAtBlock0LeafNode,
					},
				},
			},
		},
		{
			"testBlock1",
			//10000 transferred from testBankAddress to account1Addr
			sd.Args{
				OldStateRoot: block0.Root(),
				NewStateRoot: block1.Root(),
				BlockNumber:  block1.Number(),
				BlockHash:    block1.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block1.Number(),
				BlockHash:   block1.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: bankAccountAtBlock1,
							LeafKey: test_helpers.BankLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock1LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: minerAccountAtBlock1,
							LeafKey: minerLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(minerAccountAtBlock1LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: account1AtBlock1,
							LeafKey: test_helpers.Account1LeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock1LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block1BranchRootNode)).String(),
						Content: block1BranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock1LeafNode)).String(),
						Content: bankAccountAtBlock1LeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(minerAccountAtBlock1LeafNode)).String(),
						Content: minerAccountAtBlock1LeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock1LeafNode)).String(),
						Content: account1AtBlock1LeafNode,
					},
				},
			},
		},
		{
			"testBlock2",
			// 1000 transferred from testBankAddress to account1Addr
			// 1000 transferred from account1Addr to account2Addr
			// account1addr creates a new contract
			sd.Args{
				OldStateRoot: block1.Root(),
				NewStateRoot: block2.Root(),
				BlockNumber:  block2.Number(),
				BlockHash:    block2.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block2.Number(),
				BlockHash:   block2.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: bankAccountAtBlock2,
							LeafKey: test_helpers.BankLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock2LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: minerAccountAtBlock2,
							LeafKey: minerLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(minerAccountAtBlock2LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: account1AtBlock2,
							LeafKey: test_helpers.Account1LeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock2LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: contractAccountAtBlock2,
							LeafKey: contractLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock2LeafNode)).String()},
						StorageDiff: []sdtypes.StorageLeafNode{
							{
								Removed: false,
								Value:   slot0StorageValue,
								LeafKey: slot0StorageKey.Bytes(),
								CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot0StorageLeafNode)).String(),
							},
							{
								Removed: false,
								Value:   slot1StorageValue,
								LeafKey: slot1StorageKey.Bytes(),
								CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot1StorageLeafNode)).String(),
							},
						},
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: account2AtBlock2,
							LeafKey: test_helpers.Account2LeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account2AtBlock2LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.RawBinary, test_helpers.CodeHash.Bytes()).String(),
						Content: test_helpers.ByteCodeAfterDeployment,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block2BranchRootNode)).String(),
						Content: block2BranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock2LeafNode)).String(),
						Content: bankAccountAtBlock2LeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(minerAccountAtBlock2LeafNode)).String(),
						Content: minerAccountAtBlock2LeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock2LeafNode)).String(),
						Content: account1AtBlock2LeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock2LeafNode)).String(),
						Content: contractAccountAtBlock2LeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(block2StorageBranchRootNode)).String(),
						Content: block2StorageBranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot0StorageLeafNode)).String(),
						Content: slot0StorageLeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot1StorageLeafNode)).String(),
						Content: slot1StorageLeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account2AtBlock2LeafNode)).String(),
						Content: account2AtBlock2LeafNode,
					},
				},
			},
		},
		{
			"testBlock3",
			//the contract's storage is changed
			//and the block is mined by account 2
			sd.Args{
				OldStateRoot: block2.Root(),
				NewStateRoot: block3.Root(),
				BlockNumber:  block3.Number(),
				BlockHash:    block3.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block3.Number(),
				BlockHash:   block3.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: bankAccountAtBlock3,
							LeafKey: test_helpers.BankLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock3LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: contractAccountAtBlock3,
							LeafKey: contractLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock3LeafNode)).String()},
						StorageDiff: []sdtypes.StorageLeafNode{
							{
								Removed: false,
								Value:   slot3StorageValue,
								LeafKey: slot3StorageKey.Bytes(),
								CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot3StorageLeafNode)).String(),
							},
						},
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: account2AtBlock3,
							LeafKey: test_helpers.Account2LeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account2AtBlock3LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block3BranchRootNode)).String(),
						Content: block3BranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock3LeafNode)).String(),
						Content: bankAccountAtBlock3LeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock3LeafNode)).String(),
						Content: contractAccountAtBlock3LeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(block3StorageBranchRootNode)).String(),
						Content: block3StorageBranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot3StorageLeafNode)).String(),
						Content: slot3StorageLeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account2AtBlock3LeafNode)).String(),
						Content: account2AtBlock3LeafNode,
					},
				},
			},
		},
	}

	for _, workers := range workerCounts {
		builder, _ = statediff.NewBuilder(chain.StateCache(), workers)
		for _, test := range tests {
			diff, err := builder.BuildStateDiffObject(test.startingArguments, params)
			if err != nil {
				t.Error(err)
			}
			receivedStateDiffRlp, err := rlp.EncodeToBytes(&diff)
			if err != nil {
				t.Error(err)
			}
			expectedStateDiffRlp, err := rlp.EncodeToBytes(test.expected)
			if err != nil {
				t.Error(err)
			}
			sort.Slice(receivedStateDiffRlp, func(i, j int) bool { return receivedStateDiffRlp[i] < receivedStateDiffRlp[j] })
			sort.Slice(expectedStateDiffRlp, func(i, j int) bool { return expectedStateDiffRlp[i] < expectedStateDiffRlp[j] })
			if !bytes.Equal(receivedStateDiffRlp, expectedStateDiffRlp) {
				actual, err := json.Marshal(diff)
				if err != nil {
					t.Error(err)
				}
				expected, err := json.Marshal(test.expected)
				if err != nil {
					t.Error(err)
				}
				t.Logf("Test failed: %s", test.name)
				t.Errorf("actual state diff: %s\r\n\r\n\r\nexpected state diff: %s", actual, expected)
			}
		}
	}

	// Let's also confirm that our root state nodes form the state root hash in the headers
	if !bytes.Equal(block0.Root().Bytes(), crypto.Keccak256(bankAccountAtBlock0LeafNode)) {
		t.Errorf("block0 expected root %x does not match actual root %x", block0.Root().Bytes(), crypto.Keccak256(bankAccountAtBlock0LeafNode))
	}
	if !bytes.Equal(block1.Root().Bytes(), crypto.Keccak256(block1BranchRootNode)) {
		t.Errorf("block1 expected root %x does not match actual root %x", block1.Root().Bytes(), crypto.Keccak256(block1BranchRootNode))
	}
	if !bytes.Equal(block2.Root().Bytes(), crypto.Keccak256(block2BranchRootNode)) {
		t.Errorf("block2 expected root %x does not match actual root %x", block2.Root().Bytes(), crypto.Keccak256(block2BranchRootNode))
	}
	if !bytes.Equal(block3.Root().Bytes(), crypto.Keccak256(block3BranchRootNode)) {
		t.Errorf("block3 expected root %x does not match actual root %x", block3.Root().Bytes(), crypto.Keccak256(block3BranchRootNode))
	}
}

func TestBuilderWithWatchedAddressList(t *testing.T) {
	blocks, chain := test_helpers.MakeChain(3, test_helpers.Genesis, test_helpers.TestChainGen)
	contractLeafKey = test_helpers.AddressToLeafKey(test_helpers.ContractAddr)
	defer chain.Stop()
	block0 = test_helpers.Genesis
	block1 = blocks[0]
	block2 = blocks[1]
	block3 = blocks[2]
	params := sd.Params{
		WatchedAddresses: []common.Address{test_helpers.Account1Addr, test_helpers.ContractAddr},
	}
	params.ComputeWatchedAddressesLeafPaths()

	var tests = []struct {
		name              string
		startingArguments sd.Args
		expected          *sdtypes.StateObject
	}{
		{
			"testEmptyDiff",
			sd.Args{
				OldStateRoot: block0.Root(),
				NewStateRoot: block0.Root(),
				BlockNumber:  block0.Number(),
				BlockHash:    block0.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block0.Number(),
				BlockHash:   block0.Hash(),
				Nodes:       emptyDiffs,
			},
		},
		{
			"testBlock0",
			//10000 transferred from testBankAddress to account1Addr
			sd.Args{
				OldStateRoot: test_helpers.NullHash,
				NewStateRoot: block0.Root(),
				BlockNumber:  block0.Number(),
				BlockHash:    block0.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block0.Number(),
				BlockHash:   block0.Hash(),
				Nodes:       emptyDiffs,
			},
		},
		{
			"testBlock1",
			//10000 transferred from testBankAddress to account1Addr
			sd.Args{
				OldStateRoot: block0.Root(),
				NewStateRoot: block1.Root(),
				BlockNumber:  block1.Number(),
				BlockHash:    block1.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block1.Number(),
				BlockHash:   block1.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: account1AtBlock1,
							LeafKey: test_helpers.Account1LeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock1LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block1BranchRootNode)).String(),
						Content: block1BranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock1LeafNode)).String(),
						Content: account1AtBlock1LeafNode,
					},
				},
			},
		},
		{
			"testBlock2",
			//1000 transferred from testBankAddress to account1Addr
			//1000 transferred from account1Addr to account2Addr
			sd.Args{
				OldStateRoot: block1.Root(),
				NewStateRoot: block2.Root(),
				BlockNumber:  block2.Number(),
				BlockHash:    block2.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block2.Number(),
				BlockHash:   block2.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: contractAccountAtBlock2,
							LeafKey: contractLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock2LeafNode)).String()},
						StorageDiff: []sdtypes.StorageLeafNode{
							{
								Removed: false,
								Value:   slot0StorageValue,
								LeafKey: slot0StorageKey.Bytes(),
								CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot0StorageLeafNode)).String(),
							},
							{
								Removed: false,
								Value:   slot1StorageValue,
								LeafKey: slot1StorageKey.Bytes(),
								CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot1StorageLeafNode)).String(),
							},
						},
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: account1AtBlock2,
							LeafKey: test_helpers.Account1LeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock2LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.RawBinary, test_helpers.CodeHash.Bytes()).String(),
						Content: test_helpers.ByteCodeAfterDeployment,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block2BranchRootNode)).String(),
						Content: block2BranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock2LeafNode)).String(),
						Content: contractAccountAtBlock2LeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(block2StorageBranchRootNode)).String(),
						Content: block2StorageBranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot0StorageLeafNode)).String(),
						Content: slot0StorageLeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot1StorageLeafNode)).String(),
						Content: slot1StorageLeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock2LeafNode)).String(),
						Content: account1AtBlock2LeafNode,
					},
				},
			},
		},
		{
			"testBlock3",
			//the contract's storage is changed
			//and the block is mined by account 2
			sd.Args{
				OldStateRoot: block2.Root(),
				NewStateRoot: block3.Root(),
				BlockNumber:  block3.Number(),
				BlockHash:    block3.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block3.Number(),
				BlockHash:   block3.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: contractAccountAtBlock3,
							LeafKey: contractLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock3LeafNode)).String()},
						StorageDiff: []sdtypes.StorageLeafNode{
							{
								Removed: false,
								Value:   slot3StorageValue,
								LeafKey: slot3StorageKey.Bytes(),
								CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot3StorageLeafNode)).String(),
							},
						},
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block3BranchRootNode)).String(),
						Content: block3BranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock3LeafNode)).String(),
						Content: contractAccountAtBlock3LeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(block3StorageBranchRootNode)).String(),
						Content: block3StorageBranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot3StorageLeafNode)).String(),
						Content: slot3StorageLeafNode,
					},
				},
			},
		},
	}

	for _, workers := range workerCounts {
		builder, _ = statediff.NewBuilder(chain.StateCache(), workers)
		for _, test := range tests {
			diff, err := builder.BuildStateDiffObject(test.startingArguments, params)
			if err != nil {
				t.Error(err)
			}
			receivedStateDiffRlp, err := rlp.EncodeToBytes(&diff)
			if err != nil {
				t.Error(err)
			}
			expectedStateDiffRlp, err := rlp.EncodeToBytes(test.expected)
			if err != nil {
				t.Error(err)
			}
			sort.Slice(receivedStateDiffRlp, func(i, j int) bool { return receivedStateDiffRlp[i] < receivedStateDiffRlp[j] })
			sort.Slice(expectedStateDiffRlp, func(i, j int) bool { return expectedStateDiffRlp[i] < expectedStateDiffRlp[j] })
			if !bytes.Equal(receivedStateDiffRlp, expectedStateDiffRlp) {
				actual, err := json.Marshal(diff)
				if err != nil {
					t.Error(err)
				}
				expected, err := json.Marshal(test.expected)
				if err != nil {
					t.Error(err)
				}
				t.Logf("Test failed: %s", test.name)
				t.Errorf("actual state diff: %s\r\n\r\n\r\nexpected state diff: %s", actual, expected)
			}
		}
	}

	// Let's also confirm that our root state nodes form the state root hash in the headers
	if !bytes.Equal(block0.Root().Bytes(), crypto.Keccak256(bankAccountAtBlock0LeafNode)) {
		t.Errorf("block0 expected root %x does not match actual root %x", block0.Root().Bytes(), crypto.Keccak256(bankAccountAtBlock0LeafNode))
	}
	if !bytes.Equal(block1.Root().Bytes(), crypto.Keccak256(block1BranchRootNode)) {
		t.Errorf("block1 expected root %x does not match actual root %x", block1.Root().Bytes(), crypto.Keccak256(block1BranchRootNode))
	}
	if !bytes.Equal(block2.Root().Bytes(), crypto.Keccak256(block2BranchRootNode)) {
		t.Errorf("block2 expected root %x does not match actual root %x", block2.Root().Bytes(), crypto.Keccak256(block2BranchRootNode))
	}
	if !bytes.Equal(block3.Root().Bytes(), crypto.Keccak256(block3BranchRootNode)) {
		t.Errorf("block3 expected root %x does not match actual root %x", block3.Root().Bytes(), crypto.Keccak256(block3BranchRootNode))
	}
}

func TestBuilderWithRemovedAccountAndStorage(t *testing.T) {
	blocks, chain := test_helpers.MakeChain(6, test_helpers.Genesis, test_helpers.TestChainGen)
	contractLeafKey = test_helpers.AddressToLeafKey(test_helpers.ContractAddr)
	defer chain.Stop()
	block3 = blocks[2]
	block4 = blocks[3]
	block5 = blocks[4]
	block6 = blocks[5]
	params := sd.Params{}

	var tests = []struct {
		name              string
		startingArguments sd.Args
		expected          *sdtypes.StateObject
	}{
		// blocks 0-3 are the same as in TestBuilderWithIntermediateNodes
		{
			"testBlock4",
			sd.Args{
				OldStateRoot: block3.Root(),
				NewStateRoot: block4.Root(),
				BlockNumber:  block4.Number(),
				BlockHash:    block4.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block4.Number(),
				BlockHash:   block4.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: bankAccountAtBlock4,
							LeafKey: test_helpers.BankLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock4LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: contractAccountAtBlock4,
							LeafKey: contractLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock4LeafNode)).String()},
						StorageDiff: []sdtypes.StorageLeafNode{
							{
								Removed: false,
								Value:   slot2StorageValue,
								LeafKey: slot2StorageKey.Bytes(),
								CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot2StorageLeafNode)).String(),
							},
							{
								Removed: true,
								LeafKey: slot1StorageKey.Bytes(),
								CID:     shared.RemovedNodeStorageCID,
								Value:   []byte{},
							},
							{
								Removed: true,
								LeafKey: slot3StorageKey.Bytes(),
								CID:     shared.RemovedNodeStorageCID,
								Value:   []byte{},
							},
						},
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: account2AtBlock4,
							LeafKey: test_helpers.Account2LeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account2AtBlock4LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block4BranchRootNode)).String(),
						Content: block4BranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock4LeafNode)).String(),
						Content: bankAccountAtBlock4LeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock4LeafNode)).String(),
						Content: contractAccountAtBlock4LeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(block4StorageBranchRootNode)).String(),
						Content: block4StorageBranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot2StorageLeafNode)).String(),
						Content: slot2StorageLeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account2AtBlock4LeafNode)).String(),
						Content: account2AtBlock4LeafNode,
					},
				},
			},
		},
		{
			"testBlock5",
			sd.Args{
				OldStateRoot: block4.Root(),
				NewStateRoot: block5.Root(),
				BlockNumber:  block5.Number(),
				BlockHash:    block5.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block5.Number(),
				BlockHash:   block5.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: bankAccountAtBlock5,
							LeafKey: test_helpers.BankLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock5LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: contractAccountAtBlock5,
							LeafKey: contractLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock5LeafNode)).String()},
						StorageDiff: []sdtypes.StorageLeafNode{
							{
								Removed: false,
								Value:   slot3StorageValue,
								LeafKey: slot3StorageKey.Bytes(),
								CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot3StorageLeafNode)).String(),
							},
							{
								Removed: true,
								LeafKey: slot2StorageKey.Bytes(),
								CID:     shared.RemovedNodeStorageCID,
								Value:   []byte{},
							},
						},
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: account1AtBlock5,
							LeafKey: test_helpers.Account1LeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock5LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block5BranchRootNode)).String(),
						Content: block5BranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock5LeafNode)).String(),
						Content: bankAccountAtBlock5LeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock5LeafNode)).String(),
						Content: contractAccountAtBlock5LeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(block5StorageBranchRootNode)).String(),
						Content: block5StorageBranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot3StorageLeafNode)).String(),
						Content: slot3StorageLeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock5LeafNode)).String(),
						Content: account1AtBlock5LeafNode,
					},
				},
			},
		},
		{
			"testBlock6",
			sd.Args{
				OldStateRoot: block5.Root(),
				NewStateRoot: block6.Root(),
				BlockNumber:  block6.Number(),
				BlockHash:    block6.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block6.Number(),
				BlockHash:   block6.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: true,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: nil,
							LeafKey: contractLeafKey,
							CID:     shared.RemovedNodeStateCID},
						StorageDiff: []sdtypes.StorageLeafNode{
							{
								Removed: true,
								LeafKey: slot0StorageKey.Bytes(),
								CID:     shared.RemovedNodeStorageCID,
								Value:   []byte{},
							},
							{
								Removed: true,
								LeafKey: slot3StorageKey.Bytes(),
								CID:     shared.RemovedNodeStorageCID,
								Value:   []byte{},
							},
						},
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: account2AtBlock6,
							LeafKey: test_helpers.Account2LeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account2AtBlock6LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: account1AtBlock6,
							LeafKey: test_helpers.Account1LeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock6LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block6BranchRootNode)).String(),
						Content: block6BranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account2AtBlock6LeafNode)).String(),
						Content: account2AtBlock6LeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock6LeafNode)).String(),
						Content: account1AtBlock6LeafNode,
					},
				},
			},
		},
	}

	for _, workers := range workerCounts {
		builder, _ = statediff.NewBuilder(chain.StateCache(), workers)
		for _, test := range tests {
			diff, err := builder.BuildStateDiffObject(test.startingArguments, params)
			if err != nil {
				t.Error(err)
			}
			receivedStateDiffRlp, err := rlp.EncodeToBytes(&diff)
			if err != nil {
				t.Error(err)
			}
			expectedStateDiffRlp, err := rlp.EncodeToBytes(test.expected)
			if err != nil {
				t.Error(err)
			}
			sort.Slice(receivedStateDiffRlp, func(i, j int) bool { return receivedStateDiffRlp[i] < receivedStateDiffRlp[j] })
			sort.Slice(expectedStateDiffRlp, func(i, j int) bool { return expectedStateDiffRlp[i] < expectedStateDiffRlp[j] })
			if !bytes.Equal(receivedStateDiffRlp, expectedStateDiffRlp) {
				actual, err := json.Marshal(diff)
				if err != nil {
					t.Error(err)
				}
				expected, err := json.Marshal(test.expected)
				if err != nil {
					t.Error(err)
				}
				t.Logf("Test failed: %s", test.name)
				t.Errorf("actual state diff: %s\r\n\r\n\r\nexpected state diff: %s", actual, expected)
			}
		}
	}

	// Let's also confirm that our root state nodes form the state root hash in the headers
	if !bytes.Equal(block4.Root().Bytes(), crypto.Keccak256(block4BranchRootNode)) {
		t.Errorf("block4 expected root %x does not match actual root %x", block4.Root().Bytes(), crypto.Keccak256(block4BranchRootNode))
	}
	if !bytes.Equal(block5.Root().Bytes(), crypto.Keccak256(block5BranchRootNode)) {
		t.Errorf("block5 expected root %x does not match actual root %x", block5.Root().Bytes(), crypto.Keccak256(block5BranchRootNode))
	}
	if !bytes.Equal(block6.Root().Bytes(), crypto.Keccak256(block6BranchRootNode)) {
		t.Errorf("block6 expected root %x does not match actual root %x", block6.Root().Bytes(), crypto.Keccak256(block6BranchRootNode))
	}
}

func TestBuilderWithRemovedNonWatchedAccount(t *testing.T) {
	blocks, chain := test_helpers.MakeChain(6, test_helpers.Genesis, test_helpers.TestChainGen)
	contractLeafKey = test_helpers.AddressToLeafKey(test_helpers.ContractAddr)
	defer chain.Stop()
	block3 = blocks[2]
	block4 = blocks[3]
	block5 = blocks[4]
	block6 = blocks[5]
	params := sd.Params{
		WatchedAddresses: []common.Address{test_helpers.Account1Addr, test_helpers.Account2Addr},
	}
	params.ComputeWatchedAddressesLeafPaths()

	var tests = []struct {
		name              string
		startingArguments sd.Args
		expected          *sdtypes.StateObject
	}{
		{
			"testBlock4",
			sd.Args{
				OldStateRoot: block3.Root(),
				NewStateRoot: block4.Root(),
				BlockNumber:  block4.Number(),
				BlockHash:    block4.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block4.Number(),
				BlockHash:   block4.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: account2AtBlock4,
							LeafKey: test_helpers.Account2LeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account2AtBlock4LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block4BranchRootNode)).String(),
						Content: block4BranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account2AtBlock4LeafNode)).String(),
						Content: account2AtBlock4LeafNode,
					},
				},
			},
		},
		{
			"testBlock5",
			sd.Args{
				OldStateRoot: block4.Root(),
				NewStateRoot: block5.Root(),
				BlockNumber:  block5.Number(),
				BlockHash:    block5.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block5.Number(),
				BlockHash:   block5.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: account1AtBlock5,
							LeafKey: test_helpers.Account1LeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock5LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block5BranchRootNode)).String(),
						Content: block5BranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock5LeafNode)).String(),
						Content: account1AtBlock5LeafNode,
					},
				},
			},
		},
		{
			"testBlock6",
			sd.Args{
				OldStateRoot: block5.Root(),
				NewStateRoot: block6.Root(),
				BlockNumber:  block6.Number(),
				BlockHash:    block6.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block6.Number(),
				BlockHash:   block6.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: account2AtBlock6,
							LeafKey: test_helpers.Account2LeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account2AtBlock6LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: account1AtBlock6,
							LeafKey: test_helpers.Account1LeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock6LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block6BranchRootNode)).String(),
						Content: block6BranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account2AtBlock6LeafNode)).String(),
						Content: account2AtBlock6LeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock6LeafNode)).String(),
						Content: account1AtBlock6LeafNode,
					},
				},
			},
		},
	}

	for _, workers := range workerCounts {
		builder, _ = statediff.NewBuilder(chain.StateCache(), workers)
		for _, test := range tests {
			diff, err := builder.BuildStateDiffObject(test.startingArguments, params)
			if err != nil {
				t.Error(err)
			}
			receivedStateDiffRlp, err := rlp.EncodeToBytes(&diff)
			if err != nil {
				t.Error(err)
			}

			expectedStateDiffRlp, err := rlp.EncodeToBytes(test.expected)
			if err != nil {
				t.Error(err)
			}

			sort.Slice(receivedStateDiffRlp, func(i, j int) bool { return receivedStateDiffRlp[i] < receivedStateDiffRlp[j] })
			sort.Slice(expectedStateDiffRlp, func(i, j int) bool { return expectedStateDiffRlp[i] < expectedStateDiffRlp[j] })
			if !bytes.Equal(receivedStateDiffRlp, expectedStateDiffRlp) {
				actual, err := json.Marshal(diff)
				if err != nil {
					t.Error(err)
				}
				expected, err := json.Marshal(test.expected)
				if err != nil {
					t.Error(err)
				}
				t.Logf("Test failed: %s", test.name)
				t.Errorf("actual state diff: %s\r\n\r\n\r\nexpected state diff: %s", actual, expected)
			}
		}
	}

	// Let's also confirm that our root state nodes form the state root hash in the headers
	if !bytes.Equal(block4.Root().Bytes(), crypto.Keccak256(block4BranchRootNode)) {
		t.Errorf("block4 expected root %x does not match actual root %x", block4.Root().Bytes(), crypto.Keccak256(block4BranchRootNode))
	}
	if !bytes.Equal(block5.Root().Bytes(), crypto.Keccak256(block5BranchRootNode)) {
		t.Errorf("block5 expected root %x does not match actual root %x", block5.Root().Bytes(), crypto.Keccak256(block5BranchRootNode))
	}
	if !bytes.Equal(block6.Root().Bytes(), crypto.Keccak256(block6BranchRootNode)) {
		t.Errorf("block6 expected root %x does not match actual root %x", block6.Root().Bytes(), crypto.Keccak256(block6BranchRootNode))
	}
}

func TestBuilderWithRemovedWatchedAccount(t *testing.T) {
	blocks, chain := test_helpers.MakeChain(6, test_helpers.Genesis, test_helpers.TestChainGen)
	contractLeafKey = test_helpers.AddressToLeafKey(test_helpers.ContractAddr)
	defer chain.Stop()
	block3 = blocks[2]
	block4 = blocks[3]
	block5 = blocks[4]
	block6 = blocks[5]
	params := sd.Params{
		WatchedAddresses: []common.Address{test_helpers.Account1Addr, test_helpers.ContractAddr},
	}
	params.ComputeWatchedAddressesLeafPaths()

	var tests = []struct {
		name              string
		startingArguments sd.Args
		expected          *sdtypes.StateObject
	}{
		{
			"testBlock4",
			sd.Args{
				OldStateRoot: block3.Root(),
				NewStateRoot: block4.Root(),
				BlockNumber:  block4.Number(),
				BlockHash:    block4.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block4.Number(),
				BlockHash:   block4.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: contractAccountAtBlock4,
							LeafKey: contractLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock4LeafNode)).String()},
						StorageDiff: []sdtypes.StorageLeafNode{
							{
								Removed: false,
								LeafKey: slot2StorageKey.Bytes(),
								Value:   slot2StorageValue,
								CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot2StorageLeafNode)).String(),
							},
							{
								Removed: true,
								LeafKey: slot1StorageKey.Bytes(),
								CID:     shared.RemovedNodeStorageCID,
								Value:   []byte{},
							},
							{
								Removed: true,
								LeafKey: slot3StorageKey.Bytes(),
								CID:     shared.RemovedNodeStorageCID,
								Value:   []byte{},
							},
						},
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block4BranchRootNode)).String(),
						Content: block4BranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock4LeafNode)).String(),
						Content: contractAccountAtBlock4LeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(block4StorageBranchRootNode)).String(),
						Content: block4StorageBranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot2StorageLeafNode)).String(),
						Content: slot2StorageLeafNode,
					},
				},
			},
		},
		{
			"testBlock5",
			sd.Args{
				OldStateRoot: block4.Root(),
				NewStateRoot: block5.Root(),
				BlockNumber:  block5.Number(),
				BlockHash:    block5.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block5.Number(),
				BlockHash:   block5.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: contractAccountAtBlock5,
							LeafKey: contractLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock5LeafNode)).String()},
						StorageDiff: []sdtypes.StorageLeafNode{
							{
								Removed: false,
								LeafKey: slot3StorageKey.Bytes(),
								Value:   slot3StorageValue,
								CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot3StorageLeafNode)).String(),
							},
							{
								Removed: true,
								LeafKey: slot2StorageKey.Bytes(),
								CID:     shared.RemovedNodeStorageCID,
								Value:   []byte{},
							},
						},
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: account1AtBlock5,
							LeafKey: test_helpers.Account1LeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock5LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block5BranchRootNode)).String(),
						Content: block5BranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock5LeafNode)).String(),
						Content: contractAccountAtBlock5LeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(block5StorageBranchRootNode)).String(),
						Content: block5StorageBranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot3StorageLeafNode)).String(),
						Content: slot3StorageLeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock5LeafNode)).String(),
						Content: account1AtBlock5LeafNode,
					},
				},
			},
		},
		{
			"testBlock6",
			sd.Args{
				OldStateRoot: block5.Root(),
				NewStateRoot: block6.Root(),
				BlockNumber:  block6.Number(),
				BlockHash:    block6.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block6.Number(),
				BlockHash:   block6.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: true,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: nil,
							LeafKey: contractLeafKey,
							CID:     shared.RemovedNodeStateCID},
						StorageDiff: []sdtypes.StorageLeafNode{
							{
								Removed: true,
								LeafKey: slot0StorageKey.Bytes(),
								CID:     shared.RemovedNodeStorageCID,
								Value:   []byte{},
							},
							{
								Removed: true,
								LeafKey: slot3StorageKey.Bytes(),
								CID:     shared.RemovedNodeStorageCID,
								Value:   []byte{},
							},
						},
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: account1AtBlock6,
							LeafKey: test_helpers.Account1LeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock6LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block6BranchRootNode)).String(),
						Content: block6BranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock6LeafNode)).String(),
						Content: account1AtBlock6LeafNode,
					},
				},
			},
		},
	}

	for _, workers := range workerCounts {
		builder, _ = statediff.NewBuilder(chain.StateCache(), workers)
		for _, test := range tests {
			diff, err := builder.BuildStateDiffObject(test.startingArguments, params)
			if err != nil {
				t.Error(err)
			}
			receivedStateDiffRlp, err := rlp.EncodeToBytes(&diff)
			if err != nil {
				t.Error(err)
			}

			expectedStateDiffRlp, err := rlp.EncodeToBytes(test.expected)
			if err != nil {
				t.Error(err)
			}

			sort.Slice(receivedStateDiffRlp, func(i, j int) bool { return receivedStateDiffRlp[i] < receivedStateDiffRlp[j] })
			sort.Slice(expectedStateDiffRlp, func(i, j int) bool { return expectedStateDiffRlp[i] < expectedStateDiffRlp[j] })
			if !bytes.Equal(receivedStateDiffRlp, expectedStateDiffRlp) {
				actual, err := json.Marshal(diff)
				if err != nil {
					t.Error(err)
				}
				expected, err := json.Marshal(test.expected)
				if err != nil {
					t.Error(err)
				}
				t.Logf("Test failed: %s", test.name)
				t.Errorf("actual state diff: %s\r\n\r\n\r\nexpected state diff: %s", actual, expected)
			}
		}
	}

	// Let's also confirm that our root state nodes form the state root hash in the headers
	if !bytes.Equal(block4.Root().Bytes(), crypto.Keccak256(block4BranchRootNode)) {
		t.Errorf("block4 expected root %x does not match actual root %x", block4.Root().Bytes(), crypto.Keccak256(block4BranchRootNode))
	}
	if !bytes.Equal(block5.Root().Bytes(), crypto.Keccak256(block5BranchRootNode)) {
		t.Errorf("block5 expected root %x does not match actual root %x", block5.Root().Bytes(), crypto.Keccak256(block5BranchRootNode))
	}
	if !bytes.Equal(block6.Root().Bytes(), crypto.Keccak256(block6BranchRootNode)) {
		t.Errorf("block6 expected root %x does not match actual root %x", block6.Root().Bytes(), crypto.Keccak256(block6BranchRootNode))
	}
}

var (
	slot00StorageValue = common.Hex2Bytes("9471562b71999873db5b286df957af199ec94617f7") // prefixed TestBankAddress

	slot00StorageLeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("390decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563"),
		slot00StorageValue,
	})

	contractAccountAtBlock01 = &types.StateAccount{
		Nonce:    1,
		Balance:  big.NewInt(0),
		CodeHash: common.HexToHash("0xaaea5efba4fd7b45d7ec03918ac5d8b31aa93b48986af0e6b591f0f087c80127").Bytes(),
		Root:     crypto.Keccak256Hash(block01StorageBranchRootNode),
	}
	contractAccountAtBlock01RLP, _      = rlp.EncodeToBytes(contractAccountAtBlock01)
	contractAccountAtBlock01LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("3cb2583748c26e89ef19c2a8529b05a270f735553b4d44b6f2a1894987a71c8b"),
		contractAccountAtBlock01RLP,
	})

	bankAccountAtBlock01 = &types.StateAccount{
		Nonce:    1,
		Balance:  big.NewInt(3999629697375000000),
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	bankAccountAtBlock01RLP, _      = rlp.EncodeToBytes(bankAccountAtBlock01)
	bankAccountAtBlock01LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("30bf49f440a1cd0527e4d06e2765654c0f56452257516d793a9b8d604dcfdf2a"),
		bankAccountAtBlock01RLP,
	})
	bankAccountAtBlock02 = &types.StateAccount{
		Nonce:    2,
		Balance:  big.NewInt(5999607323457344852),
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	bankAccountAtBlock02RLP, _      = rlp.EncodeToBytes(bankAccountAtBlock02)
	bankAccountAtBlock02LeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("2000bf49f440a1cd0527e4d06e2765654c0f56452257516d793a9b8d604dcfdf2a"),
		bankAccountAtBlock02RLP,
	})

	block01BranchRootNode, _ = rlp.EncodeToBytes(&[]interface{}{
		crypto.Keccak256Hash(bankAccountAtBlock01LeafNode),
		crypto.Keccak256Hash(contractAccountAtBlock01LeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
	})

	block01StorageBranchRootNode, _ = rlp.EncodeToBytes(&[]interface{}{
		[]byte{},
		[]byte{},
		crypto.Keccak256(slot00StorageLeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(slot1StorageLeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
	})
)

func TestBuilderWithMovedAccount(t *testing.T) {
	blocks, chain := test_helpers.MakeChain(2, test_helpers.Genesis, test_helpers.TestSelfDestructChainGen)
	contractLeafKey = test_helpers.AddressToLeafKey(test_helpers.ContractAddr)
	defer chain.Stop()
	block0 = test_helpers.Genesis
	block1 = blocks[0]
	block2 = blocks[1]
	params := sd.Params{}

	var tests = []struct {
		name              string
		startingArguments sd.Args
		expected          *sdtypes.StateObject
	}{
		{
			"testBlock1",
			sd.Args{
				OldStateRoot: block0.Root(),
				NewStateRoot: block1.Root(),
				BlockNumber:  block1.Number(),
				BlockHash:    block1.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block1.Number(),
				BlockHash:   block1.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: bankAccountAtBlock01,
							LeafKey: test_helpers.BankLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock01LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: contractAccountAtBlock01,
							LeafKey: contractLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock01LeafNode)).String()},
						StorageDiff: []sdtypes.StorageLeafNode{
							{
								Removed: false,
								LeafKey: slot0StorageKey.Bytes(),
								Value:   slot00StorageValue,
								CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot00StorageLeafNode)).String(),
							},
							{
								Removed: false,
								LeafKey: slot1StorageKey.Bytes(),
								Value:   slot1StorageValue,
								CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot1StorageLeafNode)).String(),
							},
						},
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.RawBinary, test_helpers.CodeHash.Bytes()).String(),
						Content: test_helpers.ByteCodeAfterDeployment,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block01BranchRootNode)).String(),
						Content: block01BranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock01LeafNode)).String(),
						Content: bankAccountAtBlock01LeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock01LeafNode)).String(),
						Content: contractAccountAtBlock01LeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(block01StorageBranchRootNode)).String(),
						Content: block01StorageBranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot00StorageLeafNode)).String(),
						Content: slot00StorageLeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot1StorageLeafNode)).String(),
						Content: slot1StorageLeafNode,
					},
				},
			},
		},
		{
			"testBlock2",
			sd.Args{
				OldStateRoot: block1.Root(),
				NewStateRoot: block2.Root(),
				BlockNumber:  block2.Number(),
				BlockHash:    block2.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block2.Number(),
				BlockHash:   block2.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: bankAccountAtBlock02,
							LeafKey: test_helpers.BankLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock02LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
					{
						Removed: true,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: nil,
							LeafKey: contractLeafKey,
							CID:     shared.RemovedNodeStateCID},
						StorageDiff: []sdtypes.StorageLeafNode{
							{
								Removed: true,
								LeafKey: slot0StorageKey.Bytes(),
								CID:     shared.RemovedNodeStorageCID,
								Value:   []byte{},
							},
							{
								Removed: true,
								LeafKey: slot1StorageKey.Bytes(),
								CID:     shared.RemovedNodeStorageCID,
								Value:   []byte{},
							},
						},
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock02LeafNode)).String(),
						Content: bankAccountAtBlock02LeafNode,
					},
				},
			},
		},
	}

	for _, workers := range workerCounts {
		builder, _ = statediff.NewBuilder(chain.StateCache(), workers)
		for _, test := range tests {
			diff, err := builder.BuildStateDiffObject(test.startingArguments, params)
			if err != nil {
				t.Error(err)
			}
			receivedStateDiffRlp, err := rlp.EncodeToBytes(&diff)
			if err != nil {
				t.Error(err)
			}
			expectedStateDiffRlp, err := rlp.EncodeToBytes(test.expected)
			if err != nil {
				t.Error(err)
			}

			sort.Slice(receivedStateDiffRlp, func(i, j int) bool { return receivedStateDiffRlp[i] < receivedStateDiffRlp[j] })
			sort.Slice(expectedStateDiffRlp, func(i, j int) bool { return expectedStateDiffRlp[i] < expectedStateDiffRlp[j] })
			if !bytes.Equal(receivedStateDiffRlp, expectedStateDiffRlp) {
				actual, err := json.Marshal(diff)
				if err != nil {
					t.Error(err)
				}
				expected, err := json.Marshal(test.expected)
				if err != nil {
					t.Error(err)
				}
				t.Logf("Test failed: %s", test.name)
				t.Errorf("actual state diff: %s\r\n\r\n\r\nexpected state diff: %s", actual, expected)
			}
		}
	}

	// Let's also confirm that our root state nodes form the state root hash in the headers
	if !bytes.Equal(block1.Root().Bytes(), crypto.Keccak256(block01BranchRootNode)) {
		t.Errorf("block01 expected root %x does not match actual root %x", block1.Root().Bytes(), crypto.Keccak256(block01BranchRootNode))
	}
	if !bytes.Equal(block2.Root().Bytes(), crypto.Keccak256(bankAccountAtBlock02LeafNode)) {
		t.Errorf("block02 expected root %x does not match actual root %x", block2.Root().Bytes(), crypto.Keccak256(bankAccountAtBlock02LeafNode))
	}
}

/*
pragma solidity ^0.5.10;

contract test {
    address payable owner;

    modifier onlyOwner {
        require(
            msg.sender == owner,
            "Only owner can call this function."
        );
        _;
    }

    uint256[100] data;

	constructor() public {
	    owner = msg.sender;
		data = [1];
	}

    function Put(uint256 addr, uint256 value) public {
        data[addr] = value;
    }

    function close() public onlyOwner { //onlyOwner is custom modifier
        selfdestruct(owner);  // `owner` is the owners address
    }
}
*/

var (
	b                    = big.NewInt(0).Sub(test_helpers.TestBIGBankFunds, test_helpers.BalanceChangeBIG)
	block1BankBigBalance = big.NewInt(0).Sub(b, big.NewInt(test_helpers.GasFees2))
	bankAccountAtBlock1b = &types.StateAccount{
		Nonce:    1,
		Balance:  block1BankBigBalance,
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	bankAccountAtBlock1bRLP, _      = rlp.EncodeToBytes(bankAccountAtBlock1b)
	bankAccountAtBlock1bLeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("30bf49f440a1cd0527e4d06e2765654c0f56452257516d793a9b8d604dcfdf2a"),
		bankAccountAtBlock1bRLP,
	})

	account1AtBlock1b = &types.StateAccount{
		Nonce:    0,
		Balance:  test_helpers.Block1bAccount1Balance,
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	account1AtBlock1bRLP, _      = rlp.EncodeToBytes(account1AtBlock1b)
	account1AtBlock1bLeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("3926db69aaced518e9b9f0f434a473e7174109c943548bb8f23be41ca76d9ad2"),
		account1AtBlock1bRLP,
	})

	account1AtBlock2bBalance, _ = big.NewInt(0).SetString("1999999999999999999999999761539571000000000", 10)
	account1AtBlock2b           = &types.StateAccount{
		Nonce:    1,
		Balance:  account1AtBlock2bBalance,
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	account1AtBlock2bRLP, _      = rlp.EncodeToBytes(account1AtBlock2b)
	account1AtBlock2bLeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("3926db69aaced518e9b9f0f434a473e7174109c943548bb8f23be41ca76d9ad2"),
		account1AtBlock2bRLP,
	})

	minerAccountAtBlock2b = &types.StateAccount{
		Nonce:    0,
		Balance:  big.NewInt(4055891787808414571),
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	minerAccountAtBlock2bRLP, _      = rlp.EncodeToBytes(minerAccountAtBlock2b)
	minerAccountAtBlock2bLeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("3380c7b7ae81a58eb98d9c78de4a1fd7fd9535fc953ed2be602daaa41767312a"),
		minerAccountAtBlock2bRLP,
	})

	contractAccountAtBlock2b = &types.StateAccount{
		Nonce:    1,
		Balance:  big.NewInt(0),
		CodeHash: test_helpers.CodeHashForInternalizedLeafNode.Bytes(),
		Root:     crypto.Keccak256Hash(block2StorageBranchRootNode),
	}
	contractAccountAtBlock2bRLP, _      = rlp.EncodeToBytes(contractAccountAtBlock2b)
	contractAccountAtBlock2bLeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("3d7e14f1723fa19b5d6d9f8b86b49acefbc9c400bf4ed686c10d6b6467fc5b3a"),
		contractAccountAtBlock2bRLP,
	})

	bankAccountAtBlock3bBalance, _ = big.NewInt(0).SetString("18000000000000000000000001999920365757724976", 10)
	bankAccountAtBlock3b           = &types.StateAccount{
		Nonce:    3,
		Balance:  bankAccountAtBlock3bBalance,
		CodeHash: test_helpers.NullCodeHash.Bytes(),
		Root:     test_helpers.EmptyContractRoot,
	}
	bankAccountAtBlock3bRLP, _      = rlp.EncodeToBytes(bankAccountAtBlock3b)
	bankAccountAtBlock3bLeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("30bf49f440a1cd0527e4d06e2765654c0f56452257516d793a9b8d604dcfdf2a"),
		bankAccountAtBlock3bRLP,
	})

	contractAccountAtBlock3b = &types.StateAccount{
		Nonce:    1,
		Balance:  big.NewInt(0),
		CodeHash: test_helpers.CodeHashForInternalizedLeafNode.Bytes(),
		Root:     crypto.Keccak256Hash(block3bStorageBranchRootNode),
	}
	contractAccountAtBlock3bRLP, _      = rlp.EncodeToBytes(contractAccountAtBlock3b)
	contractAccountAtBlock3bLeafNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("3d7e14f1723fa19b5d6d9f8b86b49acefbc9c400bf4ed686c10d6b6467fc5b3a"),
		contractAccountAtBlock3bRLP,
	})

	slot40364  = common.BigToHash(big.NewInt(40364))
	slot105566 = common.BigToHash(big.NewInt(105566))

	slot40364StorageValue  = common.Hex2Bytes("01")
	slot105566StorageValue = common.Hex2Bytes("02")

	slot40364StorageKey  = crypto.Keccak256Hash(slot40364[:])
	slot105566StorageKey = crypto.Keccak256Hash(slot105566[:])

	slot40364StorageInternalLeafNode = []interface{}{
		common.Hex2Bytes("3077bbc951a04529defc15da8c06e427cde0d7a1499c50975bbe8aab"),
		slot40364StorageValue,
	}
	slot105566StorageInternalLeafNode = []interface{}{
		common.Hex2Bytes("3c62586c18bf1ecfda161ced374b7a894630e2db426814c24e5d42af"),
		slot105566StorageValue,
	}

	block3bStorageBranchRootNode, _ = rlp.EncodeToBytes(&[]interface{}{
		[]byte{},
		[]byte{},
		crypto.Keccak256(slot0StorageLeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(slot1StorageLeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(block3bStorageExtensionNode),
		[]byte{},
	})

	block3bStorageExtensionNode, _ = rlp.EncodeToBytes(&[]interface{}{
		common.Hex2Bytes("1291631c"),
		crypto.Keccak256(block3bStorageBranchNodeWithInternalLeaves),
	})

	block3bStorageBranchNodeWithInternalLeaves, _ = rlp.EncodeToBytes(&[]interface{}{
		slot105566StorageInternalLeafNode,
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		slot40364StorageInternalLeafNode,
		[]byte{},
		[]byte{},
		[]byte{},
	})

	block1bBranchRootNode, _ = rlp.EncodeToBytes(&[]interface{}{
		crypto.Keccak256(bankAccountAtBlock1bLeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(minerAccountAtBlock1LeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(account1AtBlock1bLeafNode),
		[]byte{},
		[]byte{},
	})

	block2bBranchRootNode, _ = rlp.EncodeToBytes(&[]interface{}{
		crypto.Keccak256(bankAccountAtBlock1bLeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(minerAccountAtBlock2bLeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(account1AtBlock2bLeafNode),
		crypto.Keccak256(contractAccountAtBlock2bLeafNode),
		[]byte{},
	})

	block3bBranchRootNode, _ = rlp.EncodeToBytes(&[]interface{}{
		crypto.Keccak256(bankAccountAtBlock3bLeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(minerAccountAtBlock2bLeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(account1AtBlock2bLeafNode),
		crypto.Keccak256(contractAccountAtBlock3bLeafNode),
		[]byte{},
	})
)

func TestBuilderWithInternalizedLeafNode(t *testing.T) {
	blocks, chain := test_helpers.MakeChain(3, test_helpers.GenesisForInternalLeafNodeTest, test_helpers.TestChainGenWithInternalLeafNode)
	contractLeafKey = test_helpers.AddressToLeafKey(test_helpers.ContractAddr)
	defer chain.Stop()
	block0 = test_helpers.Genesis
	block1 = blocks[0]
	block2 = blocks[1]
	block3 = blocks[2]
	params := sd.Params{}

	var tests = []struct {
		name              string
		startingArguments sd.Args
		expected          *sdtypes.StateObject
	}{
		{
			"testEmptyDiff",
			sd.Args{
				OldStateRoot: block0.Root(),
				NewStateRoot: block0.Root(),
				BlockNumber:  block0.Number(),
				BlockHash:    block0.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block0.Number(),
				BlockHash:   block0.Hash(),
				Nodes:       emptyDiffs,
			},
		},
		{
			"testBlock0",
			//10000 transferred from testBankAddress to account1Addr
			sd.Args{
				OldStateRoot: test_helpers.NullHash,
				NewStateRoot: block0.Root(),
				BlockNumber:  block0.Number(),
				BlockHash:    block0.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block0.Number(),
				BlockHash:   block0.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: bankAccountAtBlock0,
							LeafKey: test_helpers.BankLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock0LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock0LeafNode)).String(),
						Content: bankAccountAtBlock0LeafNode,
					},
				},
			},
		},
		{
			"testBlock1",
			//10000 transferred from testBankAddress to account1Addr
			sd.Args{
				OldStateRoot: block0.Root(),
				NewStateRoot: block1.Root(),
				BlockNumber:  block1.Number(),
				BlockHash:    block1.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block1.Number(),
				BlockHash:   block1.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: bankAccountAtBlock1b,
							LeafKey: test_helpers.BankLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock1bLeafNode)).String()},
						StorageDiff: emptyStorage,
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: minerAccountAtBlock1,
							LeafKey: minerLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(minerAccountAtBlock1LeafNode)).String()},
						StorageDiff: emptyStorage,
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: account1AtBlock1b,
							LeafKey: test_helpers.Account1LeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock1bLeafNode)).String()},
						StorageDiff: emptyStorage,
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block1bBranchRootNode)).String(),
						Content: block1bBranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock1bLeafNode)).String(),
						Content: bankAccountAtBlock1bLeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(minerAccountAtBlock1LeafNode)).String(),
						Content: minerAccountAtBlock1LeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock1bLeafNode)).String(),
						Content: account1AtBlock1bLeafNode,
					},
				},
			},
		},
		{
			"testBlock2",
			// 1000 transferred from testBankAddress to account1Addr
			// 1000 transferred from account1Addr to account2Addr
			// account1addr creates a new contract
			sd.Args{
				OldStateRoot: block1.Root(),
				NewStateRoot: block2.Root(),
				BlockNumber:  block2.Number(),
				BlockHash:    block2.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block2.Number(),
				BlockHash:   block2.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: minerAccountAtBlock2b,
							LeafKey: minerLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(minerAccountAtBlock2bLeafNode)).String()},
						StorageDiff: emptyStorage,
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: account1AtBlock2b,
							LeafKey: test_helpers.Account1LeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock2bLeafNode)).String()},
						StorageDiff: emptyStorage,
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: contractAccountAtBlock2b,
							LeafKey: contractLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock2bLeafNode)).String()},
						StorageDiff: []sdtypes.StorageLeafNode{
							{
								Removed: false,
								Value:   slot0StorageValue,
								LeafKey: slot0StorageKey.Bytes(),
								CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot0StorageLeafNode)).String(),
							},
							{
								Removed: false,
								Value:   slot1StorageValue,
								LeafKey: slot1StorageKey.Bytes(),
								CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot1StorageLeafNode)).String(),
							},
						},
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.RawBinary, test_helpers.CodeHashForInternalizedLeafNode.Bytes()).String(),
						Content: test_helpers.ByteCodeAfterDeploymentForInternalLeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block2bBranchRootNode)).String(),
						Content: block2bBranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(minerAccountAtBlock2bLeafNode)).String(),
						Content: minerAccountAtBlock2bLeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(account1AtBlock2bLeafNode)).String(),
						Content: account1AtBlock2bLeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock2bLeafNode)).String(),
						Content: contractAccountAtBlock2bLeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(block2StorageBranchRootNode)).String(),
						Content: block2StorageBranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot0StorageLeafNode)).String(),
						Content: slot0StorageLeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot1StorageLeafNode)).String(),
						Content: slot1StorageLeafNode,
					},
				},
			},
		},
		{
			"testBlock3",
			//the contract's storage is changed
			//and the block is mined by account 2
			sd.Args{
				OldStateRoot: block2.Root(),
				NewStateRoot: block3.Root(),
				BlockNumber:  block3.Number(),
				BlockHash:    block3.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block3.Number(),
				BlockHash:   block3.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: bankAccountAtBlock3b,
							LeafKey: test_helpers.BankLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock3bLeafNode)).String()},
						StorageDiff: emptyStorage,
					},
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: contractAccountAtBlock3b,
							LeafKey: contractLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock3bLeafNode)).String()},
						StorageDiff: []sdtypes.StorageLeafNode{
							{
								Removed: false,
								Value:   slot105566StorageValue,
								LeafKey: slot105566StorageKey.Bytes(),
								CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(block3bStorageBranchNodeWithInternalLeaves)).String(),
							},
							{
								Removed: false,
								Value:   slot40364StorageValue,
								LeafKey: slot40364StorageKey.Bytes(),
								CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(block3bStorageBranchNodeWithInternalLeaves)).String(),
							},
						},
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block3bBranchRootNode)).String(),
						Content: block3bBranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(bankAccountAtBlock3bLeafNode)).String(),
						Content: bankAccountAtBlock3bLeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock3bLeafNode)).String(),
						Content: contractAccountAtBlock3bLeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(block3bStorageBranchRootNode)).String(),
						Content: block3bStorageBranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(block3bStorageExtensionNode)).String(),
						Content: block3bStorageExtensionNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(block3bStorageBranchNodeWithInternalLeaves)).String(),
						Content: block3bStorageBranchNodeWithInternalLeaves,
					},
				},
			},
		},
	}

	for _, workers := range workerCounts {
		builder, _ = statediff.NewBuilder(chain.StateCache(), workers)
		for _, test := range tests {
			diff, err := builder.BuildStateDiffObject(test.startingArguments, params)
			if err != nil {
				t.Error(err)
			}
			receivedStateDiffRlp, err := rlp.EncodeToBytes(&diff)
			if err != nil {
				t.Error(err)
			}
			expectedStateDiffRlp, err := rlp.EncodeToBytes(test.expected)
			if err != nil {
				t.Error(err)
			}
			sort.Slice(receivedStateDiffRlp, func(i, j int) bool { return receivedStateDiffRlp[i] < receivedStateDiffRlp[j] })
			sort.Slice(expectedStateDiffRlp, func(i, j int) bool { return expectedStateDiffRlp[i] < expectedStateDiffRlp[j] })
			if !bytes.Equal(receivedStateDiffRlp, expectedStateDiffRlp) {
				actual, err := json.Marshal(diff)
				if err != nil {
					t.Error(err)
				}
				expected, err := json.Marshal(test.expected)
				if err != nil {
					t.Error(err)
				}
				t.Logf("Test failed: %s", test.name)
				t.Errorf("actual state diff: %s\r\n\r\n\r\nexpected state diff: %s", actual, expected)
			}
		}
	}

	// Let's also confirm that our root state nodes form the state root hash in the headers
	if !bytes.Equal(block1.Root().Bytes(), crypto.Keccak256(block1bBranchRootNode)) {
		t.Errorf("block1 expected root %x does not match actual root %x", block1.Root().Bytes(), crypto.Keccak256(block1bBranchRootNode))
	}
	if !bytes.Equal(block2.Root().Bytes(), crypto.Keccak256(block2bBranchRootNode)) {
		t.Errorf("block2 expected root %x does not match actual root %x", block2.Root().Bytes(), crypto.Keccak256(block2bBranchRootNode))
	}
	if !bytes.Equal(block3.Root().Bytes(), crypto.Keccak256(block3bBranchRootNode)) {
		t.Errorf("block3 expected root %x does not match actual root %x", block3.Root().Bytes(), crypto.Keccak256(block3bBranchRootNode))
	}
}

func TestBuilderWithInternalizedLeafNodeAndWatchedAddress(t *testing.T) {
	blocks, chain := test_helpers.MakeChain(3, test_helpers.GenesisForInternalLeafNodeTest, test_helpers.TestChainGenWithInternalLeafNode)
	contractLeafKey = test_helpers.AddressToLeafKey(test_helpers.ContractAddr)
	defer chain.Stop()
	block0 = test_helpers.Genesis
	block1 = blocks[0]
	block2 = blocks[1]
	block3 = blocks[2]
	params := sd.Params{
		WatchedAddresses: []common.Address{
			test_helpers.ContractAddr,
		},
	}
	params.ComputeWatchedAddressesLeafPaths()

	var tests = []struct {
		name              string
		startingArguments sd.Args
		expected          *sdtypes.StateObject
	}{
		{
			"testEmptyDiff",
			sd.Args{
				OldStateRoot: block0.Root(),
				NewStateRoot: block0.Root(),
				BlockNumber:  block0.Number(),
				BlockHash:    block0.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block0.Number(),
				BlockHash:   block0.Hash(),
				Nodes:       emptyDiffs,
			},
		},
		{
			"testBlock0",
			//10000 transferred from testBankAddress to account1Addr
			sd.Args{
				OldStateRoot: test_helpers.NullHash,
				NewStateRoot: block0.Root(),
				BlockNumber:  block0.Number(),
				BlockHash:    block0.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block0.Number(),
				BlockHash:   block0.Hash(),
				Nodes:       []sdtypes.StateLeafNode{},
				IPLDs:       []sdtypes.IPLD{}, // there's some kind of weird behavior where if our root node is a leaf node
				// even though it is along the path to the watched leaf (necessarily, as it is the root) it doesn't get included
				// unconsequential, but kinda odd.
			},
		},
		{
			"testBlock1",
			//10000 transferred from testBankAddress to account1Addr
			sd.Args{
				OldStateRoot: block0.Root(),
				NewStateRoot: block1.Root(),
				BlockNumber:  block1.Number(),
				BlockHash:    block1.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block1.Number(),
				BlockHash:   block1.Hash(),
				Nodes:       []sdtypes.StateLeafNode{},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block1bBranchRootNode)).String(),
						Content: block1bBranchRootNode,
					},
				},
			},
		},
		{
			"testBlock2",
			// 1000 transferred from testBankAddress to account1Addr
			// 1000 transferred from account1Addr to account2Addr
			// account1addr creates a new contract
			sd.Args{
				OldStateRoot: block1.Root(),
				NewStateRoot: block2.Root(),
				BlockNumber:  block2.Number(),
				BlockHash:    block2.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block2.Number(),
				BlockHash:   block2.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: contractAccountAtBlock2b,
							LeafKey: contractLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock2bLeafNode)).String()},
						StorageDiff: []sdtypes.StorageLeafNode{
							{
								Removed: false,
								Value:   slot0StorageValue,
								LeafKey: slot0StorageKey.Bytes(),
								CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot0StorageLeafNode)).String(),
							},
							{
								Removed: false,
								Value:   slot1StorageValue,
								LeafKey: slot1StorageKey.Bytes(),
								CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot1StorageLeafNode)).String(),
							},
						},
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.RawBinary, test_helpers.CodeHashForInternalizedLeafNode.Bytes()).String(),
						Content: test_helpers.ByteCodeAfterDeploymentForInternalLeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block2bBranchRootNode)).String(),
						Content: block2bBranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock2bLeafNode)).String(),
						Content: contractAccountAtBlock2bLeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(block2StorageBranchRootNode)).String(),
						Content: block2StorageBranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot0StorageLeafNode)).String(),
						Content: slot0StorageLeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(slot1StorageLeafNode)).String(),
						Content: slot1StorageLeafNode,
					},
				},
			},
		},
		{
			"testBlock3",
			//the contract's storage is changed
			//and the block is mined by account 2
			sd.Args{
				OldStateRoot: block2.Root(),
				NewStateRoot: block3.Root(),
				BlockNumber:  block3.Number(),
				BlockHash:    block3.Hash(),
			},
			&sdtypes.StateObject{
				BlockNumber: block3.Number(),
				BlockHash:   block3.Hash(),
				Nodes: []sdtypes.StateLeafNode{
					{
						Removed: false,
						AccountWrapper: struct {
							Account *types.StateAccount
							LeafKey []byte
							CID     string
						}{
							Account: contractAccountAtBlock3b,
							LeafKey: contractLeafKey,
							CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock3bLeafNode)).String()},
						StorageDiff: []sdtypes.StorageLeafNode{
							{
								Removed: false,
								Value:   slot105566StorageValue,
								LeafKey: slot105566StorageKey.Bytes(),
								CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(block3bStorageBranchNodeWithInternalLeaves)).String(),
							},
							{
								Removed: false,
								Value:   slot40364StorageValue,
								LeafKey: slot40364StorageKey.Bytes(),
								CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(block3bStorageBranchNodeWithInternalLeaves)).String(),
							},
						},
					},
				},
				IPLDs: []sdtypes.IPLD{
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(block3bBranchRootNode)).String(),
						Content: block3bBranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStateTrie, crypto.Keccak256(contractAccountAtBlock3bLeafNode)).String(),
						Content: contractAccountAtBlock3bLeafNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(block3bStorageBranchRootNode)).String(),
						Content: block3bStorageBranchRootNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(block3bStorageExtensionNode)).String(),
						Content: block3bStorageExtensionNode,
					},
					{
						CID:     ipld.Keccak256ToCid(ipld.MEthStorageTrie, crypto.Keccak256(block3bStorageBranchNodeWithInternalLeaves)).String(),
						Content: block3bStorageBranchNodeWithInternalLeaves,
					},
				},
			},
		},
	}

	for _, workers := range workerCounts {
		builder, _ = statediff.NewBuilder(chain.StateCache(), workers)
		for _, test := range tests {
			diff, err := builder.BuildStateDiffObject(test.startingArguments, params)
			if err != nil {
				t.Error(err)
			}
			receivedStateDiffRlp, err := rlp.EncodeToBytes(&diff)
			if err != nil {
				t.Error(err)
			}
			expectedStateDiffRlp, err := rlp.EncodeToBytes(test.expected)
			if err != nil {
				t.Error(err)
			}
			sort.Slice(receivedStateDiffRlp, func(i, j int) bool { return receivedStateDiffRlp[i] < receivedStateDiffRlp[j] })
			sort.Slice(expectedStateDiffRlp, func(i, j int) bool { return expectedStateDiffRlp[i] < expectedStateDiffRlp[j] })
			if !bytes.Equal(receivedStateDiffRlp, expectedStateDiffRlp) {
				actual, err := json.Marshal(diff)
				if err != nil {
					t.Error(err)
				}
				expected, err := json.Marshal(test.expected)
				if err != nil {
					t.Error(err)
				}
				t.Logf("Test failed: %s", test.name)
				t.Errorf("actual state diff: %s\r\n\r\n\r\nexpected state diff: %s", actual, expected)
			}
		}
	}

	// Let's also confirm that our root state nodes form the state root hash in the headers
	if !bytes.Equal(block1.Root().Bytes(), crypto.Keccak256(block1bBranchRootNode)) {
		t.Errorf("block1 expected root %x does not match actual root %x", block1.Root().Bytes(), crypto.Keccak256(block1bBranchRootNode))
	}
	if !bytes.Equal(block2.Root().Bytes(), crypto.Keccak256(block2bBranchRootNode)) {
		t.Errorf("block2 expected root %x does not match actual root %x", block2.Root().Bytes(), crypto.Keccak256(block2bBranchRootNode))
	}
	if !bytes.Equal(block3.Root().Bytes(), crypto.Keccak256(block3bBranchRootNode)) {
		t.Errorf("block3 expected root %x does not match actual root %x", block3.Root().Bytes(), crypto.Keccak256(block3bBranchRootNode))
	}
}

/*
pragma solidity ^0.5.10;

contract test {
    address payable owner;

    modifier onlyOwner {
        require(
            msg.sender == owner,
            "Only owner can call this function."
        );
        _;
    }

    uint256[105566] data;

	constructor() public {
	    owner = msg.sender;
		data = [1];
	}

    function Put(uint256 addr, uint256 value) public {
        data[addr] = value;
    }

    function close() public onlyOwner { //onlyOwner is custom modifier
        selfdestruct(owner);  // `owner` is the owners address
    }
}
*/
