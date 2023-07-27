package main

import (
	"fmt"
	"github.com/xuliangTang/mykubelet/pkg/core"
	"github.com/xuliangTang/mykubelet/pkg/kubelet/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const hostName = "mylain"

func main() {
	client := initClient()
	myKubelet := core.NewMyKubelet(client, hostName)

	fmt.Println("开始监听")
	myKubelet.StartStatusManager()
	for item := range myKubelet.PodConfig.Updates() {
		switch item.Op {
		case types.ADD:
			myKubelet.HandlePodAdditions(item.Pods)
		case types.UPDATE, types.DELETE:
			myKubelet.HandlePodUpdates(item.Pods)
		case types.REMOVE:
			myKubelet.HandlePodRemoves(item.Pods)
		}
	}
}

func initClient() *kubernetes.Clientset {
	kubeConfig, _ := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	client, _ := kubernetes.NewForConfig(kubeConfig)
	return client
}
