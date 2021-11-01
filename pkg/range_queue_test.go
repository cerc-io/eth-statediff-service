package statediff

import (
	"sync"
	"testing"

	sd "github.com/ethereum/go-ethereum/statediff"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	reqs     []RangeRequest
	expected interface{}
}

func newBlockRangeReq() blockRangeQueue {
	return blockRangeQueue{
		RWMutex:    sync.RWMutex{},
		rangeQueue: nil,
	}
}

func TestRangeQueuePush(t *testing.T) {
	brq := newBlockRangeReq()
	cases := &testCase{
		reqs: []RangeRequest{
			{1, 5, sd.Params{}},
			{4, 10, sd.Params{}},
		},
		expected: 2,
	}

	for _, req := range cases.reqs {
		brq.push(req)
	}

	require.Equal(t, cases.expected, len(brq.rangeQueue))
}

func TestRangeQueuePop(t *testing.T) {
	brq := newBlockRangeReq()

	cases := &testCase{
		reqs: []RangeRequest{
			{1, 5, sd.Params{}},
			{4, 10, sd.Params{}},
		},
		expected: RangeRequest{1, 5, sd.Params{}},
	}

	for _, req := range cases.reqs {
		brq.push(req)
	}

	pop, err := brq.pop()
	require.NoError(t, err)

	require.Equal(t, 1, len(brq.rangeQueue))
	require.Equal(t, cases.expected, pop)
}

func TestRangeQueueRemove(t *testing.T) {
	brq := newBlockRangeReq()

	cases := &testCase{
		reqs: []RangeRequest{
			{1, 5, sd.Params{}},
			{4, 10, sd.Params{}},
			{11, 20, sd.Params{}},
			{20, 30, sd.Params{}},
		},
		expected: []RangeRequest{
			{1, 5, sd.Params{}},
			{4, 10, sd.Params{}},
			{20, 30, sd.Params{}},
		},
	}

	for _, req := range cases.reqs {
		brq.push(req)
	}

	err := brq.remove(cases.reqs[2])
	require.NoError(t, err)

	require.Equal(t, 3, len(brq.rangeQueue))
	require.Equal(t, cases.expected, brq.rangeQueue)
}

func TestRangeQueueGet(t *testing.T) {
	brq := newBlockRangeReq()

	cases := &testCase{
		reqs: []RangeRequest{
			{1, 5, sd.Params{}},
			{4, 10, sd.Params{}},
		},
		expected: []RangeRequest{
			{1, 5, sd.Params{}},
			{4, 10, sd.Params{}},
		},
	}

	for _, req := range cases.reqs {
		brq.push(req)
	}

	rQueue := brq.get()
	require.Equal(t, cases.expected, rQueue)
}
