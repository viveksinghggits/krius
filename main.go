package main

import (
	"flag"
	"log"
	"path/filepath"
	"time"

	kclientcmd "github.com/viveksinghggits/krius/pkg/client/clientset/versioned"
	genkoninformer "github.com/viveksinghggits/krius/pkg/client/informers/externalversions"
	"github.com/viveksinghggits/krius/pkg/controller"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Printf("Building config from flags, %s", err.Error())
	}

	kubeclient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("Building kubeclient %s\n", err.Error())
	}

	kcs, err := kclientcmd.NewForConfig(config)
	if err != nil {
		log.Printf("Building clientset from config, %s", err.Error())
	}

	konfigInfFactory := genkoninformer.NewSharedInformerFactory(kcs, 30*time.Second)

	stopCh := make(chan struct{})

	c := controller.NewController(kubeclient, kcs, konfigInfFactory.Krius().V1alpha1().Konfigs(), konfigInfFactory.Krius().V1alpha1().Sekrets())
	konfigInfFactory.Start(stopCh)

	if err = c.Run(2, stopCh); err != nil {
		log.Fatalf("Running controller %s\n", err.Error())
	}
}
