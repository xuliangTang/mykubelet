package queue

import (
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/clock"
)

// WorkQueue allows queuing items with a timestamp. An item is
// considered ready to process if the timestamp has expired.
type WorkQueue interface {
	// GetWork dequeues and returns all ready items.
	GetWork() []types.UID
	// Enqueue inserts a new item or overwrites an existing item.
	Enqueue(item types.UID, delay time.Duration)
}

type basicWorkQueue struct {
	clock clock.Clock
	lock  sync.Mutex
	queue map[types.UID]time.Time
}

var _ WorkQueue = &basicWorkQueue{}

// NewBasicWorkQueue returns a new basic WorkQueue with the provided clock
func NewBasicWorkQueue(clock clock.Clock) WorkQueue {
	queue := make(map[types.UID]time.Time)
	return &basicWorkQueue{queue: queue, clock: clock}
}

func (q *basicWorkQueue) GetWork() []types.UID {
	q.lock.Lock()
	defer q.lock.Unlock()
	now := q.clock.Now()
	var items []types.UID
	for k, v := range q.queue {
		if v.Before(now) {
			items = append(items, k)
			delete(q.queue, k)
		}
	}
	return items
}

func (q *basicWorkQueue) Enqueue(item types.UID, delay time.Duration) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.queue[item] = q.clock.Now().Add(delay)
}
