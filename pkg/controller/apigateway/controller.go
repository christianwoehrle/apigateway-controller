package apigateway

import (
	"log"

	"github.com/kubernetes-sigs/kubebuilder/pkg/controller"
	"github.com/kubernetes-sigs/kubebuilder/pkg/controller/types"
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
)

// EDIT THIS FILE
// This files was created by "kubebuilder create resource" for you to edit.
// Controller implementation logic for ApiGateway resources goes here.

// Reconcile compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (bc *ApiGatewayController) ReconcileService(k types.ReconcileKey) error {
	log.Printf("-------> Chrissis reconcileService Methode: %v %T %s %s", k, k, k.Namespace, k.Name)
	//apigw, err := bc.apigatewayLister.ApiGateways(k.Namespace).Get(k.Name)
	return nil
}

// Reconcile compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (bc *ApiGatewayController) Reconcile(k types.ReconcileKey) error {
	// INSERT YOUR CODE HERE
	log.Printf("Chrissis reconcile Methode: %v %T %s %s", k, k, k.Namespace, k.Name)
	apigw, err := bc.apigatewayLister.ApiGateways(k.Namespace).Get(k.Name)
	log.Printf("err: %v \n apigw %v", err, apigw)

	apigw, err = bc.Informers.Apigateway().V1beta1().ApiGateways().Lister().ApiGateways(k.Namespace).Get(k.Name)
	if apigw == nil {
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

	log.Printf("err: %v \n apigw %v", err, apigw)
	log.Printf("Implement the Reconcile function on apigateway.ApiGatewayController to reconcile %s\n", k.Name)

	// Get the deployment with the name specified in Foo.spec
	ingress, err := bc.KubernetesInformers.Extensions().V1beta1().Ingresses().Lister().Ingresses(k.Namespace).Get(k.Name)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		ingress, err = bc.KubernetesClientSet.ExtensionsV1beta1().Ingresses(k.Namespace).Create(newIngress(apigw))
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// If the Deployment is not controlled by this Foo resource, we should log
	// a warning to the event recorder and ret
	if !metav1.IsControlledBy(ingress, apigw) {
		msg := fmt.Sprintf(MessageResourceExists, ingress.Name)
		bc.apigatewayrecorder.Event(apigw, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}

	// If this number of the replicas on the Foo resource is specified, and the
	// number does not equal the current desired replicas on the Deployment, we
	// should update the Deployment resource.

	ingressLabel := ingress.Labels["sericeLabel"]
	if apigw.Spec.ServiceLabel != "" && apigw.Spec.ServiceLabel != ingressLabel {
		glog.V(4).Infof("Foo %s replicas: %d, ServiceLabel: %d", apigw.Name, apigw.Spec.ServiceLabel, ingressLabel)
		ingress, err = bc.KubernetesClientSet.ExtensionsV1beta1().Ingresses(apigw.Namespace).Update(newIngress(apigw))
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

// +kubebuilder:controller:group=apigateway,version=v1beta1,kind=ApiGateway,resource=apigateways
// +kubebuilder:informers:group=core,version=v1,kind=Service
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

	log.Printf("watch apigateway")
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
					fmt.Printf("Service Updated %v\n %T", obj, obj)
					//q.AddRateLimited(eventhandlers.MapToSelf(obj))
				},
				DeleteFunc: func(obj interface{}) {
					fmt.Printf("Service Deleted %v\n %T", obj, obj)
					//q.AddRateLimited(eventhandlers.MapToSelf(obj))
				},
			}
		})
	if err != nil {
		log.Fatalf("%v", err)
	}

	log.Printf("watch services")
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
	fmt.Printf("Service 1 Added %v \n %T\n", obj, obj)
	//q.AddRateLimited(eventhandlers.MapToSelf(obj))
	service, ok := obj.(corev1.Service)
	if !ok {
		fmt.Println("Not of Type Service %T", obj)
		return

	}
	val, ok := service.Labels[ServiceLabel]
	if !ok {
		return
	}
	fmt.Printf("Value of %s: %s\n", ServiceLabel, val)

	ingresses, ret := bc.KubernetesInformers.Extensions().V1beta1().Ingresses().Lister().Ingresses(service.Namespace).List(selector{})

	fmt.Printf("Return Type of List%T", ret)
	if ret != nil {
		return
	}
	fmt.Printf("# of ingresses matching %n", len(ingresses))
}

type selector struct{}

func (n selector) Matches(lbls labels.Labels) bool {
	return lbls.Has(ServiceLabel)
}
func (n selector) Empty() bool                                 { return false }
func (n selector) String() string                              { return "" }
func (n selector) Add(_ ...labels.Requirement) labels.Selector { return n }
func (n selector) Requirements() (labels.Requirements, bool)   { return nil, false }
func (n selector) DeepCopySelector() labels.Selector           { return n }

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

// newDeployment creates a new Deployment for a Foo resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Foo resource that 'owns' it.
func newIngress(apigw *apigatewayv1beta1.ApiGateway) *v1beta1.Ingress {
	labels := map[string]string{
		"app":        "nginx",
		"controller": apigw.Name,
		ServiceLabel: apigw.Spec.ServiceLabel,
	}
	return &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apigw.Name,
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
		//		Spec: v1beta1.Ingress{
		//
		//		},
	}
}
