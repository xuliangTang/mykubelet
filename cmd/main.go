package main

import (
	"fmt"
	"github.com/xuliangTang/mykubelet/pkg/core"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const hostName = "mylain"

func main() {
	client := initClient()
	myKubelet := core.NewMyKubelet(client, hostName)

	myKubelet.SetOnPreAdd(func(opts *core.CallBackOptions) error {
		fmt.Println("onPreAdd()", opts.Pod.Name)
		return nil
	})

	myKubelet.SetOnAdd(func(opts *core.CallBackOptions) error {
		fmt.Println("onAdd()", opts.Pod.Name)
		opts.AddEvent("onAdd", "success")
		return nil
	})

	myKubelet.SetOnUpdate(func(opts *core.CallBackOptions) error {
		fmt.Println("onUpdate()", opts.Pod.Name)
		opts.AddEvent("onUpdate", "success")
		return nil
	})

	myKubelet.SetOnDelete(func(opts *core.CallBackOptions) error {
		fmt.Println("onDelete()", opts.Pod.Name)
		return nil
	})

	myKubelet.SetOnRemove(func(opts *core.CallBackOptions) error {
		fmt.Println("onRemove()", opts.Pod.Name)
		return nil
	})

	myKubelet.Run()
}

func initClient() *kubernetes.Clientset {
	kubeConfig, _ := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	client, _ := kubernetes.NewForConfig(kubeConfig)
	return client
}
