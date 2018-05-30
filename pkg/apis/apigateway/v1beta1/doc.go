// Api versions allow the api contract for a resource to be changed while keeping
// backward compatibility by support multiple concurrent versions
// of the same resource

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package,register
// +k8s:conversion-gen=github.com/christianwoehrle/apigateway-controller/pkg/apis/apigateway
// +k8s:defaulter-gen=TypeMeta
// +groupName=apigateway.cw.com
package v1beta1 // import "github.com/christianwoehrle/apigateway-controller/pkg/apis/apigateway/v1beta1"
