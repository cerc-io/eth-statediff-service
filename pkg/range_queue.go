package statediff

import (
	"errors"
	"sync"
)

var (
	errEmptyQueue = errors.New("queue is empty")
)

type blockRangeQueue struct {
	sync.RWMutex
	rangeQueue []RangeRequest
}

func (q *blockRangeQueue) push(rng RangeRequest) {
	q.Lock()
	defer q.Unlock()
	q.rangeQueue = append(q.rangeQueue, rng)
}

func (q *blockRangeQueue) pop() (RangeRequest, error) {
	q.Lock()
	defer q.Unlock()
	if len(q.rangeQueue) > 0 {
		rng := q.rangeQueue[0]
		q.rangeQueue = q.rangeQueue[1:]
		return rng, nil
	}
	return RangeRequest{}, errEmptyQueue
}

func (q *blockRangeQueue) search(rngReq RangeRequest) (int, error) {
	for idx, r := range q.rangeQueue {
		if r.Start == rngReq.Start && r.Stop == rngReq.Stop {
			return idx, nil
		}
	}

	return 0, errEmptyQueue
}

func (q *blockRangeQueue) get() []RangeRequest {
	q.RLock()
	defer q.RUnlock()
	queueRanges := make([]RangeRequest, len(q.rangeQueue))
	copy(queueRanges, q.rangeQueue)
	return queueRanges
}

func (q *blockRangeQueue) remove(req RangeRequest) error {
	q.Lock()
	defer q.Unlock()
	idx, err := q.search(req)
	if err != nil {
		return err
	}

	q.rangeQueue = append(q.rangeQueue[:idx], q.rangeQueue[idx+1:]...)
	return nil
}
