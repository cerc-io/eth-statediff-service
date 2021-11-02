package statediff

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/statediff"
	"github.com/stretchr/testify/require"
)

type testCases struct {
	name        string
	start, stop uint64
	expected    interface{}
}

func newTestService(rngReq []RangeRequest) *Service {
	return &Service{
		blockRngQueue: &blockRangeQueue{
			RWMutex:    sync.RWMutex{},
			rangeQueue: rngReq,
		},
	}
}

func TestWriteStateDiffsInRange(t *testing.T) {
	var rngReq []RangeRequest

	testSrvc := newTestService(rngReq)
	cases := []testCases{
		{
			name:     "invalid block range range",
			start:    5,
			stop:     1,
			expected: fmt.Sprintf("invalid block range (%d, %d): stop height must be greater or equal to start height", 5, 1),
		},
		{
			name:     "valid block range",
			start:    1,
			stop:     5,
			expected: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := testSrvc.WriteStateDiffsInRange(tc.start, tc.stop, statediff.Params{})
			if err != nil {
				require.Equal(t, tc.expected, err.Error())
				return
			}

			require.Equal(t, len(testSrvc.blockRngQueue.rangeQueue), 1)
		})
	}
}

func TestRemoveRange(t *testing.T) {
	rngReq := []RangeRequest{{
		Start:  1,
		Stop:   5,
		Params: statediff.Params{},
	}}

	testSrvc := newTestService(rngReq)
	cases := []testCases{
		{
			name:     "invalid block range range",
			start:    0,
			stop:     5,
			expected: errRangeDoesNotExist,
		},
		{
			name:     "valid block range",
			start:    1,
			stop:     5,
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rngReqs := testSrvc.ViewCurrentRanges()
			require.Equal(t, len(rngReqs), 1)

			err := testSrvc.RemoveRange(tc.start, tc.stop)
			if err != nil {
				require.Equal(t, tc.expected, err)
				return
			}

			rngReqs = testSrvc.ViewCurrentRanges()
			require.Equal(t, len(rngReqs), 0)
		})
	}
}

func TestViewCurrentRanges(t *testing.T) {
	rngReq := []RangeRequest{{
		Start:  1,
		Stop:   5,
		Params: statediff.Params{},
	},
		{
			Start:  5,
			Stop:   10,
			Params: statediff.Params{},
		}}

	testSrvc := newTestService(rngReq)
	rngReqs := testSrvc.ViewCurrentRanges()
	require.Equal(t, len(rngReqs), 2)
}
