// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1beta1 "github.com/christianwoehrle/apigateway-controller/pkg/client/clientset/versioned/typed/apigateway/v1beta1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeApigatewayV1beta1 struct {
	*testing.Fake
}

func (c *FakeApigatewayV1beta1) ApiGateways(namespace string) v1beta1.ApiGatewayInterface {
	return &FakeApiGateways{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeApigatewayV1beta1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
