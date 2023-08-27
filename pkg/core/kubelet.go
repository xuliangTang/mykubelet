package core

import (
	"context"
	"fmt"
	"github.com/xuliangTang/mykubelet/pkg/api/legacyscheme"
	apisv1 "github.com/xuliangTang/mykubelet/pkg/apis/core/v1"
	"github.com/xuliangTang/mykubelet/pkg/kubelet/config"
	"github.com/xuliangTang/mykubelet/pkg/kubelet/configmap"
	kubecontainer "github.com/xuliangTang/mykubelet/pkg/kubelet/container"
	"github.com/xuliangTang/mykubelet/pkg/kubelet/pod"
	"github.com/xuliangTang/mykubelet/pkg/kubelet/prober"
	"github.com/xuliangTang/mykubelet/pkg/kubelet/prober/results"
	"github.com/xuliangTang/mykubelet/pkg/kubelet/secret"
	"github.com/xuliangTang/mykubelet/pkg/kubelet/status"
	kubetypes "github.com/xuliangTang/mykubelet/pkg/kubelet/types"
	"github.com/xuliangTang/mykubelet/pkg/kubelet/util/queue"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
	"os/exec"
	"time"
)

// CallBackOptions 回调参数
type CallBackOptions struct {
	Pod *v1.Pod

	eventRecorder record.EventRecorder
	podCache      kubecontainer.Cache
}

// GetCmdAndArgs 获取pod的commands和args
func (c *CallBackOptions) GetCmdAndArgs() []*exec.Cmd {
	var ret []*exec.Cmd

	for _, container := range c.Pod.Spec.Containers {
		if len(container.Command) == 0 {
			continue
		}

		var args []string
		if len(container.Command) > 1 {
			args = append(args, container.Command[1:]...)
		}
		args = append(args, container.Args...)
		ret = append(ret, exec.Command(container.Command[0], args...))
	}

	return ret
}

// GetContainerCmds 获取pod的commands和args 封装为ContainerCmd对象
func (c *CallBackOptions) GetContainerCmds() []*ContainerCmd {
	var ret []*ContainerCmd

	for _, container := range c.Pod.Spec.Containers {
		if len(container.Command) == 0 {
			continue
		}

		var args []string
		if len(container.Command) > 1 {
			args = append(args, container.Command[1:]...)
		}
		args = append(args, container.Args...)
		cmd := exec.Command(container.Command[0], args...)
		ret = append(ret, &ContainerCmd{
			Cmd:           cmd,
			ContainerName: container.Name,
		})
	}

	return ret
}

// AddEvent 记录normal事件
func (c *CallBackOptions) AddEvent(reason, msg string) {
	c.eventRecorder.Event(c.Pod, v1.EventTypeNormal, reason, msg)
}

// SetPodCompleted 设置pod状态为completed
func (c *CallBackOptions) SetPodCompleted() {
	podStatus := SetPodCompleted(c.Pod)
	c.podCache.Set(c.Pod.UID, podStatus, nil, time.Now())
}

// SetContainerExit 设置pod其中一个容器为退出
func (c *CallBackOptions) SetContainerExit(containerName string, exitCode int) {
	// 获取缓存中的podStatus
	getPodStatus, err := c.podCache.Get(c.Pod.UID)
	if err != nil {
		klog.Error(err)
		return
	}

	// 重新设置指定容器的状态
	podStatus := SetContainerExit(getPodStatus, c.Pod, containerName, exitCode)

	c.podCache.Set(c.Pod.UID, podStatus, nil, time.Now())
}

type CallBackFn func(opts *CallBackOptions) error

type MyKubelet struct {
	KubeClient    kubernetes.Interface
	HostName      string
	PodConfig     *config.PodConfig
	PodManager    pod.Manager
	PodWorkers    PodWorkers
	PodCache      kubecontainer.Cache
	statusManager status.Manager
	probeManager  prober.Manager
	reasonCache   *ReasonCache
	Clock         clock.RealClock

	// 回调
	onAdd, onUpdate, onDelete, onRemove CallBackFn
}

func NewMyKubelet(client kubernetes.Interface, hostName string) *MyKubelet {
	fact := informers.NewSharedInformerFactory(client, 0)
	fact.Core().V1().Nodes().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{})
	nodeLister := fact.Core().V1().Nodes().Lister()
	ch := make(chan struct{})
	fact.Start(ch)
	fact.WaitForCacheSync(ch)

	// 初始化podManager
	mirrorPodClient := pod.NewBasicMirrorClient(client, hostName, nodeLister)
	secretManager := secret.NewSimpleSecretManager(client)
	configMapManager := configmap.NewSimpleConfigMapManager(client)
	podManager := pod.NewBasicPodManager(mirrorPodClient, secretManager, configMapManager)

	// 初始化podConfig
	eventBroadcaster := record.NewBroadcaster()                     // 事件分发器广播
	if err := apisv1.AddToScheme(legacyscheme.Scheme); err != nil { // 注册scheme
		klog.Fatalln(err)
	}
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{ // 指定事件接收
		Interface: client.CoreV1().Events(""),
	})
	eventRecorder := eventBroadcaster.NewRecorder(legacyscheme.Scheme, v1.EventSource{Component: "kubelet", Host: hostName})
	podConfig := config.NewPodConfig(config.PodConfigNotificationIncremental, eventRecorder)
	// 注入clientset
	config.NewSourceApiserver(client, types.NodeName(hostName), func() bool {
		return fact.Core().V1().Nodes().Informer().HasSynced()
	}, podConfig.Channel(kubetypes.ApiserverSource)) // 关联configCh，会把相关的内容注入到ch里

	mykubelet := &MyKubelet{
		KubeClient:  client,
		HostName:    hostName,
		PodConfig:   podConfig,
		PodManager:  podManager,
		reasonCache: NewReasonCache(),
	}

	// 初始化podWorker
	mykubelet.Clock = clock.RealClock{}
	mykubelet.PodCache = kubecontainer.NewCache()
	workQueue := queue.NewBasicWorkQueue(mykubelet.Clock)
	mykubelet.PodWorkers = NewPodWorkers(
		mykubelet.syncPod,
		mykubelet.syncTerminatingPod,
		mykubelet.syncTerminatedPod,
		eventRecorder,
		workQueue,
		time.Second*1,
		time.Second*10,
		mykubelet.PodCache,
		mykubelet.PodManager,
	)

	// 初始化statusManager
	mykubelet.statusManager = status.NewManager(client, mykubelet.PodManager, mykubelet)

	// 初始化probeManager
	lm, rm, sm := results.NewManager(), results.NewManager(), results.NewManager()
	mykubelet.probeManager = prober.NewManager(
		mykubelet.statusManager,
		lm,
		rm,
		sm,
		&ContainerCommandRunner{},
		eventRecorder)

	return mykubelet
}

// SetOnPreAdd 设置podWorker接收到updatess事件时的回调
func (m *MyKubelet) SetOnPreAdd(onPreAdd CallBackFn) {
	m.PodWorkers.(*podWorkers).OnPreAdd = onPreAdd
}

func (m *MyKubelet) SetOnAdd(onAdd CallBackFn) {
	m.onAdd = onAdd
}

func (m *MyKubelet) SetOnUpdate(onUpdate CallBackFn) {
	m.onUpdate = onUpdate
}

func (m *MyKubelet) SetOnDelete(onDelete CallBackFn) {
	m.onDelete = onDelete
}

func (m *MyKubelet) SetOnRemove(onRemove CallBackFn) {
	m.onRemove = onRemove
}

func (m *MyKubelet) StartStatusManager() {
	klog.Info("statusManager开始启动")
	m.statusManager.Start()
}

func (m *MyKubelet) Run() {
	klog.Info("边缘Kubelet开始启动")
	m.StartStatusManager()

	for item := range m.PodConfig.Updates() {
		switch item.Op {
		case kubetypes.ADD:
			m.HandlePodAdditions(item.Pods)
		case kubetypes.UPDATE:
			m.HandlePodUpdates(item.Pods)
		case kubetypes.DELETE:
			m.HandlePodDelete(item.Pods)
		case kubetypes.REMOVE:
			m.HandlePodRemoves(item.Pods)
		}
	}
}

func (m *MyKubelet) HandlePodAdditions(pods []*v1.Pod) {
	for _, p := range pods {
		m.PodManager.AddPod(p)
		m.dispatchWork(kubetypes.SyncPodCreate, p, m.Clock.Now())

		if m.onAdd != nil {
			opts := &CallBackOptions{
				Pod:           p,
				eventRecorder: m.PodWorkers.(*podWorkers).recorder,
				podCache:      m.PodCache,
			}
			if err := m.onAdd(opts); err != nil {
				klog.Errorln(err)
			}
		}
	}
}

func (m *MyKubelet) HandlePodUpdates(pods []*v1.Pod) {
	for _, p := range pods {
		m.PodManager.UpdatePod(p)
		m.dispatchWork(kubetypes.SyncPodUpdate, p, m.Clock.Now())

		if m.onUpdate != nil {
			opts := &CallBackOptions{
				Pod:           p,
				eventRecorder: m.PodWorkers.(*podWorkers).recorder,
				podCache:      m.PodCache,
			}
			if err := m.onUpdate(opts); err != nil {
				klog.Errorln(err)
			}
		}
	}
}

func (m *MyKubelet) HandlePodDelete(pods []*v1.Pod) {
	for _, p := range pods {
		m.PodManager.UpdatePod(p)
		m.dispatchWork(kubetypes.SyncPodUpdate, p, m.Clock.Now())

		if m.onDelete != nil {
			opts := &CallBackOptions{
				Pod:           p,
				eventRecorder: m.PodWorkers.(*podWorkers).recorder,
				podCache:      m.PodCache,
			}
			if err := m.onDelete(opts); err != nil {
				klog.Errorln(err)
			}
		}
	}
}

func (m *MyKubelet) HandlePodRemoves(pods []*v1.Pod) {
	for _, p := range pods {
		m.PodManager.DeletePod(p)
		m.dispatchWork(kubetypes.SyncPodKill, p, m.Clock.Now())

		if m.onRemove != nil {
			opts := &CallBackOptions{
				Pod:           p,
				eventRecorder: m.PodWorkers.(*podWorkers).recorder,
				podCache:      m.PodCache,
			}
			if err := m.onRemove(opts); err != nil {
				klog.Errorln(err)
			}
		}
	}
}

func (m *MyKubelet) dispatchWork(updateType kubetypes.SyncPodType, pod *v1.Pod, start time.Time) {
	m.PodWorkers.UpdatePod(UpdatePodOptions{
		UpdateType: updateType,
		Pod:        pod,
		StartTime:  start,
	})
}

func (m *MyKubelet) syncPod(ctx context.Context, updateType kubetypes.SyncPodType, pod, mirrorPod *v1.Pod, podStatus *kubecontainer.PodStatus) (isTerminal bool, err error) {
	fmt.Println("测试的syncPod")

	//isTerminal = false
	apiPodStatus := m.generateAPIPodStatus(pod, podStatus)
	m.statusManager.SetPodStatus(pod, apiPodStatus)
	//if apiPodStatus.Phase == v1.PodSucceeded || apiPodStatus.Phase == v1.PodFailed {
	//	isTerminal = true
	//}
	return true, nil
}

func (m *MyKubelet) syncTerminatingPod(ctx context.Context, pod *v1.Pod, podStatus *kubecontainer.PodStatus, runningPod *kubecontainer.Pod, gracePeriod *int64, podStatusFn func(*v1.PodStatus)) error {
	fmt.Println("测试的syncTerminatingPod")
	apiPodStatus := m.generateAPIPodStatus(pod, podStatus)
	m.statusManager.SetPodStatus(pod, apiPodStatus)
	return nil
}

func (m *MyKubelet) syncTerminatedPod(ctx context.Context, pod *v1.Pod, podStatus *kubecontainer.PodStatus) error {
	fmt.Println("测试的syncTerminatedPod 收尾工作")
	return nil
}

func (m *MyKubelet) PodResourcesAreReclaimed(pod *v1.Pod, status v1.PodStatus) bool {
	return true
}

func (m *MyKubelet) PodCouldHaveRunningContainers(pod *v1.Pod) bool {
	return true
}

var _ status.PodDeletionSafetyProvider = &MyKubelet{}

type ContainerCommandRunner struct{}

func (c ContainerCommandRunner) RunInContainer(id kubecontainer.ContainerID, cmd []string, timeout time.Duration) ([]byte, error) {
	return []byte(""), nil
}

var _ kubecontainer.CommandRunner = &ContainerCommandRunner{}
