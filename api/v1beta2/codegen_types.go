/*
Copyright 2022 APIMatic.io.

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

package v1beta2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CodeGenSpec defines the desired state of CodeGen
type CodeGenSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// replicas is the desired number of instances of CodeGen. Minimum is 0. Defaults to 1 if not provided
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=1
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors="urn:alm:descriptor:com.tectonic.ui:podCount"
	Replicas int32 `json:"replicas,omitempty"`

	// labels contains the desired key-value pair of selectors to apply to generated Services and Deployment pods. If not given, default label selectors will be generated. Can not be updated.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinProperties=1
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	Labels map[string]string `json:"labels,omitempty"`

	// CodeGenPodSpec contains desired configuration for created CodeGen pods
	// +kubebuilder:validation:Optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	PodSpec *CodeGenPodSpec `json:"podSpec,omitempty"`

	// CodeGenLicenseSpec contains desired configuration for license associated with created APIMatic CodeGen pods
	// +kubebuilder:validation:Required
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	LicenseSpec CodeGenLicenseSpec `json:"licenseSpec"`

	// CodeGenServiceSpec contains desired configuration for the service that exposes the APIMatic pods
	// +kubebuilder:validation:Optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	ServiceSpec *CodeGenServiceSpec `json:"serviceSpec,omitempty"`

	// CodeGenPodPlacementSpec configures the desired CodeGen pod scheduling policy
	// +kubebuilder:validation:Optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	CodeGenPodPlacementSpec *CodeGenPodPlacementSpec `json:"podPlacementSpec,omitempty"`
}

// CodeGenStatus defines the observed state of CodeGen
type CodeGenStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// replicas is the actual number of instances of CodeGen.
	// +kubebuilder:validation:Optional
	//+operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:com.tectonic.ui:number"
	Replicas int32 `json:"replicas,omitempty"`
}

type CodeGenContainerSpec struct {
	// APIMatic CodeGen container image. Defaults to RedHat certified image
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Image *string `json:"image,omitempty"`

	// Image pull policy. One of Always, Never, IfNotPresent. Defaults to Always if :latest tag is specified, or IfNotPresent otherwise. Cannot be updated.More info: https://kubernetes.io/docs/concepts/containers/images#updating-images
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=Always;Never;IfNotPresent
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors="urn:alm:descriptor:com.tectonic.ui:imagePullPolicy"
	ImagePullPolicy *corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// ImagePullSecret is an optional reference to a secret in the same namespace to use for pulling the APIMatic CodeGen container image.
	// If specified, this secrets will be passed to the puller implementation to use. For example,
	// in the case of docker, only DockerConfig type secrets are honored.
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	ImagePullSecret *string `json:"imagePullSecret,omitempty"`

	// Resource Requirements represents the compute resource requirements of the APIMatic CodeGen container
	// +kubebuilder:validation:Optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Resource Requirements",xDescriptors="urn:alm:descriptor:com.tectonic.ui:resourceRequirements"
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

type CodeGenPodSpec struct {

	// APIMaticContainerSpec defines the configurations used for the APIMatic CodeGen container
	// +kubebuilder:validation:Optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	CodeGenContainerSpec *CodeGenContainerSpec `json:"containerSpec,omitempty"`

	// Optional duration in seconds the pod needs to terminate gracefully. May be decreased in delete request.
	// Value must be non-negative integer. The value zero indicates delete immediately.
	// If this value is nil, the default grace period of 30 seconds will be used instead.
	// The grace period is the duration in seconds after the processes running in the pod are sent
	// a termination signal and the time when the processes are forcibly halted with a kill signal.
	// Set this value longer than the expected cleanup time for your process.
	// Defaults to 30 seconds.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=30
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors="urn:alm:descriptor:com.tectonic.ui:number"
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`

	// ServiceAccountName is the name of the ServiceAccount to use to run CodeGen pod. It must be a valid RFC 1123 name with a maximum of 253 characters,
	// contain only lower case characters, '-' or '.'. It should start and end with an alphanumeric character
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	ServiceAccountName *string `json:"serviceAccountName,omitempty"`

	// SecurityContext holds pod-level security attributes and common container settings.
	// Optional: Defaults to empty.  See type description for default values of each field.
	// +kubebuilder:validation:Optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`

	// If specified, the pod will be dispatched by specified scheduler.
	// If not specified, the pod will be dispatched by default scheduler.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	SchedulerName *string `json:"schedulerName,omitempty"`

	// If specified, indicates the pod's priority. "system-node-critical" and
	// "system-cluster-critical" are two special keywords which indicate the
	// highest priorities with the former being the highest priority. Any other
	// name must be defined by creating a PriorityClass object with that name.
	// If not specified, the pod priority will be default or zero if there is no
	// default. It must be a valid RFC 1123 name with a maximum of 253 characters,
	// contain only lower case characters, '-' or '.'. It should start and end with an alphanumeric character
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	PriorityClassName *string `json:"priorityClassName,omitempty"`
}

type CodeGenLicenseSpec struct {

	// The type of resource that includes the APIMatic CodeGen license file information. Valid values are ConfigMap, LicenseBlob and ConfigSecret.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=ConfigMap;ConfigSecret;LicenseBlob
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors={"urn:alm:descriptor:com.tectonic.ui:select:ConfigMap","urn:alm:descriptor:com.tectonic.ui:select:ConfigSecret","urn:alm:descriptor:com.tectonic.ui:select:LicenseBlob"}
	CodeGenLicenseSourceType string `json:"licenseSourceType"`

	// The name of the resource that includes the APIMatic CodeGen license file information if licenseSourceType is ConfigMap or ConfigSecret, or the base64 encoded License string if licenseSourceType is LicenseBlob.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	CodeGenLicenseSourceValue string `json:"licenseSourceValue"`
}

type CodeGenServiceSpec struct {

	// Type string describes ingress methods for a service. Valid values are ClusterIP, NodePort, LoadBalancer. Defaults to ClusterIP
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=ClusterIP
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors={"urn:alm:descriptor:com.tectonic.ui:select:ClusterIP","urn:alm:descriptor:com.tectonic.ui:select:NodePort","urn:alm:descriptor:com.tectonic.ui:select:LoadBalancer"}
	Type corev1.ServiceType `json:"serviceType,omitempty"`

	// clusterIP is the IP address of the service and is usually assigned
	// randomly. If an address is specified manually, is in-range (as per
	// system configuration), and is not in use, it will be allocated to the
	// service; otherwise creation of the service will fail. This field may not
	// be changed through updates. Valid values are "None" or a valid IP address. Setting this to "None" makes a
	// "headless service" (no virtual IP), which is useful when direct endpoint
	// connections are preferred and proxying is not required. Setting this value to "None" for
	// service of type LoadBalancer or NodePort will result in failure.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	ClusterIP *string `json:"clusterIP,omitempty"`

	// externalTrafficPolicy denotes if this Service desires to route external
	// traffic to node-local or cluster-wide endpoints. "Local" preserves the
	// client source IP and avoids a second hop for LoadBalancer and Nodeport
	// type services, but risks potentially imbalanced traffic spreading.
	// "Cluster" obscures the client source IP and may cause a second hop to
	// another node, but should have good overall load-spreading. If not defined for LoadBalancer or Nodeport type, defaults to Cluster.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=Local;Cluster
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors={"urn:alm:descriptor:com.tectonic.ui:select:Local","urn:alm:descriptor:com.tectonic.ui:select:Cluster"}
	ExternalTrafficPolicy *corev1.ServiceExternalTrafficPolicyType `json:"externalTrafficPolicy,omitempty"`

	//CodeGen Service Port specifies how the APIMatic CodeGen service is exposed within the pod
	// +kubebuilder:validation:Optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	CodeGenServicePort *CodeGenServicePort `json:"servicePort,omitempty"`

	// Only applies to Service Type: LoadBalancer
	// LoadBalancer will get created with the IP specified in this field.
	// This feature depends on whether the underlying cloud-provider supports specifying
	// the loadBalancerIP when a load balancer is created.
	// This field will be ignored if the cloud-provider does not support the feature.
	// Creation will fail if service type is not LoadBalancer
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	LoadBalancerIP *string `json:"loadBalancerIP,omitempty"`

	//Custom annotations to apply to services. This is recommended use instead of LoadBalancerIP when dealing with cloud-provided load balancers
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinProperties=1
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	ServiceCustomAnnotations map[string]string `json:"serviceCustomAnnotations,omitempty"`

	//loadBalancerClass is the class of the load balancer implementation this Service belongs to. If specified, the value of this field must be a label-style identifier, with an optional prefix, e.g. "internal-vip" or "example.com/internal-vip". Unprefixed names are reserved for end-users. This field can only be set when the Service type is 'LoadBalancer'. If not set, the default load balancer implementation is used, today this is typically done through the cloud provider integration, but should apply for any default implementation. If set, it is assumed that a load balancer implementation is watching for Services with a matching class. Any default load balancer implementation (e.g. cloud providers) should ignore Services that set this field. This field can only be set when creating or updating a Service to type 'LoadBalancer'. Once set, it can not be changed. This field will be wiped when a service is updated to a non 'LoadBalancer' type.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	LoadBalancerClass *string `json:"loadBalancerClass,omitempty"`
}

// CodeGenServicePort configures the APIMatic CodeGen container ports exposed by the service
type CodeGenServicePort struct {
	// The name of the APIMatic CodeGen service port within the service. This must be a DNS_LABEL. Defaults to codegen
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=codegen
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=15
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Name string `json:"name,omitempty"`

	// The port on each node on which this service is exposed when type is
	// NodePort or LoadBalancer.  Usually assigned by the system. If a value is
	// specified, in-range, and not in use it will be used, otherwise the
	// operation will fail.  If not specified, a port will be allocated if this
	// Service requires one.  If this field is specified when creating a
	// Service which does not need it, creation will fail. This field will be
	// wiped when updating a Service to no longer need it (e.g. changing type
	// from NodePort to ClusterIP).
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=30000
	// +kubebuilder:validation:Maximum=32767
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors="urn:alm:descriptor:com.tectonic.ui:number"
	NodePort *int32 `json:"nodePort,omitempty"`
}

// +kubebuilder:validation:MinProperties=1
type CodeGenPodPlacementSpec struct {
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinProperties=1
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// NodeName is a request to schedule this pod onto a specific node. If it is non-empty,
	// the scheduler simply schedules this pod onto that node, assuming that it fits resource
	// requirements. It must be a valid RFC 1123 name with a maximum of 253 characters,
	// contain only lower case characters, '-' or '.'. It should start and end with an alphanumeric character
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	NodeName *string `json:"nodeName,omitempty"`

	// If specified, the pod's tolerations.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinItems=1
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Describes node affinity scheduling rules for the pod.
	// +kubebuilder:validation:Optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors="urn:alm:descriptor:com.tectonic.ui:nodeAffinity"
	NodeAffinity *corev1.NodeAffinity `json:"nodeAffinity,omitempty"`
	// Describes pod affinity scheduling rules (e.g. co-locate this pod in the same node, zone, etc. as some other pod(s)).
	// +kubebuilder:validation:Optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors="urn:alm:descriptor:com.tectonic.ui:podAffinity"
	PodAffinity *corev1.PodAffinity `json:"podAffinity,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=cgn
//+kubebuilder:subresource:status
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas

// CodeGen is the Schema for the codegens API
//+operator-sdk:csv:customresourcedefinitions:displayName="APIMatic CodeGen App",resources={{Service,v1,codegen-service},{Deployment,apps/v1,codegen-deployment}}
type CodeGen struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CodeGenSpec   `json:"spec,omitempty"`
	Status CodeGenStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CodeGenList contains a list of APIMatic
type CodeGenList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CodeGen `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CodeGen{}, &CodeGenList{})
}
