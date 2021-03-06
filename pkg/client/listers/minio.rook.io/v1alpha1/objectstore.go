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
	v1alpha1 "github.com/rook/rook/pkg/apis/minio.rook.io/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ObjectStoreLister helps list ObjectStores.
type ObjectStoreLister interface {
	// List lists all ObjectStores in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.ObjectStore, err error)
	// ObjectStores returns an object that can list and get ObjectStores.
	ObjectStores(namespace string) ObjectStoreNamespaceLister
	ObjectStoreListerExpansion
}

// objectStoreLister implements the ObjectStoreLister interface.
type objectStoreLister struct {
	indexer cache.Indexer
}

// NewObjectStoreLister returns a new ObjectStoreLister.
func NewObjectStoreLister(indexer cache.Indexer) ObjectStoreLister {
	return &objectStoreLister{indexer: indexer}
}

// List lists all ObjectStores in the indexer.
func (s *objectStoreLister) List(selector labels.Selector) (ret []*v1alpha1.ObjectStore, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ObjectStore))
	})
	return ret, err
}

// ObjectStores returns an object that can list and get ObjectStores.
func (s *objectStoreLister) ObjectStores(namespace string) ObjectStoreNamespaceLister {
	return objectStoreNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ObjectStoreNamespaceLister helps list and get ObjectStores.
type ObjectStoreNamespaceLister interface {
	// List lists all ObjectStores in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.ObjectStore, err error)
	// Get retrieves the ObjectStore from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.ObjectStore, error)
	ObjectStoreNamespaceListerExpansion
}

// objectStoreNamespaceLister implements the ObjectStoreNamespaceLister
// interface.
type objectStoreNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ObjectStores in the indexer for a given namespace.
func (s objectStoreNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.ObjectStore, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ObjectStore))
	})
	return ret, err
}

// Get retrieves the ObjectStore from the indexer for a given namespace and name.
func (s objectStoreNamespaceLister) Get(name string) (*v1alpha1.ObjectStore, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("objectstore"), name)
	}
	return obj.(*v1alpha1.ObjectStore), nil
}
