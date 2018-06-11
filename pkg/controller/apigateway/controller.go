package apigateway

import (
	"log"

	"github.com/kubernetes-sigs/kubebuilder/pkg/controller"
	"github.com/kubernetes-sigs/kubebuilder/pkg/controller/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/record"

	"fmt"

	apigatewayv1beta1 "github.com/christianwoehrle/apigateway-controller/pkg/apis/apigateway/v1beta1"
	apigatewayv1beta1client "github.com/christianwoehrle/apigateway-controller/pkg/client/clientset/versioned/typed/apigateway/v1beta1"
	apigatewayv1beta1informer "github.com/christianwoehrle/apigateway-controller/pkg/client/informers/externalversions/apigateway/v1beta1"
	apigatewayv1beta1lister "github.com/christianwoehrle/apigateway-controller/pkg/client/listers/apigateway/v1beta1"
	"github.com/christianwoehrle/apigateway-controller/pkg/inject/args"
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Foo"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "Foo synced successfully"
	// ServiceLabel
	// Ingress has a Label named Service Label and a specific Value
	// Every Time a Service with a Label ServiceLabel and the same value is created/updated/deleted
	// the ingress adds/updates/deleted the handling of this service
	ServiceLabel = "ServiceLabel"

	// Hostname is the Hostname, through which a Service wants to be available
	// A Service specifies the hostname through the label Hostname
	ServiceHostname = "ServiceHostname"
	// ServicePath is the Path through which a Service wants to be available
	// A Service specifies the path through the label ServicePath
	ServicePath = "ServicePath"
)

// +kubebuilder:controller:group=apigateway,version=v1beta1,kind=ApiGateway,resource=apigateways
// +kubebuilder:informers:group=core,version=v1,kind=Service
// +kubebuilder:informers:group=extensions,version=v1beta1,kind=Ingress
// +kubebuilder:informers:group=core,version=v1,kind=Pod
type ApiGatewayController struct {
	// INSERT ADDITIONAL FIELDS HERE
	apigatewayLister apigatewayv1beta1lister.ApiGatewayLister
	apigatewayclient apigatewayv1beta1client.ApigatewayV1beta1Interface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	apigatewayrecorder record.EventRecorder
	args.InjectArgs
}

// EDIT THIS FILE
// This files was created by "kubebuilder create resource" for you to edit.
// Controller implementation logic for ApiGateway resources goes here.

// Reconcile compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (bc *ApiGatewayController) Reconcile(k types.ReconcileKey) error {
	// INSERT YOUR CODE HERE
	log.Printf("Reconcile Methode: %T Namespace: %s  Name: %s", k, k.Namespace, k.Name)
	apigw, err := bc.apigatewayLister.ApiGateways(k.Namespace).Get(k.Name)
	if err != nil {
		log.Printf("Couldn't get an ApiGateway with name: %s, Error: %v", k.Name, err)
	}
	apigw, err = bc.Informers.Apigateway().V1beta1().ApiGateways().Lister().ApiGateways(k.Namespace).Get(k.Name)
	if apigw == nil {
		log.Printf("Couldn't get an ApiGateway with name: %s, Error: %v", k.Name, err)
		return nil
	}
	ingressName := apigw.Spec.IngressName

	if ingressName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		runtime.HandleError(fmt.Errorf("%s: ingress name must be specified", k))
		return nil
	}

	log.Printf("apigw found for Name: %s", apigw.Name)

	// Get the deployment with the name specified in Foo.spec
	ingress, err := bc.KubernetesInformers.Extensions().V1beta1().Ingresses().Lister().Ingresses(k.Namespace).Get(ingressName)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		ingress, err = bc.KubernetesClientSet.ExtensionsV1beta1().Ingresses(k.Namespace).Create(newIngress(bc.KubernetesInformers, apigw))
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		log.Printf("Couldn't create the Ingress: %v", err)
		return err
	}

	// If the Ingress is not controlled by this Foo resource, we should log
	// a warning to the event recorder and ret
	if !metav1.IsControlledBy(ingress, apigw) {
		msg := fmt.Sprintf(MessageResourceExists, ingress.Name)
		bc.apigatewayrecorder.Event(apigw, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}

	// If this number of the replicas on the Foo resource is specified, and the
	// number does not equal the current desired replicas on the Deployment, we
	// should update the Deployment resource.

	ingressLabel := ingress.Labels[ServiceLabel]
	if apigw.Spec.ServiceLabel != "" && apigw.Spec.ServiceLabel != ingressLabel {
		glog.V(4).Infof("Foo %s replicas: %d, ServiceLabel: %d", apigw.Name, apigw.Spec.ServiceLabel, ingressLabel)
		ingress, err = bc.KubernetesClientSet.ExtensionsV1beta1().Ingresses(apigw.Namespace).Update(newIngress(bc.KubernetesInformers, apigw))
	}

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. THis could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// Finally, we update the status block of the Foo resource to reflect the
	// current state of the world
	err = bc.updateApigwStatus(*apigw, ingress)
	if err != nil {
		return err
	}

	bc.apigatewayrecorder.Event(apigw, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

// ProvideController provides a controller that will be run at startup.  Kubebuilder will use codegeneration
// to automatically register this controller in the inject package
func ProvideController(arguments args.InjectArgs) (*controller.GenericController, error) {
	// INSERT INITIALIZATIONS FOR ADDITIONAL FIELDS HERE
	bc := &ApiGatewayController{
		apigatewayLister: arguments.ControllerManager.GetInformerProvider(&apigatewayv1beta1.ApiGateway{}).(apigatewayv1beta1informer.ApiGatewayInformer).Lister(),

		apigatewayclient:   arguments.Clientset.ApigatewayV1beta1(),
		apigatewayrecorder: arguments.CreateRecorder("ApiGatewayController"),
		InjectArgs:         arguments,
	}

	// Create a new controller that will call ApiGatewayController.Reconcile on changes to ApiGateways
	gc := &controller.GenericController{
		Name:             "ApiGatewayController",
		Reconcile:        bc.Reconcile,
		InformerRegistry: arguments.ControllerManager,
	}

	//log.Printf("watch apigateway")
	if err := gc.Watch(&apigatewayv1beta1.ApiGateway{}); err != nil {
		return gc, err
	}

	err := gc.WatchEvents(&corev1.Service{},
		// This function returns the callbacks that will be invoked for events
		func(q workqueue.RateLimitingInterface) cache.ResourceEventHandler {
			// This function implements the same functionality as GenericController.Watch
			return cache.ResourceEventHandlerFuncs{
				AddFunc: bc.AddService,
				UpdateFunc: func(old, obj interface{}) {
					log.Printf("Service Updated %T %s\n", obj, obj.(*corev1.Service).Name)
					bc.DeleteService(obj)
					//q.AddRateLimited(eventhandlers.MapToSelf(obj))
				},
				DeleteFunc: func(obj interface{}) {
					log.Printf("Service Deleted %s\n", obj.(*corev1.Service).Name)
					//q.AddRateLimited(eventhandlers.MapToSelf(obj))
				},
			}
		})
	if err != nil {
		log.Fatalf("%v", err)
	}

	//log.Printf("watch services")
	// IMPORTANT:
	// To watch additional resource types - such as those created by your controller - add gc.Watch* function calls here
	// Watch function calls will transform each object event into a ApiGateway Key to be reconciled by the controller.
	//
	// **********
	// For any new Watched types, you MUST add the appropriate // +kubebuilder:informer and // +kubebuilder:rbac
	// annotations to the ApiGatewayController and run "kubebuilder generate.
	// This will generate the code to start the informers and create the RBAC rules needed for running in a cluster.
	// See:
	// https://godoc.org/github.com/kubernetes-sigs/kubebuilder/pkg/gen/controller#example-package
	// **********
	return gc, nil
}

func (bc ApiGatewayController) AddService(obj interface{}) {
	log.Printf("AddService called for Type %T\n", obj)
	//q.AddRateLimited(eventhandlers.MapToSelf(obj))
	service, ok := obj.(*corev1.Service)
	if !ok {
		log.Printf("Not of Type Service %T\n", obj)
		return

	}
	log.Printf("Service %s was created in Namespace %s\n", service.Name, service.Namespace)

	servicelabelval, ok := service.Labels[ServiceLabel]

	if !ok {
		log.Printf("No Label ServiceLabel in Service, skip processing event for Service: %s\n", service.Name)
		return
	}

	log.Printf("Value of %s: %s\n", ServiceLabel, servicelabelval)

	changedIngresses := bc.addServiceToIngress(service)

	for _, ingress := range changedIngresses {
		bc.KubernetesClientSet.ExtensionsV1beta1().Ingresses(ingress.Namespace).Update(&ingress)
	}
}

func (bc ApiGatewayController) DeleteService(obj interface{}) {
	log.Printf("DeleteService called for Type %T\n", obj)
	//q.AddRateLimited(eventhandlers.MapToSelf(obj))
	service, ok := obj.(*corev1.Service)
	if !ok {
		log.Printf("Not of Type Service %T\n", obj)
		return

	}
	log.Printf("Service %s was deleted in Namespace %s\n", service.Name, service.Namespace)

	servicelabelval, ok := service.Labels[ServiceLabel]

	if !ok {
		log.Printf("No Label ServiceLabel in Service, skip processing event for Service: %s\n", service.Name)
		return
	}

	log.Printf("Value of %s: %s\n", ServiceLabel, servicelabelval)

	changedIngresses := bc.deleteServiceFromIngress(service)

	for _, ingress := range changedIngresses {
		bc.KubernetesClientSet.ExtensionsV1beta1().Ingresses(ingress.Namespace).Update(&ingress)
	}
}

type selector struct {
	ServiceLabelValue string
}

func (n selector) Matches(lbls labels.Labels) bool {
	value := lbls.Get(ServiceLabel)
	if value == "" {
		return false
	}
	if value != n.ServiceLabelValue {
		return false
	}
	return true

}
func (n selector) Empty() bool                                 { return false }
func (n selector) String() string                              { return "" }
func (n selector) Add(_ ...labels.Requirement) labels.Selector { return n }
func (n selector) Requirements() (labels.Requirements, bool)   { return nil, false }
func (n selector) DeepCopySelector() labels.Selector           { return n }

func (bc ApiGatewayController) lookupIngressesForServiceLabel(namespace string, serviceLabelValue string) ([]*v1beta1.Ingress, error) {
	ingresses, err := bc.KubernetesInformers.Extensions().V1beta1().Ingresses().Lister().Ingresses(namespace).List(selector{serviceLabelValue})
	return ingresses, err
}

// LookupFoo looksup a Foo from the lister
func (bc ApiGatewayController) LookupFoo(r types.ReconcileKey) (interface{}, error) {
	return bc.Informers.Apigateway().V1beta1().ApiGateways().Lister().ApiGateways(r.Namespace).Get(r.Name)
}

func (bc *ApiGatewayController) updateApigwStatus(apigw apigatewayv1beta1.ApiGateway, ingress *v1beta1.Ingress) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	apigwCopy := apigw.DeepCopy()
	//apigwCopy.Status.AvailableReplicas = apigw.Status.AvailableReplicas
	// Until #38113 is merged, we must use Update instead of UpdateStatus to
	// update the Status block of the Foo resource. UpdateStatus will not
	// allow changes to the Spec of the resource, which is ideal for ensuring
	// nothing other than resource status has been updated.
	_, err := bc.Clientset.ApigatewayV1beta1().ApiGateways(apigw.Namespace).Update(apigwCopy)
	return err
}

// newDeployment creates a new Deployment for an ApiGateway resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Foo resource that 'owns' it.
// For the new INgress, all relevant services are added
func newIngress(kubernetesInformers informers.SharedInformerFactory, apigw *apigatewayv1beta1.ApiGateway) *v1beta1.Ingress {

	log.Printf("Create new Ingress for apigateway %s", apigw.Name)

	labels := map[string]string{
		"app":        "nginx",
		"controller": apigw.Name,
		ServiceLabel: apigw.Spec.ServiceLabel,
	}

	ingress := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apigw.Spec.IngressName,
			Namespace: apigw.Namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(apigw, schema.GroupVersionKind{
					Group:   apigatewayv1beta1.SchemeGroupVersion.Group,
					Version: apigatewayv1beta1.SchemeGroupVersion.Version,
					Kind:    "ApiGateway",
				}),
			},
		},
	}
	services, err := kubernetesInformers.Core().V1().Services().Lister().Services(apigw.Namespace).List(selector{ServiceLabelValue: apigw.Spec.ServiceLabel})
	log.Printf("Error occurred when listing Services for Label=Value %s=%s: %v", ServiceLabel, apigw.Spec.ServiceLabel, err)

	var rules []v1beta1.IngressRule
	rules = make([]v1beta1.IngressRule, 0)
	if len(services) == 0 {
		rules = append(rules,
			v1beta1.IngressRule{
				"testhost",
				v1beta1.IngressRuleValue{},
			})

	} else {
		for _, service := range services {
			serviceHost := service.Labels[ServiceHostname]
			if serviceHost == "" {
				log.Printf("Service has no Label %s, using default", ServiceHostname)
				serviceHost = "default"
			}

			servicePath := service.Labels[ServicePath]
			if servicePath == "" {
				log.Printf("Service has no Label %s, using default", ServicePath)
				servicePath = "default"
			}

			rules = append(rules,
				v1beta1.IngressRule{
					Host: "testhost",
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								v1beta1.HTTPIngressPath{
									Path: "/" + servicePath,
									Backend: v1beta1.IngressBackend{
										ServiceName: service.Name,
										ServicePort: intstr.IntOrString{
											IntVal: service.Spec.Ports[0].Port,
										},
									},
								},
							},
						},
					},
				})

		}
	}

	ingress.Spec.Rules = rules
	return ingress
}

// newDeployment creates a new Deployment for a Foo resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Foo resource that 'owns' it.
func (bc ApiGatewayController) addServiceToIngress(service *corev1.Service) []v1beta1.Ingress {
	servicelabelval := service.Labels[ServiceLabel]

	ingresses, err := bc.lookupIngressesForServiceLabel(service.Namespace, servicelabelval)

	if err != nil {
		log.Printf("Error getting Ingress for ServiceLabel %s: %v", servicelabelval, err)
		return nil
	}
	log.Printf("# of ingresses matching ServiceLabel %s : %d", servicelabelval, len(ingresses))

	serviceHost := service.Labels[ServiceHostname]
	if serviceHost == "" {
		log.Printf("Service has no Label %s, using default", ServiceHostname)
		serviceHost = "default"
	}

	servicePath := service.Labels[ServicePath]
	if servicePath == "" {
		log.Printf("Service has no Label %s, using default", ServicePath)
		servicePath = "default"
	}

	var newIngresses []v1beta1.Ingress
	for _, ingress := range ingresses {
		newIngress := ingress.DeepCopy()

		rule := v1beta1.IngressRule{
			serviceHost,
			v1beta1.IngressRuleValue{
				HTTP: &v1beta1.HTTPIngressRuleValue{
					Paths: []v1beta1.HTTPIngressPath{
						v1beta1.HTTPIngressPath{
							Path: "/" + servicePath,
							Backend: v1beta1.IngressBackend{
								ServiceName: service.Name,
								ServicePort: intstr.IntOrString{IntVal: service.Spec.Ports[0].Port},
							},
						},
					},
				},
			},
		}
		newrules := append(newIngress.Spec.Rules, rule)
		newIngress.Spec.Rules = newrules

		// Update aufrufen
		newIngresses = append(newIngresses, *newIngress)
	}
	return newIngresses

}

// deleteServiceFromIngress removes a service from the ingresses that have registrerd this service
func (bc ApiGatewayController) deleteServiceFromIngress(service *corev1.Service) []v1beta1.Ingress {
	servicelabelval := service.Labels[ServiceLabel]

	if len(servicelabelval) == 0 {
		log.Printf("Info Service has no service Label and is not used in any ingress %v", servicelabelval)
		return nil
	}

	ingresses, err := bc.lookupIngressesForServiceLabel(service.Namespace, servicelabelval)

	log.Printf("Ingresses  Type of List%T", err)
	if err != nil {
		log.Printf("Error getting Ingress for ServiceLabel %s: %v", servicelabelval, err)
		return nil
	}
	log.Printf("# of ingresses matching ServiceLabel of %d", len(ingresses))

	serviceHost := service.Labels[ServiceHostname]
	if serviceHost == "" {
		log.Printf("Service has no Label %s, using default", ServiceHostname)
		return nil
	}

	servicePath := service.Labels[ServicePath]
	if servicePath == "" {
		log.Printf("Service has no Label %s, using default", ServicePath)
		servicePath = "default"
	}

	var changedIngresses []v1beta1.Ingress
	for _, ingress := range ingresses {
		changed := false
		newIngress := ingress.DeepCopy()
		for _, rule := range newIngress.Spec.Rules {

			if rule.Host != serviceHost {
				continue
			}
			var newIngressPaths []v1beta1.HTTPIngressPath
			for _, path := range rule.HTTP.Paths {
				if path.Path != servicePath {
					newIngressPaths = append(newIngressPaths, path)
				} else {
					changed = true
				}
			}
			if changed {
				rule.HTTP.Paths = newIngressPaths
			}

		}
		if changed {
			changedIngresses = append(changedIngresses, *newIngress)

			log.Printf("Result of INgress Update: %v", err)
		}
	}
	return changedIngresses

}
