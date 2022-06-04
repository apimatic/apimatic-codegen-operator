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
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	unversionedvalidation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
)

const (
	DEFAULTIMAGENAME   string = "registry.connect.redhat.com/apimatic/apimatic-codegen-ubi8:v1.8.3"
	DEFAULTLABELKEY1   string = "app"
	DEFAULTLABELVALUE1 string = "apimatic-codegen"
	DEFAULTLABELKEY2   string = "codegen_cr"

	SysctlSegmentFmt      string = "[a-z0-9]([-_a-z0-9]*[a-z0-9])?"
	SysctlFmt             string = "(" + SysctlSegmentFmt + "\\.)*" + SysctlSegmentFmt
	SysctlContainSlashFmt string = "(" + SysctlSegmentFmt + "[\\./])*" + SysctlSegmentFmt
	SysctlMaxLength       int    = 253

	isNegativeErrorMsg   string = `must be greater than or equal to 0`
	isNotIntegerErrorMsg string = `must be an integer`
)

var (
	validateNodeName            = apimachineryvalidation.NameIsDNSSubdomain
	nodeFieldSelectorValidators = map[string]func(string, bool) []string{
		metav1.ObjectNameField: validateNodeName,
	}
	sysctlRegexp               = regexp.MustCompile("^" + SysctlFmt + "$")
	sysctlContainSlashRegexp   = regexp.MustCompile("^" + SysctlContainSlashFmt + "$")
	validFSGroupChangePolicies = sets.NewString(string(corev1.FSGroupChangeOnRootMismatch), string(corev1.FSGroupChangeAlways))

	integerResources = sets.NewString(
		string(corev1.ResourcePods),
		string(corev1.ResourceQuotas),
		string(corev1.ResourceServices),
		string(corev1.ResourceReplicationControllers),
		string(corev1.ResourceSecrets),
		string(corev1.ResourceConfigMaps),
		string(corev1.ResourcePersistentVolumeClaims),
		string(corev1.ResourceServicesNodePorts),
		string(corev1.ResourceServicesLoadBalancers),
	)
	standardResources = sets.NewString(
		string(corev1.ResourceCPU),
		string(corev1.ResourceMemory),
		string(corev1.ResourceEphemeralStorage),
		string(corev1.ResourceRequestsCPU),
		string(corev1.ResourceRequestsMemory),
		string(corev1.ResourceRequestsEphemeralStorage),
		string(corev1.ResourceLimitsCPU),
		string(corev1.ResourceLimitsMemory),
		string(corev1.ResourceLimitsEphemeralStorage),
		string(corev1.ResourcePods),
		string(corev1.ResourceQuotas),
		string(corev1.ResourceServices),
		string(corev1.ResourceReplicationControllers),
		string(corev1.ResourceSecrets),
		string(corev1.ResourceConfigMaps),
		string(corev1.ResourcePersistentVolumeClaims),
		string(corev1.ResourceStorage),
		string(corev1.ResourceRequestsStorage),
		string(corev1.ResourceServicesNodePorts),
		string(corev1.ResourceServicesLoadBalancers),
	)
	standardContainerResources = sets.NewString(
		string(corev1.ResourceCPU),
		string(corev1.ResourceMemory),
		string(corev1.ResourceEphemeralStorage),
	)
)

type podValidationOptions struct {
	// Allow pod spec to use hugepages in downward API
	AllowDownwardAPIHugePages bool
	// Allow invalid pod-deletion-cost annotation value for backward compatibility.
	AllowInvalidPodDeletionCost bool
	// Allow pod spec to use non-integer multiple of huge page unit size
	AllowIndivisibleHugePagesValues bool
	// Allow hostProcess field to be set in windows security context
	AllowWindowsHostProcessField bool
	// Allow more DNSSearchPaths and longer DNSSearchListChars
	AllowExpandedDNSConfig bool
	// Allow OSField to be set in the pod spec
	AllowOSField bool
	// Allow sysctl name to contain a slash
	AllowSysctlRegexContainSlash bool
}

// log is for logging in this package.
var apimaticlog = logf.Log.WithName("codegen-resource")

func (r *CodeGen) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-apps-apimatic-io-v1beta2-codegen,mutating=true,failurePolicy=fail,sideEffects=None,groups=apps.apimatic.io,resources=codegens,verbs=create;update,versions=v1beta2,name=mcodegen.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &CodeGen{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *CodeGen) Default() {
	apimaticlog.Info("default", "name", r.Name, "namespace", r.Namespace)

	r.defaultPodSpec()

	r.defaultServiceSpec()

}

func (r *CodeGen) defaultPodSpec() {

	if r.Spec.PodSpec == nil {
		apimaticlog.Info("default", "name", r.Name, "namespace", r.Namespace, "message", "Adding default podspec")
		var podSpec CodeGenPodSpec = CodeGenPodSpec{}
		r.Spec.PodSpec = &podSpec
	}

	if r.Spec.PodSpec.CodeGenContainerSpec == nil {
		apimaticlog.Info("default", "name", r.Name, "namespace", r.Namespace, "message", "Adding default containerSpec")
		var containerSpec CodeGenContainerSpec = CodeGenContainerSpec{}
		r.Spec.PodSpec.CodeGenContainerSpec = &containerSpec
	}

	if r.Spec.PodSpec.CodeGenContainerSpec.Image == nil {
		var defaultName string = DEFAULTIMAGENAME
		if envVarImageName, envSet := os.LookupEnv("DEFAULTCODEGENIMAGE"); envSet {
			defaultName = envVarImageName
		}
		apimaticlog.Info("default", "name", r.Name, "namespace", r.Namespace, "message", "Adding default image "+defaultName)
		r.Spec.PodSpec.CodeGenContainerSpec.Image = &defaultName
	}

}

func (r *CodeGen) defaultServiceSpec() {

	if r.Spec.ServiceSpec == nil {
		apimaticlog.Info("default", "name", r.Name, "namespace", r.Namespace, "message", "Adding default serviceSpec")
		var serviceSpec CodeGenServiceSpec = CodeGenServiceSpec{}
		r.Spec.ServiceSpec = &serviceSpec
	}

	if r.Spec.Labels == nil {
		apimaticlog.Info("default", "name", r.Name, "namespace", r.Namespace, "message", "adding default labels")
		labelsForAPIMatic := make(map[string]string)
		labelsForAPIMatic[DEFAULTLABELKEY1] = DEFAULTLABELVALUE1
		labelsForAPIMatic[DEFAULTLABELKEY2] = r.Name
		r.Spec.Labels = labelsForAPIMatic
	}

	if r.Spec.ServiceSpec.CodeGenServicePort == nil {
		apimaticlog.Info("default", "name", r.Name, "namespace", r.Namespace, "message", "Adding default servicePort")
		var apimaticServicePort CodeGenServicePort = CodeGenServicePort{}
		r.Spec.ServiceSpec.CodeGenServicePort = &apimaticServicePort
	}

	var serviceTypeNodePort bool = strings.Compare(string(r.Spec.ServiceSpec.Type), "NodePort") == 0
	var serviceTypeLoadBalancer bool = strings.Compare(string(r.Spec.ServiceSpec.Type), "LoadBalancer") == 0

	if r.Spec.ServiceSpec.ExternalTrafficPolicy == nil && (serviceTypeNodePort || serviceTypeLoadBalancer) {
		apimaticlog.Info("default", "name", r.Name, "namespace", r.Namespace, "message", "Adding default externalTrafficPolicy for service "+corev1.ServiceExternalTrafficPolicyTypeCluster)
		var trafficPolicy corev1.ServiceExternalTrafficPolicyType = corev1.ServiceExternalTrafficPolicyTypeCluster
		r.Spec.ServiceSpec.ExternalTrafficPolicy = &trafficPolicy
	}

}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-apps-apimatic-io-v1beta2-codegen,mutating=false,failurePolicy=fail,sideEffects=None,groups=apps.apimatic.io,resources=codegens,verbs=create;update,versions=v1beta2,name=vcodegen.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &CodeGen{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CodeGen) ValidateCreate() error {
	apimaticlog.Info("validate create", "name", r.Name, "namespace", r.Namespace)
	allErrors := r.validateCodeGen()
	if allErrors != nil {
		return apierrors.NewInvalid(schema.GroupKind{Group: "apps.apimatic.io", Kind: "CodeGen"}, r.Name, *allErrors)
	}
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CodeGen) ValidateUpdate(old runtime.Object) error {
	apimaticlog.Info("validate update", "name", r.Name, "namespace", r.Namespace)

	oldAPIMatic, ok := old.(*CodeGen)

	var allErrors field.ErrorList

	if !ok {
		return fmt.Errorf("expect old object to a %T instead of %T", oldAPIMatic, old)
	}

	if clusterIPChanged(r, oldAPIMatic) {
		err := field.Forbidden(field.NewPath("spec").Child("serviceSpec").Child("clusterIP"), "clusterIP field can not be updated")
		allErrors = append(allErrors, err)
	}

	if loadBalancerClassNameChanged(r, oldAPIMatic) {
		err := field.Forbidden(field.NewPath("spec").Child("serviceSpec").Child("loadBalancerClass"), "loadBalancerClass field can not be updated")
		allErrors = append(allErrors, err)
	}

	if labelsChanged(r, oldAPIMatic) {
		err := field.Forbidden(field.NewPath("spec").Child("labels"), "labels field can not be changed")
		allErrors = append(allErrors, err)
	}

	otherErrors := r.validateCodeGen()
	if otherErrors != nil {
		allErrors = append(allErrors, *otherErrors...)
	}

	if len(allErrors) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "apps.apimatic.io", Kind: "CodeGen"}, r.Name, allErrors)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CodeGen) ValidateDelete() error {
	apimaticlog.Info("validate delete", "name", r.Name, "namespace", r.Namespace)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *CodeGen) validateCodeGen() *field.ErrorList {
	var allErrors field.ErrorList

	if podSpecErrors := r.validatePodSpec(); podSpecErrors != nil {
		allErrors = append(allErrors, *podSpecErrors...)
	}
	if serviceSpecErrors := r.validateServiceSpec(); serviceSpecErrors != nil {
		allErrors = append(allErrors, *serviceSpecErrors...)
	}

	if podPlacementSpecErrors := r.validatePodPlacementSpec(); podPlacementSpecErrors != nil {
		allErrors = append(allErrors, *podPlacementSpecErrors...)
	}

	if len(allErrors) == 0 {
		return nil
	}

	return &allErrors
}

func (r *CodeGen) validateServiceSpec() *field.ErrorList {
	var allErrors field.ErrorList

	if r.Spec.ServiceSpec.ClusterIP != nil &&
		strings.Compare(*r.Spec.ServiceSpec.ClusterIP, "None") != 0 {
		if net.ParseIP(*r.Spec.ServiceSpec.ClusterIP) == nil {
			err := field.Invalid(field.NewPath("spec").Child("serviceSpec").Child("clusterIP"), string(*r.Spec.ServiceSpec.ClusterIP), "clusterIP is not a valid IP address")
			allErrors = append(allErrors, err)
		}
	}

	if annotations := r.Spec.ServiceSpec.ServiceCustomAnnotations; annotations != nil {
		if errorList := apimachineryvalidation.ValidateAnnotations(annotations, field.NewPath("spec").Child("serviceSpec").Child("serviceCustomAnnotations")); len(errorList) != 0 {
			allErrors = append(allErrors, errorList...)
		}
	}

	var serviceTypeNodePort bool = strings.Compare(string(r.Spec.ServiceSpec.Type), string(corev1.ServiceTypeNodePort)) == 0
	var serviceTypeLoadBalancer bool = strings.Compare(string(r.Spec.ServiceSpec.Type), string(corev1.ServiceTypeLoadBalancer)) == 0

	if r.Spec.ServiceSpec.ClusterIP != nil &&
		strings.Compare(*r.Spec.ServiceSpec.ClusterIP, "None") == 0 &&
		(serviceTypeLoadBalancer || serviceTypeNodePort) {
		err := field.Invalid(field.NewPath("Spec").Child("serviceSpec").Child("clusterIP"), string(*r.Spec.ServiceSpec.ClusterIP), "clusterIP can not be set to None for NodePort and LoadBalancer services")
		allErrors = append(allErrors, err)
	}

	if r.Spec.ServiceSpec.LoadBalancerIP != nil {
		if net.ParseIP(*r.Spec.ServiceSpec.LoadBalancerIP) == nil {
			err := field.Invalid(field.NewPath("spec").Child("serviceSpec").Child("loadBalancerIP"), string(*r.Spec.ServiceSpec.LoadBalancerIP), "loadBalancerIP is not a valid IP address")
			allErrors = append(allErrors, err)
		}
	}

	if r.Spec.ServiceSpec.LoadBalancerIP != nil && !serviceTypeLoadBalancer {
		err := field.Forbidden(field.NewPath("spec").Child("serviceSpec").Child("loadBalancerIP"), "loadBalancerIP can not be set for service not of type LoadBalancer")
		allErrors = append(allErrors, err)
	}

	if r.Spec.ServiceSpec.LoadBalancerClass != nil && !serviceTypeLoadBalancer {
		err := field.Forbidden(field.NewPath("spec").Child("serviceSpec").Child("loadBalancerClass"), "loadBalancerClass can not be set for service not of type LoadBalancer")
		allErrors = append(allErrors, err)
	}

	if r.Spec.ServiceSpec.ExternalTrafficPolicy != nil && !(serviceTypeLoadBalancer || serviceTypeNodePort) {
		err := field.Forbidden(field.NewPath("spec").Child("serviceSpec").Child("externalTrafficPolicy"), "externalTrafficPolicy can not be set for service not of type LoadBalancer or NodePort")
		allErrors = append(allErrors, err)
	}

	if r.Spec.ServiceSpec.CodeGenServicePort.NodePort != nil && !(serviceTypeNodePort || serviceTypeLoadBalancer) {
		err := field.Forbidden(field.NewPath("spec").Child("serviceSpec").Child("servicePort").Child("nodePort"), "nodePort can not be set for service of type ClusterIP")
		allErrors = append(allErrors, err)
	}

	if len(allErrors) == 0 {
		return nil
	}

	return &allErrors
}

func clusterIPChanged(old *CodeGen, new *CodeGen) bool {
	return !reflect.DeepEqual(old.Spec.ServiceSpec.ClusterIP, new.Spec.ServiceSpec.ClusterIP)
}

func loadBalancerClassNameChanged(old *CodeGen, new *CodeGen) bool {
	return strings.Compare(string(old.Spec.ServiceSpec.Type), string(corev1.ServiceTypeLoadBalancer)) == 0 &&
		strings.Compare(string(new.Spec.ServiceSpec.Type), string(corev1.ServiceTypeLoadBalancer)) == 0 &&
		!reflect.DeepEqual(old.Spec.ServiceSpec.LoadBalancerClass, new.Spec.ServiceSpec.LoadBalancerClass)
}

func labelsChanged(old *CodeGen, new *CodeGen) bool {
	return !reflect.DeepEqual(old.Spec.Labels, new.Spec.Labels)
}

func (r *CodeGen) validatePodPlacementSpec() *field.ErrorList {

	if r.Spec.CodeGenPodPlacementSpec == nil {
		return nil
	}
	var allErrors field.ErrorList

	if nodeAffinity := r.Spec.CodeGenPodPlacementSpec.NodeAffinity; nodeAffinity != nil {
		if errorList := validateNodeAffinity(nodeAffinity, field.NewPath("spec").Child("podPlacementSpec").Child("nodeAffinity")); len(errorList) != 0 {
			allErrors = append(allErrors, errorList...)
		}
	}

	if podAffinity := r.Spec.CodeGenPodPlacementSpec.PodAffinity; podAffinity != nil {
		if errorList := validatePodAffinity(podAffinity, field.NewPath("spec").Child("podPlacementSpec").Child("tolerations")); len(errorList) != 0 {
			allErrors = append(allErrors, errorList...)
		}
	}

	if tolerations := r.Spec.CodeGenPodPlacementSpec.Tolerations; tolerations != nil {
		if errorList := validateTolerations(tolerations, field.NewPath("spec").Child("podPlacementSpec").Child("podAffinity")); len(errorList) != 0 {
			allErrors = append(allErrors, errorList...)
		}
	}

	if nodeSelector := r.Spec.CodeGenPodPlacementSpec.NodeSelector; nodeSelector != nil {
		if errorList := unversionedvalidation.ValidateLabels(nodeSelector, field.NewPath("spec").Child("podPlacementSpec").Child("nodeSelector")); len(errorList) != 0 {
			allErrors = append(allErrors, errorList...)
		}
	}

	if len(allErrors) == 0 {
		return nil
	}

	return &allErrors
}

func validatePodAffinity(podAffinity *corev1.PodAffinity, fldPath *field.Path) field.ErrorList {
	allErrors := field.ErrorList{}
	if podAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		allErrors = append(allErrors, validatePodAffinityTerms(podAffinity.RequiredDuringSchedulingIgnoredDuringExecution, fldPath.Child("requiredDuringSchedulingIgnoredDuringExecution"))...)
	}
	if podAffinity.PreferredDuringSchedulingIgnoredDuringExecution != nil {
		allErrors = append(allErrors, validateWeightedPodAffinityTerms(podAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
			fldPath.Child("preferredDuringSchedulingIgnoredDuringExecution"))...)
	}
	return allErrors
}

func validatePodAffinityTerms(podAffinityTerms []corev1.PodAffinityTerm, fldPath *field.Path) field.ErrorList {
	allErrors := field.ErrorList{}
	for i, podAffinityTerm := range podAffinityTerms {
		allErrors = append(allErrors, validatePodAffinityTerm(podAffinityTerm, fldPath.Index(i))...)
	}

	return allErrors
}

func validateWeightedPodAffinityTerms(weightedPodAffinityTerms []corev1.WeightedPodAffinityTerm, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	for j, weightedTerm := range weightedPodAffinityTerms {
		if weightedTerm.Weight <= 0 || weightedTerm.Weight > 100 {
			allErrs = append(allErrs, field.Invalid(fldPath.Index(j).Child("weight"), weightedTerm.Weight, "must be in the range 1-100"))
		}
		allErrs = append(allErrs, validatePodAffinityTerm(weightedTerm.PodAffinityTerm, fldPath.Index(j).Child("podAffinityTerm"))...)
	}
	return allErrs
}

func validatePodAffinityTerm(podAffinityTerm corev1.PodAffinityTerm, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, unversionedvalidation.ValidateLabelSelector(podAffinityTerm.LabelSelector, fldPath.Child("labelSelector"))...)
	allErrs = append(allErrs, unversionedvalidation.ValidateLabelSelector(podAffinityTerm.NamespaceSelector, fldPath.Child("namespaceSelector"))...)

	for _, name := range podAffinityTerm.Namespaces {
		for _, msg := range apimachineryvalidation.ValidateNamespaceName(name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("namespace"), name, msg))
		}
	}
	if len(podAffinityTerm.TopologyKey) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("topologyKey"), "can not be empty"))
	}
	return append(allErrs, unversionedvalidation.ValidateLabelName(podAffinityTerm.TopologyKey, fldPath.Child("topologyKey"))...)
}

func validateNodeAffinity(na *corev1.NodeAffinity, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if na.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		allErrs = append(allErrs, validateNodeSelector(na.RequiredDuringSchedulingIgnoredDuringExecution, fldPath.Child("requiredDuringSchedulingIgnoredDuringExecution"))...)
	}
	if len(na.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
		allErrs = append(allErrs, validatePreferredSchedulingTerms(na.PreferredDuringSchedulingIgnoredDuringExecution, fldPath.Child("preferredDuringSchedulingIgnoredDuringExecution"))...)
	}
	return allErrs
}

func validateNodeSelector(nodeSelector *corev1.NodeSelector, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	termFldPath := fldPath.Child("nodeSelectorTerms")
	if len(nodeSelector.NodeSelectorTerms) == 0 {
		return append(allErrs, field.Required(termFldPath, "must have at least one node selector term"))
	}

	for i, term := range nodeSelector.NodeSelectorTerms {
		allErrs = append(allErrs, validateNodeSelectorTerm(term, termFldPath.Index(i))...)
	}

	return allErrs
}

func validatePreferredSchedulingTerms(terms []corev1.PreferredSchedulingTerm, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	for i, term := range terms {
		if term.Weight <= 0 || term.Weight > 100 {
			allErrs = append(allErrs, field.Invalid(fldPath.Index(i).Child("weight"), term.Weight, "must be in the range 1-100"))
		}

		allErrs = append(allErrs, validateNodeSelectorTerm(term.Preference, fldPath.Index(i).Child("preference"))...)
	}
	return allErrs
}

func validateNodeSelectorTerm(term corev1.NodeSelectorTerm, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	for j, req := range term.MatchExpressions {
		allErrs = append(allErrs, validateNodeSelectorRequirement(req, fldPath.Child("matchExpressions").Index(j))...)
	}

	for j, req := range term.MatchFields {
		allErrs = append(allErrs, validateNodeFieldSelectorRequirement(req, fldPath.Child("matchFields").Index(j))...)
	}

	return allErrs
}

func validateNodeSelectorRequirement(rq corev1.NodeSelectorRequirement, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	switch rq.Operator {
	case corev1.NodeSelectorOpIn, corev1.NodeSelectorOpNotIn:
		if len(rq.Values) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("values"), "must be specified when `operator` is 'In' or 'NotIn'"))
		}
	case corev1.NodeSelectorOpExists, corev1.NodeSelectorOpDoesNotExist:
		if len(rq.Values) > 0 {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("values"), "may not be specified when `operator` is 'Exists' or 'DoesNotExist'"))
		}

	case corev1.NodeSelectorOpGt, corev1.NodeSelectorOpLt:
		if len(rq.Values) != 1 {
			allErrs = append(allErrs, field.Required(fldPath.Child("values"), "must be specified single value when `operator` is 'Lt' or 'Gt'"))
		}
	default:
		allErrs = append(allErrs, field.Invalid(fldPath.Child("operator"), rq.Operator, "not a valid selector operator"))
	}

	allErrs = append(allErrs, unversionedvalidation.ValidateLabelName(rq.Key, fldPath.Child("key"))...)

	return allErrs
}

func validateNodeFieldSelectorRequirement(req corev1.NodeSelectorRequirement, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	switch req.Operator {
	case corev1.NodeSelectorOpIn, corev1.NodeSelectorOpNotIn:
		if len(req.Values) != 1 {
			allErrs = append(allErrs, field.Required(fldPath.Child("values"),
				"must be only one value when `operator` is 'In' or 'NotIn' for node field selector"))
		}
	default:
		allErrs = append(allErrs, field.Invalid(fldPath.Child("operator"), req.Operator, "not a valid selector operator"))
	}

	if vf, found := nodeFieldSelectorValidators[req.Key]; !found {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("key"), req.Key, "not a valid field selector key"))
	} else {
		for i, v := range req.Values {
			for _, msg := range vf(v, false) {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("values").Index(i), v, msg))
			}
		}
	}

	return allErrs
}

func validateTolerations(tolerations []corev1.Toleration, fldPath *field.Path) field.ErrorList {
	allErrors := field.ErrorList{}
	for i, toleration := range tolerations {
		idxPath := fldPath.Index(i)
		// validate the toleration key
		if len(toleration.Key) > 0 {
			allErrors = append(allErrors, unversionedvalidation.ValidateLabelName(toleration.Key, idxPath.Child("key"))...)
		}

		// empty toleration key with Exists operator and empty value means match all taints
		if len(toleration.Key) == 0 && toleration.Operator != corev1.TolerationOpExists {
			allErrors = append(allErrors, field.Invalid(idxPath.Child("operator"), toleration.Operator,
				"operator must be Exists when `key` is empty, which means \"match all values and all keys\""))
		}

		if toleration.TolerationSeconds != nil && toleration.Effect != corev1.TaintEffectNoExecute {
			allErrors = append(allErrors, field.Invalid(idxPath.Child("effect"), toleration.Effect,
				"effect must be 'NoExecute' when `tolerationSeconds` is set"))
		}

		// validate toleration operator and value
		switch toleration.Operator {
		// empty operator means Equal
		case corev1.TolerationOpEqual, "":
			if errs := validation.IsValidLabelValue(toleration.Value); len(errs) != 0 {
				allErrors = append(allErrors, field.Invalid(idxPath.Child("operator"), toleration.Value, strings.Join(errs, ";")))
			}
		case corev1.TolerationOpExists:
			if len(toleration.Value) > 0 {
				allErrors = append(allErrors, field.Invalid(idxPath.Child("operator"), toleration, "value must be empty when `operator` is 'Exists'"))
			}
		default:
			validValues := []string{string(corev1.TolerationOpEqual), string(corev1.TolerationOpExists)}
			allErrors = append(allErrors, field.NotSupported(idxPath.Child("operator"), toleration.Operator, validValues))
		}

		// validate toleration effect, empty toleration effect means match all taint effects
		if len(toleration.Effect) > 0 {
			allErrors = append(allErrors, validateTaintEffect(&toleration.Effect, true, idxPath.Child("effect"))...)
		}
	}
	return allErrors
}

func validateTaintEffect(effect *corev1.TaintEffect, allowEmpty bool, fldPath *field.Path) field.ErrorList {
	if !allowEmpty && len(*effect) == 0 {
		return field.ErrorList{field.Required(fldPath, "")}
	}

	allErrors := field.ErrorList{}
	switch *effect {
	// TODO: Replace next line with subsequent commented-out line when implement TaintEffectNoScheduleNoAdmit.
	case corev1.TaintEffectNoSchedule, corev1.TaintEffectPreferNoSchedule, corev1.TaintEffectNoExecute:
		// case core.TaintEffectNoSchedule, core.TaintEffectPreferNoSchedule, core.TaintEffectNoScheduleNoAdmit, core.TaintEffectNoExecute:
	default:
		validValues := []string{
			string(corev1.TaintEffectNoSchedule),
			string(corev1.TaintEffectPreferNoSchedule),
			string(corev1.TaintEffectNoExecute),
			// TODO: Uncomment this block when implement TaintEffectNoScheduleNoAdmit.
			// string(core.TaintEffectNoScheduleNoAdmit),
		}
		allErrors = append(allErrors, field.NotSupported(fldPath, *effect, validValues))
	}
	return allErrors
}

func (r *CodeGen) validatePodSpec() *field.ErrorList {
	var allErrors field.ErrorList
	if podSecurityContext := r.Spec.PodSpec.SecurityContext; podSecurityContext != nil {
		if errorList := validatePodSecurityContext(podSecurityContext, field.NewPath("spec").Child("podSpec").Child("securityContext"), podValidationOptions{}); len(errorList) != 0 {
			allErrors = append(allErrors, errorList...)
		}
	}

	if resourceRequirements := r.Spec.PodSpec.CodeGenContainerSpec.Resources; resourceRequirements != nil {
		if errorList := validateResourceRequirements(resourceRequirements, field.NewPath("spec").Child("podSpec").Child("containerSpec").Child("resources")); len(errorList) != 0 {
			allErrors = append(allErrors, errorList...)
		}
	}

	if len(allErrors) == 0 {
		return nil
	}

	return &allErrors
}

func validatePodSecurityContext(securityContext *corev1.PodSecurityContext, fldPath *field.Path, opts podValidationOptions) field.ErrorList {
	allErrs := field.ErrorList{}

	if securityContext != nil {
		if securityContext.FSGroup != nil {
			for _, msg := range validation.IsValidGroupID(*securityContext.FSGroup) {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("fsGroup"), *(securityContext.FSGroup), msg))
			}
		}
		if securityContext.RunAsUser != nil {
			for _, msg := range validation.IsValidUserID(*securityContext.RunAsUser) {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("runAsUser"), *(securityContext.RunAsUser), msg))
			}
		}
		if securityContext.RunAsGroup != nil {
			for _, msg := range validation.IsValidGroupID(*securityContext.RunAsGroup) {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("runAsGroup"), *(securityContext.RunAsGroup), msg))
			}
		}
		for g, gid := range securityContext.SupplementalGroups {
			for _, msg := range validation.IsValidGroupID(gid) {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("supplementalGroups").Index(g), gid, msg))
			}
		}

		if len(securityContext.Sysctls) != 0 {
			allErrs = append(allErrs, validateSysctls(securityContext.Sysctls, fldPath.Child("sysctls"), opts.AllowSysctlRegexContainSlash)...)
		}

		if securityContext.FSGroupChangePolicy != nil {
			allErrs = append(allErrs, validateFSGroupChangePolicy(securityContext.FSGroupChangePolicy, fldPath.Child("fsGroupChangePolicy"))...)
		}

		allErrs = append(allErrs, validateSeccompProfileField(securityContext.SeccompProfile, fldPath.Child("seccompProfile"))...)
		allErrs = append(allErrs, validateWindowsSecurityContextOptions(securityContext.WindowsOptions, fldPath.Child("windowsOptions"))...)
	}

	return allErrs
}

func validateSysctls(sysctls []corev1.Sysctl, fldPath *field.Path, allowSysctlRegexContainSlash bool) field.ErrorList {
	allErrs := field.ErrorList{}
	names := make(map[string]struct{})
	for i, s := range sysctls {
		if len(s.Name) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Index(i).Child("name"), ""))
		} else if !isValidSysctlName(s.Name, allowSysctlRegexContainSlash) {
			allErrs = append(allErrs, field.Invalid(fldPath.Index(i).Child("name"), s.Name, fmt.Sprintf("must have at most %d characters and match regex %s", SysctlMaxLength, getSysctlFmt(allowSysctlRegexContainSlash))))
		} else if _, ok := names[s.Name]; ok {
			allErrs = append(allErrs, field.Duplicate(fldPath.Index(i).Child("name"), s.Name))
		}
		names[s.Name] = struct{}{}
	}
	return allErrs
}

func isValidSysctlName(name string, canContainSlash bool) bool {
	if len(name) > SysctlMaxLength {
		return false
	}
	if canContainSlash {
		return sysctlContainSlashRegexp.MatchString(name)
	}
	return sysctlRegexp.MatchString(name)
}

func getSysctlFmt(canContainSlash bool) string {
	if canContainSlash {
		// use relaxed validation everywhere in 1.24
		return SysctlContainSlashFmt
	}
	// Will be removed in 1.24
	return SysctlFmt
}

func validateFSGroupChangePolicy(fsGroupPolicy *corev1.PodFSGroupChangePolicy, fldPath *field.Path) field.ErrorList {
	allErrors := field.ErrorList{}
	if !validFSGroupChangePolicies.Has(string(*fsGroupPolicy)) {
		allErrors = append(allErrors, field.NotSupported(fldPath, fsGroupPolicy, validFSGroupChangePolicies.List()))
	}
	return allErrors
}

func validateSeccompProfileField(sp *corev1.SeccompProfile, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if sp == nil {
		return allErrs
	}

	if err := validateSeccompProfileType(fldPath.Child("type"), sp.Type); err != nil {
		allErrs = append(allErrs, err)
	}

	if sp.Type == corev1.SeccompProfileTypeLocalhost {
		if sp.LocalhostProfile == nil {
			allErrs = append(allErrs, field.Required(fldPath.Child("localhostProfile"), "must be set when seccomp type is Localhost"))
		} else {
			allErrs = append(allErrs, validateLocalDescendingPath(*sp.LocalhostProfile, fldPath.Child("localhostProfile"))...)
		}
	} else {
		if sp.LocalhostProfile != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("localhostProfile"), sp, "can only be set when seccomp type is Localhost"))
		}
	}

	return allErrs
}

func validateSeccompProfileType(fldPath *field.Path, seccompProfileType corev1.SeccompProfileType) *field.Error {
	switch seccompProfileType {
	case corev1.SeccompProfileTypeLocalhost, corev1.SeccompProfileTypeRuntimeDefault, corev1.SeccompProfileTypeUnconfined:
		return nil
	case "":
		return field.Required(fldPath, "type is required when seccompProfile is set")
	default:
		return field.NotSupported(fldPath, seccompProfileType, []string{string(corev1.SeccompProfileTypeLocalhost), string(corev1.SeccompProfileTypeRuntimeDefault), string(corev1.SeccompProfileTypeUnconfined)})
	}
}

func validateLocalDescendingPath(targetPath string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if path.IsAbs(targetPath) {
		allErrs = append(allErrs, field.Invalid(fldPath, targetPath, "must be a relative path"))
	}

	allErrs = append(allErrs, validatePathNoBacksteps(targetPath, fldPath)...)

	return allErrs
}

func validatePathNoBacksteps(targetPath string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	parts := strings.Split(filepath.ToSlash(targetPath), "/")
	for _, item := range parts {
		if item == ".." {
			allErrs = append(allErrs, field.Invalid(fldPath, targetPath, "must not contain '..'"))
			break // even for `../../..`, one error is sufficient to make the point
		}
	}
	return allErrs
}

func validateWindowsSecurityContextOptions(windowsOptions *corev1.WindowsSecurityContextOptions, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if windowsOptions == nil {
		return allErrs
	} else {
		allErrs = append(allErrs, field.Forbidden(fieldPath, "windowsOptions does not apply"))
		return allErrs
	}
}

func validateResourceRequirements(requirements *corev1.ResourceRequirements, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	limPath := fldPath.Child("limits")
	reqPath := fldPath.Child("requests")
	for resourceName, quantity := range requirements.Limits {
		fldPath := limPath.Key(string(resourceName))
		// Validate resource name.
		allErrs = append(allErrs, validateContainerResourceName(string(resourceName), fldPath)...)

		// Validate resource quantity.
		allErrs = append(allErrs, validateResourceQuantityValue(string(resourceName), quantity, fldPath)...)

	}
	for resourceName, quantity := range requirements.Requests {
		fldPath := reqPath.Key(string(resourceName))
		// Validate resource name.
		allErrs = append(allErrs, validateContainerResourceName(string(resourceName), fldPath)...)
		// Validate resource quantity.
		allErrs = append(allErrs, validateResourceQuantityValue(string(resourceName), quantity, fldPath)...)

		// Check that request <= limit.
		limitQuantity, exists := requirements.Limits[resourceName]
		if exists {
			// For GPUs, not only requests can't exceed limits, they also can't be lower, i.e. must be equal.
			if quantity.Cmp(limitQuantity) != 0 && !isOvercommitAllowed(resourceName) {
				allErrs = append(allErrs, field.Invalid(reqPath, quantity.String(), fmt.Sprintf("must be equal to %s limit", resourceName)))
			} else if quantity.Cmp(limitQuantity) > 0 {
				allErrs = append(allErrs, field.Invalid(reqPath, quantity.String(), fmt.Sprintf("must be less than or equal to %s limit", resourceName)))
			}
		}
	}

	return allErrs
}

func validateContainerResourceName(value string, fldPath *field.Path) field.ErrorList {
	allErrs := validateResourceName(value, fldPath)
	if len(strings.Split(value, "/")) == 1 {
		if !isStandardContainerResourceName(value) {
			return append(allErrs, field.Invalid(fldPath, value, "must be a standard resource for containers"))
		}
	} else if !isNativeResource(corev1.ResourceName(value)) {
		if !isExtendedResourceName(corev1.ResourceName(value)) {
			return append(allErrs, field.Invalid(fldPath, value, "doesn't follow extended resource name standard"))
		}
	}
	return allErrs
}

func validateResourceQuantityValue(resource string, value resource.Quantity, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, validateNonnegativeQuantity(value, fldPath)...)
	if isIntegerResourceName(resource) {
		if value.MilliValue()%int64(1000) != int64(0) {
			allErrs = append(allErrs, field.Invalid(fldPath, value, isNotIntegerErrorMsg))
		}
	}
	return allErrs
}

func validateResourceName(value string, fldPath *field.Path) field.ErrorList {
	allErrs := validateQualifiedName(value, fldPath)
	if len(allErrs) != 0 {
		return allErrs
	}

	if len(strings.Split(value, "/")) == 1 {
		if !isStandardResourceName(value) {
			return append(allErrs, field.Invalid(fldPath, value, "must be a standard resource type or fully qualified"))
		}
	}

	return allErrs
}

func isOvercommitAllowed(name corev1.ResourceName) bool {
	return isNativeResource(name) &&
		!isHugePageResourceName(name)
}

func isNativeResource(name corev1.ResourceName) bool {
	return !strings.Contains(string(name), "/") ||
		isPrefixedNativeResource(name)
}

func isPrefixedNativeResource(name corev1.ResourceName) bool {
	return strings.Contains(string(name), corev1.ResourceDefaultNamespacePrefix)
}

func isHugePageResourceName(name corev1.ResourceName) bool {
	return strings.HasPrefix(string(name), corev1.ResourceHugePagesPrefix)
}

func validateNonnegativeQuantity(value resource.Quantity, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if value.Cmp(resource.Quantity{}) < 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, value.String(), isNegativeErrorMsg))
	}
	return allErrs
}

func isIntegerResourceName(str string) bool {
	return integerResources.Has(str) || isExtendedResourceName(corev1.ResourceName(str))
}

func isExtendedResourceName(name corev1.ResourceName) bool {
	if isNativeResource(name) || strings.HasPrefix(string(name), corev1.DefaultResourceRequestsPrefix) {
		return false
	}
	// Ensure it satisfies the rules in IsQualifiedName() after converted into quota resource name
	nameForQuota := fmt.Sprintf("%s%s", corev1.DefaultResourceRequestsPrefix, string(name))
	if errs := validation.IsQualifiedName(string(nameForQuota)); len(errs) != 0 {
		return false
	}
	return true
}

func validateQualifiedName(value string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	for _, msg := range validation.IsQualifiedName(value) {
		allErrs = append(allErrs, field.Invalid(fldPath, value, msg))
	}
	return allErrs
}

func isStandardResourceName(str string) bool {
	return standardResources.Has(str) || isQuotaHugePageResourceName(corev1.ResourceName(str))
}

func isStandardContainerResourceName(str string) bool {
	return standardContainerResources.Has(str) || isHugePageResourceName(corev1.ResourceName(str))
}

func isQuotaHugePageResourceName(name corev1.ResourceName) bool {
	return strings.HasPrefix(string(name), corev1.ResourceHugePagesPrefix) || strings.HasPrefix(string(name), corev1.ResourceRequestsHugePagesPrefix)
}
