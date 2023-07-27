package events

// Container event reason list
const (
	CreatedContainer        = "Created"
	StartedContainer        = "Started"
	FailedToCreateContainer = "Failed"
	FailedToStartContainer  = "Failed"
	KillingContainer        = "Killing"
	PreemptContainer        = "Preempting"
	BackOffStartContainer   = "BackOff"
	ExceededGracePeriod     = "ExceededGracePeriod"
)

// Pod event reason list
const (
	FailedToKillPod                = "FailedKillPod"
	FailedToCreatePodContainer     = "FailedCreatePodContainer"
	FailedToMakePodDataDirectories = "Failed"
	NetworkNotReady                = "NetworkNotReady"
)

// Image event reason list
const (
	PullingImage            = "Pulling"
	PulledImage             = "Pulled"
	FailedToPullImage       = "Failed"
	FailedToInspectImage    = "InspectFailed"
	ErrImageNeverPullPolicy = "ErrImageNeverPull"
	BackOffPullImage        = "BackOff"
)

// kubelet event reason list
const (
	NodeReady                            = "NodeReady"
	NodeNotReady                         = "NodeNotReady"
	NodeSchedulable                      = "NodeSchedulable"
	NodeNotSchedulable                   = "NodeNotSchedulable"
	StartingKubelet                      = "Starting"
	KubeletSetupFailed                   = "KubeletSetupFailed"
	FailedAttachVolume                   = "FailedAttachVolume"
	FailedMountVolume                    = "FailedMount"
	VolumeResizeFailed                   = "VolumeResizeFailed"
	VolumeResizeSuccess                  = "VolumeResizeSuccessful"
	FileSystemResizeFailed               = "FileSystemResizeFailed"
	FileSystemResizeSuccess              = "FileSystemResizeSuccessful"
	FailedMapVolume                      = "FailedMapVolume"
	WarnAlreadyMountedVolume             = "AlreadyMountedVolume"
	SuccessfulAttachVolume               = "SuccessfulAttachVolume"
	SuccessfulMountVolume                = "SuccessfulMountVolume"
	NodeRebooted                         = "Rebooted"
	ContainerGCFailed                    = "ContainerGCFailed"
	ImageGCFailed                        = "ImageGCFailed"
	FailedNodeAllocatableEnforcement     = "FailedNodeAllocatableEnforcement"
	SuccessfulNodeAllocatableEnforcement = "NodeAllocatableEnforced"
	SandboxChanged                       = "SandboxChanged"
	FailedCreatePodSandBox               = "FailedCreatePodSandBox"
	FailedStatusPodSandBox               = "FailedPodSandBoxStatus"
	FailedMountOnFilesystemMismatch      = "FailedMountOnFilesystemMismatch"
)

// Image manager event reason list
const (
	InvalidDiskCapacity = "InvalidDiskCapacity"
	FreeDiskSpaceFailed = "FreeDiskSpaceFailed"
)

// Probe event reason list
const (
	ContainerUnhealthy    = "Unhealthy"
	ContainerProbeWarning = "ProbeWarning"
)

// Pod worker event reason list
const (
	FailedSync = "FailedSync"
)

// Config event reason list
const (
	FailedValidation = "FailedValidation"
)

// Lifecycle hooks
const (
	FailedPostStartHook = "FailedPostStartHook"
	FailedPreStopHook   = "FailedPreStopHook"
)
