// Code generated by lister-gen. DO NOT EDIT.

package v1beta1

import (
	v1beta1 "github.com/christianwoehrle/apigateway-controller/pkg/apis/apigateway/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ApiGatewayLister helps list ApiGateways.
type ApiGatewayLister interface {
	// List lists all ApiGateways in the indexer.
	List(selector labels.Selector) (ret []*v1beta1.ApiGateway, err error)
	// ApiGateways returns an object that can list and get ApiGateways.
	ApiGateways(namespace string) ApiGatewayNamespaceLister
	ApiGatewayListerExpansion
}

// apiGatewayLister implements the ApiGatewayLister interface.
type apiGatewayLister struct {
	indexer cache.Indexer
}

// NewApiGatewayLister returns a new ApiGatewayLister.
func NewApiGatewayLister(indexer cache.Indexer) ApiGatewayLister {
	return &apiGatewayLister{indexer: indexer}
}

// List lists all ApiGateways in the indexer.
func (s *apiGatewayLister) List(selector labels.Selector) (ret []*v1beta1.ApiGateway, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.ApiGateway))
	})
	return ret, err
}

// ApiGateways returns an object that can list and get ApiGateways.
func (s *apiGatewayLister) ApiGateways(namespace string) ApiGatewayNamespaceLister {
	return apiGatewayNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ApiGatewayNamespaceLister helps list and get ApiGateways.
type ApiGatewayNamespaceLister interface {
	// List lists all ApiGateways in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1beta1.ApiGateway, err error)
	// Get retrieves the ApiGateway from the indexer for a given namespace and name.
	Get(name string) (*v1beta1.ApiGateway, error)
	ApiGatewayNamespaceListerExpansion
}

// apiGatewayNamespaceLister implements the ApiGatewayNamespaceLister
// interface.
type apiGatewayNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ApiGateways in the indexer for a given namespace.
func (s apiGatewayNamespaceLister) List(selector labels.Selector) (ret []*v1beta1.ApiGateway, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.ApiGateway))
	})
	return ret, err
}

// Get retrieves the ApiGateway from the indexer for a given namespace and name.
func (s apiGatewayNamespaceLister) Get(name string) (*v1beta1.ApiGateway, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1beta1.Resource("apigateway"), name)
	}
	return obj.(*v1beta1.ApiGateway), nil
}
