package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// EDIT THIS FILE!
// Created by "kubebuilder create resource" for you to implement the ApiGateway resource schema definition
// as a go struct.
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ApiGatewaySpec defines the desired state of ApiGateway
type ApiGatewaySpec struct {
	ServiceLabel string `json:"serviceLabel"`
	IngressName  string
	Host         string
	Backend      Backend
}

// ApiGatewayStatus defines the observed state of ApiGateway
type ApiGatewayStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "kubebuilder generate" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ApiGateway
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=apigateways
type ApiGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApiGatewaySpec   `json:"spec,omitempty"`
	Status ApiGatewayStatus `json:"status,omitempty"`
}


type Backend struct {
	ServiceName string
	ServicePort intstr.IntOrString
}

