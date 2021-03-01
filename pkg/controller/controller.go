package controller

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"

	kriusv1alpha1 "github.com/viveksinghggits/krius/pkg/apis/krius.dev/v1alpha1"
	kclientcmd "github.com/viveksinghggits/krius/pkg/client/clientset/versioned"
	skeme "github.com/viveksinghggits/krius/pkg/client/clientset/versioned/scheme"
	crinformer "github.com/viveksinghggits/krius/pkg/client/informers/externalversions/krius.dev/v1alpha1"
	crlisters "github.com/viveksinghggits/krius/pkg/client/listers/krius.dev/v1alpha1"
	"github.com/viveksinghggits/krius/pkg/util"

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

	konlister    crlisters.KonfigLister
	konfigSynced cache.InformerSynced

	seklister crlisters.SekretLister
	sekSynced cache.InformerSynced
	// workqueue is a rate limited work queue. This is used to queue work
	// to be processed instead of performing it as soon  as change happen
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

func NewController(kubeclientset kubernetes.Interface, konclientset kclientcmd.Interface, kinformer crinformer.KonfigInformer, sekinformer crinformer.SekretInformer) *Controller {
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
		workqueue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "konsek"),
		konfigSynced: kinformer.Informer().HasSynced,
		seklister:    sekinformer.Lister(),
		sekSynced:    sekinformer.Informer().HasSynced,
		recorder:     recorder,
	}

	// add events for konfig resource
	kinformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: controller.enqueueAddEvent,
			// UpdateFunc: func(old, new interface{}) {
			// 	controller.enqueueUpdateEvent(new)
			// },
		},
	)

	// add event for sekret resources
	sekinformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: controller.enqueueAddEvent,
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
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}
	log.Printf("obejct that we got from queue %+v\n", obj)
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

		// since we have added kind as well in the key, let's split that
		s := strings.Split(key, ":")
		log.Printf("s that we have is %+v\n", s)
		switch s[1] {
		case "konfig":
			if err := c.SyncKonfig(s[0]); err != nil {
				// was not successful, add again to the queue
				c.workqueue.AddRateLimited(key)
				uruntime.HandleError(fmt.Errorf("error syncing '%s': %s requeueing", key, err.Error()))
				return nil
			}
			c.workqueue.Forget(key)
			fmt.Printf("Successfully synced %+v\n", key)
		case "sekret":
			if err := c.SyncSekret(s[0]); err != nil {
				c.workqueue.AddRateLimited(key)
				uruntime.HandleError(fmt.Errorf("error syncing '%s': %s requeueing", key, err.Error()))
				return nil
			}
			c.workqueue.Forget(key)
			fmt.Printf("Successfully synced %+v\n", key)
		}

		return nil
	}(obj)

	if err != nil {
		uruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) SyncSekret(key string) error {
	log.Println("syncsekret was called")
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		uruntime.HandleError(fmt.Errorf("invalid resource key : %s", key))
		return nil
	}

	// get the sekret resource that was created
	sek, err := c.seklister.Sekrets(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			uruntime.HandleError(fmt.Errorf("sekret '%s', that was in workqueue but doesn't exist in cluster anymore", sek.Name))
			return nil
		}
		return err
	}

	ctx := context.Background()
	if err := c.createSekrets(ctx, sek); err != nil {
		return err
	}

	if err := c.updateSekretStatus(ctx, sek); err != nil {
		return err
	}

	return nil
}

func (c *Controller) createSekrets(ctx context.Context, s *kriusv1alpha1.Sekret) error {
	var err error
	nss, err := c.getNonSystemNSs(ctx, s.ObjectMeta)
	if err != nil {
		return err
	}

	for _, ns := range nss {
		_, err := c.kubeclient.CoreV1().Secrets(ns.Name).Get(ctx, s.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			sec := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      s.Name,
					Namespace: ns.Name,
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(s, kriusv1alpha1.SchemeGroupVersion.WithKind("Sekret")),
					},
				},
				Data: formSecretData(s.Spec.Data),
				Type: s.Spec.Type,
			}
			_, err = c.kubeclient.CoreV1().Secrets(ns.Name).Create(ctx, sec, metav1.CreateOptions{})
			if err != nil {
				c.recorder.Event(sec, corev1.EventTypeWarning, "Not Synced", fmt.Sprintf("Error syncing Sekret to the namespace %s", ns.Name))
				return err
			}
			// update the event
			c.recorder.Event(s, corev1.EventTypeNormal, "Synced", fmt.Sprintf("to the namespace %s", ns.Name))
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func formSecretData(plain map[string]string) map[string][]byte {
	decoded := map[string][]byte{}
	for k, v := range plain {
		decoded[k] = []byte(base64.StdEncoding.EncodeToString([]byte(v)))
	}
	return decoded
}

func (c *Controller) updateSekretStatus(ctx context.Context, s *kriusv1alpha1.Sekret) error {
	sek := s.DeepCopy()
	sek.Status.Synced = true
	_, err := c.konclient.KriusV1alpha1().Sekrets(s.Namespace).Update(ctx, sek, metav1.UpdateOptions{})
	return err
}

func (c *Controller) SyncKonfig(key string) error {
	log.Printf("SyncKonfig wa called with key %s\n", key)
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
	kData := kon.Spec.Data
	// iterate over namespaces create secret
	// TODO: @viveksinghggits, we can just pass kon from here
	if err := c.createKonfigs(ctx, kon, kData); err != nil {
		return err
	}

	if err := c.updateKonfigStatus(ctx, kon); err != nil {
		return err
	}

	return nil
}

func (c *Controller) getNonSystemNSs(ctx context.Context, objectMeta metav1.ObjectMeta) ([]corev1.Namespace, error) {
	namespaces, err := c.kubeclient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	if metav1.HasAnnotation(objectMeta, util.IncludeKubeSysNSAnn) {
		log.Println("Annotation is not found")
		return namespaces.Items, nil
	}

	var nss []corev1.Namespace
	for _, ns := range namespaces.Items {
		if !strings.HasPrefix(ns.Name, util.SystemNSsPrefix) {
			nss = append(nss, ns)
		}
	}
	return nss, nil
}

func (c *Controller) createKonfigs(ctx context.Context, kon *kriusv1alpha1.Konfig, data map[string]string) error {
	var err error
	nss, err := c.getNonSystemNSs(ctx, kon.ObjectMeta)
	if err != nil {
		return err
	}

	for _, ns := range nss {
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
				Data: data,
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
	return nil
}

func (c *Controller) updateKonfigStatus(ctx context.Context, k *kriusv1alpha1.Konfig) error {
	kcopy := k.DeepCopy()
	kcopy.Status.Synced = true

	_, err := c.konclient.KriusV1alpha1().Konfigs(k.Namespace).Update(ctx, kcopy, metav1.UpdateOptions{})
	if err != nil {
		log.Printf("error is %s", err.Error())
	}
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
	log.Printf("enqueue Add event was called with %+v\n", obj)
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		uruntime.HandleError(err)
		return
	}

	var kind string

	switch a := obj.(type) {
	case *kriusv1alpha1.Konfig:
		kind = "konfig"
	case *kriusv1alpha1.Sekret:
		kind = "sekret"
	default:
		log.Printf("Expecting type konfig or sekret but got %+v\n", a)
	}

	// namespace/name:kind
	// Konfig, Sekret
	// this doesn't look good, we will have to change this so that we can figure out type without adding type here
	c.workqueue.Add(fmt.Sprintf("%s:%s", key, kind))
}
