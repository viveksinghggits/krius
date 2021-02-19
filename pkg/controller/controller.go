package controller

import (
	"context"
	"fmt"
	"log"
	"time"

	kriusv1alpha1 "github.com/viveksinghggits/krius/pkg/apis/krius.dev/v1alpha1"
	kclientcmd "github.com/viveksinghggits/krius/pkg/client/clientset/versioned"
	skeme "github.com/viveksinghggits/krius/pkg/client/clientset/versioned/scheme"
	koninformer "github.com/viveksinghggits/krius/pkg/client/informers/externalversions/krius.dev/v1alpha1"
	klisters "github.com/viveksinghggits/krius/pkg/client/listers/krius.dev/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	uruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	kubeclient kubernetes.Interface
	konclient  kclientcmd.Interface

	konlister klisters.KonfigLister

	konfigSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work
	// to be processed instead of performing it as soon  as change happen
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

func NewController(kubeclientset kubernetes.Interface, konclientset kclientcmd.Interface, kinformer koninformer.KonfigInformer) *Controller {
	// add controller types to the scheme
	// so that event can be recoreded for these types
	uruntime.Must(skeme.AddToScheme(scheme.Scheme))
	log.Println("Creating even broadcraster")
	eveBroadCaster := record.NewBroadcaster()
	eveBroadCaster.StartStructuredLogging(0)
	eveBroadCaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{
		Interface: kubeclientset.CoreV1().Events(""),
	})
	recorder := eveBroadCaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "Krius"})

	controller := &Controller{
		kubeclient:   kubeclientset,
		konclient:    konclientset,
		konlister:    kinformer.Lister(),
		workqueue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Konfigs"),
		konfigSynced: kinformer.Informer().HasSynced,
		recorder:     recorder,
	}

	kinformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: controller.enqueueAddEvent,
			UpdateFunc: func(old, new interface{}) {
				controller.enqueueUpdateEvent(new)
			},
		},
	)

	return controller
}

func (c *Controller) Run(thrediness int, stopCh <-chan struct{}) error {
	log.Println("Starting Konfig Controlller")

	log.Println("Waiting for informer cache to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.konfigSynced); !ok {
		return fmt.Errorf("failed to wait for cache to sync")
	}

	log.Println("Starting workers")
	for i := 0; i < thrediness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	log.Println("Started workers")
	<-stopCh
	log.Println("Shutting down workers")

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	fmt.Println("ProcessNextItem")
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	err := func(o interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			// there is something else in the queue, we expect string values
			c.workqueue.Forget(obj)
			uruntime.HandleError(fmt.Errorf("expecting string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.SyncKonfig(key); err != nil {
			// was not successful, add again to the queue
			c.workqueue.AddRateLimited(key)
			uruntime.HandleError(fmt.Errorf("error syncing '%s': %s requeueing", key, err.Error()))
			return nil
		}
		c.workqueue.Forget(key)
		fmt.Printf("Successfully synced %+v\n", key)
		return nil
	}(obj)

	if err != nil {
		uruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) SyncKonfig(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		uruntime.HandleError(fmt.Errorf("invalid resource key : %s", key))
		return nil
	}

	kon, err := c.konlister.Konfigs(namespace).Get(name)
	if err != nil {
		// the konfig that triggered this even, is no longer there
		if errors.IsNotFound(err) {
			uruntime.HandleError(fmt.Errorf("konfig '%s', in workqueue no londer exists", kon.Name))
			return nil
		}
		return err
	}

	ctx := context.Background()
	secretData := kon.Spec.Data
	// iterate over namespaces create secret
	namespaces, err := c.kubeclient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, ns := range namespaces.Items {
		_, err := c.kubeclient.CoreV1().ConfigMaps(ns.Name).Get(ctx, kon.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			konf := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      kon.Name,
					Namespace: ns.Name,
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(kon, kriusv1alpha1.SchemeGroupVersion.WithKind("Konfig")),
					},
				},
				Data: secretData,
			}
			_, err = c.kubeclient.CoreV1().ConfigMaps(ns.Name).Create(ctx, konf, metav1.CreateOptions{})
			if err != nil {
				c.recorder.Event(kon, corev1.EventTypeWarning, "Not Synced", fmt.Sprintf("Error syncing Konfig to the namespace %s", ns.Name))
				return err
			}
			// update the event
			c.recorder.Event(kon, corev1.EventTypeNormal, "Synced", fmt.Sprintf("to the namespace %s", ns.Name))
		}

		if err != nil {
			return err
		}
	}

	if err := c.updateKonfigStatus(ctx, kon, "Success"); err != nil {
		return err
	}

	return nil
}

func (c *Controller) updateKonfigStatus(ctx context.Context, k *kriusv1alpha1.Konfig, msg string) error {
	kcopy := k.DeepCopy()
	kcopy.Status.Synced = true
	log.Printf("Before updating the konfig status %+v", kcopy)
	updatedk, err := c.konclient.KriusV1alpha1().Konfigs(k.Namespace).Update(ctx, kcopy, metav1.UpdateOptions{})
	if err != nil {
		log.Printf("error is %s", err.Error())
	}
	log.Printf("updated kon is %+v", updatedk)
	return err
}

func (c *Controller) enqueueUpdateEvent(obj interface{}) {
	// var key string
	// var err error
	// if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
	// 	uruntime.HandleError(err)
	// 	return
	// }
	// c.workqueue.Add(key)
}

func (c *Controller) enqueueAddEvent(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		uruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}
