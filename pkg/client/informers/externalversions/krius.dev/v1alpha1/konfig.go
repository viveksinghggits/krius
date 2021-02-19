/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	kriusdevv1alpha1 "github.com/viveksinghggits/krius/pkg/apis/krius.dev/v1alpha1"
	versioned "github.com/viveksinghggits/krius/pkg/client/clientset/versioned"
	internalinterfaces "github.com/viveksinghggits/krius/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/viveksinghggits/krius/pkg/client/listers/krius.dev/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// KonfigInformer provides access to a shared informer and lister for
// Konfigs.
type KonfigInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.KonfigLister
}

type konfigInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewKonfigInformer constructs a new informer for Konfig type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewKonfigInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredKonfigInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredKonfigInformer constructs a new informer for Konfig type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredKonfigInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.KriusV1alpha1().Konfigs(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.KriusV1alpha1().Konfigs(namespace).Watch(context.TODO(), options)
			},
		},
		&kriusdevv1alpha1.Konfig{},
		resyncPeriod,
		indexers,
	)
}

func (f *konfigInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredKonfigInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *konfigInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&kriusdevv1alpha1.Konfig{}, f.defaultInformer)
}

func (f *konfigInformer) Lister() v1alpha1.KonfigLister {
	return v1alpha1.NewKonfigLister(f.Informer().GetIndexer())
}
