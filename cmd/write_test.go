package cmd

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	gethsd "github.com/ethereum/go-ethereum/statediff"
	"github.com/stretchr/testify/require"
)

type mockService struct {
	reqCount int
}

func (ms *mockService) WriteStateDiffAt(_ uint64, _ gethsd.Params) error {
	ms.reqCount++
	return nil
}

func TestProcessRanges(t *testing.T) {
	blockRangesCh := make(chan blockRange)
	srv := new(mockService)

	go func() {
		blockRangesCh <- blockRange{uint64(1), uint64(5)}
		blockRangesCh <- blockRange{uint64(8), uint64(10)}
		blockRangesCh <- blockRange{uint64(6), uint64(7)}
		blockRangesCh <- blockRange{uint64(50), uint64(100)}
		blockRangesCh <- blockRange{uint64(5), uint64(8)}
		close(blockRangesCh)
	}()

	processRanges(srv, gethsd.Params{}, blockRangesCh)
	require.Equal(t, 65, srv.reqCount)
}

func TestHttpEndpoint(t *testing.T) {
	addr := ":11111"
	queueSize := 5
	blockRangesCh := make(chan blockRange, queueSize)
	srv := new(mockService)

	go startServer(addr, blockRangesCh)

	go func() {
		br := []blockRange{
			{uint64(1), uint64(5)},
			{uint64(8), uint64(10)},
			{uint64(6), uint64(7)},
			{uint64(50), uint64(100)},
			{uint64(5), uint64(8)},
			// Below request should fail since server has queue size of 5
			{uint64(52), uint64(328)},
			{uint64(35), uint64(428)},
			{uint64(45), uint64(844)},
		}

		for idx, r := range br {
			res, err := http.Get(fmt.Sprintf("http://localhost:11111/writeDiff?start=%d&end=%d", r[0], r[1]))
			require.NoError(t, err)
			require.NotNil(t, res)
			if idx < queueSize {
				require.Equal(t, res.StatusCode, 200)
			} else {
				require.Equal(t, res.StatusCode, 500)
			}
			require.NoError(t, res.Body.Close())
		}
		processRanges(srv, gethsd.Params{}, blockRangesCh)
	}()

	time.Sleep(1 * time.Second)
	require.Equal(t, 65, srv.reqCount)
}
