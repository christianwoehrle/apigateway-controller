package inject

import (
	apigatewayv1beta1 "github.com/christianwoehrle/apigateway-controller/pkg/apis/apigateway/v1beta1"
	rscheme "github.com/christianwoehrle/apigateway-controller/pkg/client/clientset/versioned/scheme"
	"github.com/christianwoehrle/apigateway-controller/pkg/controller/apigateway"
	"github.com/christianwoehrle/apigateway-controller/pkg/inject/args"
	"github.com/kubernetes-sigs/kubebuilder/pkg/inject/run"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
)

func init() {
	rscheme.AddToScheme(scheme.Scheme)

	// Inject Informers
	Inject = append(Inject, func(arguments args.InjectArgs) error {
		Injector.ControllerManager = arguments.ControllerManager

		if err := arguments.ControllerManager.AddInformerProvider(&apigatewayv1beta1.ApiGateway{}, arguments.Informers.Apigateway().V1beta1().ApiGateways()); err != nil {
			return err
		}

		// Add Kubernetes informers
		if err := arguments.ControllerManager.AddInformerProvider(&corev1.Service{}, arguments.KubernetesInformers.Core().V1().Services()); err != nil {
			return err
		}

		if c, err := apigateway.ProvideController(arguments); err != nil {
			return err
		} else {
			arguments.ControllerManager.AddController(c)
		}
		return nil
	})

	// Inject CRDs
	Injector.CRDs = append(Injector.CRDs, &apigatewayv1beta1.ApiGatewayCRD)
	// Inject PolicyRules
	Injector.PolicyRules = append(Injector.PolicyRules, rbacv1.PolicyRule{
		APIGroups: []string{"apigateway.cw.com"},
		Resources: []string{"*"},
		Verbs:     []string{"*"},
	})
	// Inject GroupVersions
	Injector.GroupVersions = append(Injector.GroupVersions, schema.GroupVersion{
		Group:   "apigateway.cw.com",
		Version: "v1beta1",
	})
	Injector.RunFns = append(Injector.RunFns, func(arguments run.RunArguments) error {
		Injector.ControllerManager.RunInformersAndControllers(arguments)
		return nil
	})
}
