package core

import (
	"github.com/xuliangTang/mykubelet/pkg/kubelet/container"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"time"
)

// SetPodReady 构建PodStatus的sandbox为ready containers为running
func SetPodReady(pod *v1.Pod) *container.PodStatus {
	status := &container.PodStatus{
		ID:        types.UID(pod.UID),
		Name:      pod.Name,
		Namespace: pod.Namespace,
		SandboxStatuses: []*runtimeapi.PodSandboxStatus{
			{
				Id:    string(pod.UID),
				State: runtimeapi.PodSandboxState_SANDBOX_READY,
			},
		},
	}

	var containerStatus []*container.Status
	for _, c := range pod.Spec.Containers {
		cs := &container.Status{
			Name:      c.Name,
			Image:     c.Image,
			State:     container.ContainerStateRunning,
			CreatedAt: time.Now(),
			StartedAt: time.Now().Add(time.Second * 3),
		}
		containerStatus = append(containerStatus, cs)
	}
	status.ContainerStatuses = containerStatus

	return status
}

// SetPodCompleted 构建podStatus状态为completed
func SetPodCompleted(pod *v1.Pod) *container.PodStatus {
	status := &container.PodStatus{
		ID:        types.UID(pod.UID),
		Name:      pod.Name,
		Namespace: pod.Namespace,
		SandboxStatuses: []*runtimeapi.PodSandboxStatus{
			{
				Id:    string(pod.UID),
				State: runtimeapi.PodSandboxState_SANDBOX_NOTREADY,
			},
		},
	}

	var containerStatus []*container.Status
	for _, c := range pod.Spec.Containers {
		cs := &container.Status{
			Name:       c.Name,
			Image:      c.Image,
			State:      container.ContainerStateExited,
			ExitCode:   0,
			Reason:     "Completed",
			FinishedAt: time.Now(),
		}
		containerStatus = append(containerStatus, cs)
	}
	status.ContainerStatuses = containerStatus

	return status
}
