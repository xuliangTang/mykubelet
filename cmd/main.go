package main

import (
	"fmt"
	"github.com/xuliangTang/mykubelet/pkg/core"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const hostName = "mylain"

func main() {
	client := initClient()
	myKubelet := core.NewMyKubelet(client, hostName)
	myKubelet.SetOnAdd(func(pod *v1.Pod) error {
		fmt.Println("onAdd()", pod.Name)
		return nil
	})
	myKubelet.Run()
}

func initClient() *kubernetes.Clientset {
	kubeConfig, _ := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	client, _ := kubernetes.NewForConfig(kubeConfig)
	return client
}
