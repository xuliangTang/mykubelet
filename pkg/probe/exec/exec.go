package exec

import (
	"github.com/xuliangTang/mykubelet/pkg/probe"

	"k8s.io/utils/exec"
)

const (
	maxReadLength = 10 * 1 << 10 // 10KB
)

// New creates a Prober.
func New() Prober {
	return execProber{}
}

// Prober is an interface defining the Probe object for container readiness/liveness checks.
type Prober interface {
	Probe(e exec.Cmd) (probe.Result, string, error)
}

type execProber struct{}

// Probe 不支持
func (pr execProber) Probe(e exec.Cmd) (probe.Result, string, error) {
	return probe.Success, string(""), nil
}
