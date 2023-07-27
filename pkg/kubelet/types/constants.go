package types

const (
	// ResolvConfDefault is the system default DNS resolver configuration.
	ResolvConfDefault = "/etc/resolv.conf"
	// RFC3339NanoFixed is the fixed width version of time.RFC3339Nano.
	RFC3339NanoFixed = "2006-01-02T15:04:05.000000000Z07:00"
	// RFC3339NanoLenient is the variable width RFC3339 time format for lenient parsing of strings into timestamps.
	RFC3339NanoLenient = "2006-01-02T15:04:05.999999999Z07:00"
)

// Different container runtimes.
const (
	DockerContainerRuntime = "docker"
	RemoteContainerRuntime = "remote"
)

// User visible keys for managing node allocatable enforcement on the node.
const (
	NodeAllocatableEnforcementKey = "pods"
	SystemReservedEnforcementKey  = "system-reserved"
	KubeReservedEnforcementKey    = "kube-reserved"
	NodeAllocatableNoneKey        = "none"
)

// SwapBehavior types
const (
	LimitedSwap   = "LimitedSwap"
	UnlimitedSwap = "UnlimitedSwap"
)
