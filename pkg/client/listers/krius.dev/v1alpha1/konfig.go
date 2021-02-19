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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/viveksinghggits/krius/pkg/apis/krius.dev/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// KonfigLister helps list Konfigs.
// All objects returned here must be treated as read-only.
type KonfigLister interface {
	// List lists all Konfigs in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Konfig, err error)
	// Konfigs returns an object that can list and get Konfigs.
	Konfigs(namespace string) KonfigNamespaceLister
	KonfigListerExpansion
}

// konfigLister implements the KonfigLister interface.
type konfigLister struct {
	indexer cache.Indexer
}

// NewKonfigLister returns a new KonfigLister.
func NewKonfigLister(indexer cache.Indexer) KonfigLister {
	return &konfigLister{indexer: indexer}
}

// List lists all Konfigs in the indexer.
func (s *konfigLister) List(selector labels.Selector) (ret []*v1alpha1.Konfig, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Konfig))
	})
	return ret, err
}

// Konfigs returns an object that can list and get Konfigs.
func (s *konfigLister) Konfigs(namespace string) KonfigNamespaceLister {
	return konfigNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// KonfigNamespaceLister helps list and get Konfigs.
// All objects returned here must be treated as read-only.
type KonfigNamespaceLister interface {
	// List lists all Konfigs in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Konfig, err error)
	// Get retrieves the Konfig from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.Konfig, error)
	KonfigNamespaceListerExpansion
}

// konfigNamespaceLister implements the KonfigNamespaceLister
// interface.
type konfigNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Konfigs in the indexer for a given namespace.
func (s konfigNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Konfig, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Konfig))
	})
	return ret, err
}

// Get retrieves the Konfig from the indexer for a given namespace and name.
func (s konfigNamespaceLister) Get(name string) (*v1alpha1.Konfig, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("konfig"), name)
	}
	return obj.(*v1alpha1.Konfig), nil
}
