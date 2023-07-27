package results

import (
	"sync"

	kubecontainer "github.com/xuliangTang/mykubelet/pkg/kubelet/container"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Manager provides a probe results cache and channel of updates.
type Manager interface {
	// Get returns the cached result for the container with the given ID.
	Get(kubecontainer.ContainerID) (Result, bool)
	// Set sets the cached result for the container with the given ID.
	// The pod is only included to be sent with the update.
	Set(kubecontainer.ContainerID, Result, *v1.Pod)
	// Remove clears the cached result for the container with the given ID.
	Remove(kubecontainer.ContainerID)
	// Updates creates a channel that receives an Update whenever its result changes (but not
	// removed).
	// NOTE: The current implementation only supports a single updates channel.
	Updates() <-chan Update
}

// Result is the type for probe results.
type Result int

const (
	// Unknown is encoded as -1 (type Result)
	Unknown Result = iota - 1

	// Success is encoded as 0 (type Result)
	Success

	// Failure is encoded as 1 (type Result)
	Failure
)

func (r Result) String() string {
	switch r {
	case Success:
		return "Success"
	case Failure:
		return "Failure"
	default:
		return "UNKNOWN"
	}
}

// ToPrometheusType translates a Result to a form which is better understood by prometheus.
func (r Result) ToPrometheusType() float64 {
	switch r {
	case Success:
		return 0
	case Failure:
		return 1
	default:
		return -1
	}
}

// Update is an enum of the types of updates sent over the Updates channel.
type Update struct {
	ContainerID kubecontainer.ContainerID
	Result      Result
	PodUID      types.UID
}

// Manager implementation.
type manager struct {
	// guards the cache
	sync.RWMutex
	// map of container ID -> probe Result
	cache map[kubecontainer.ContainerID]Result
	// channel of updates
	updates chan Update
}

var _ Manager = &manager{}

// NewManager creates and returns an empty results manager.
func NewManager() Manager {
	return &manager{
		cache:   make(map[kubecontainer.ContainerID]Result),
		updates: make(chan Update, 20),
	}
}

func (m *manager) Get(id kubecontainer.ContainerID) (Result, bool) {
	m.RLock()
	defer m.RUnlock()
	result, found := m.cache[id]
	return result, found
}

func (m *manager) Set(id kubecontainer.ContainerID, result Result, pod *v1.Pod) {
	if m.setInternal(id, result) {
		m.updates <- Update{id, result, pod.UID}
	}
}

// Internal helper for locked portion of set. Returns whether an update should be sent.
func (m *manager) setInternal(id kubecontainer.ContainerID, result Result) bool {
	m.Lock()
	defer m.Unlock()
	prev, exists := m.cache[id]
	if !exists || prev != result {
		m.cache[id] = result
		return true
	}
	return false
}

func (m *manager) Remove(id kubecontainer.ContainerID) {
	m.Lock()
	defer m.Unlock()
	delete(m.cache, id)
}

func (m *manager) Updates() <-chan Update {
	return m.updates
}
