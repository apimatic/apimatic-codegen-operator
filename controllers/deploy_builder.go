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

package controllers

import (
	"encoding/json"
	"reflect"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	apicodegenv1beta2 "github.com/apimatic/apimatic-codegen-operator/api/v1beta2"
)

const (
	LICENSEBLOBKEY   = "LICENSEBLOB"
	LICENSEPATHKEY   = "LICENSEPATH"
	LICENSEPATHVALUE = "/usr/apimatic/license"
	LICENSEVOLUME    = "apimaticlicense"
	PROBEROUTE       = "/api/template"
	CONTAINERNAME    = "codegen"
)

var (
	ContainerProbe corev1.Probe = corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: PROBEROUTE,
				Port: intstr.FromInt(80),
				HTTPHeaders: []corev1.HTTPHeader{{
					Name:  "Accept",
					Value: "application/json",
				}},
			},
		},
		InitialDelaySeconds: int32(5),
		TimeoutSeconds:      int32(5),
		FailureThreshold:    int32(1),
		PeriodSeconds:       int32(15),
	}
)

// GenericResource have both meta and runtime interfaces
type GenericResource interface {
	metav1.Object
	runtime.Object
}

// generateResources produces a list of rbac and deployment resources suitable for APIMatic.
// Note that these objects are used as a subset reference for testing cluster object correctness, so be careful
// not to fill in ephemeral fields here like timestamps, uuids. Return the expected reference objects.
func generateResources(codegen *apicodegenv1beta2.CodeGen) []GenericResource {
	resources := []GenericResource{}
	service := serviceForCodeGen(codegen)
	deployment := deploymentForCodeGen(codegen)

	resources = append(resources, service, deployment)

	return resources
}

func serviceForCodeGen(codegen *apicodegenv1beta2.CodeGen) *corev1.Service {
	labels := codegen.Spec.Labels
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      codegen.Name,
			Namespace: codegen.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Type:     codegen.Spec.ServiceSpec.Type,
			Ports: []corev1.ServicePort{{
				Name:       codegen.Spec.ServiceSpec.CodeGenServicePort.Name,
				TargetPort: intstr.FromInt(80),
				Port:       80,
			}},
		},
	}

	if clusterIP := codegen.Spec.ServiceSpec.ClusterIP; clusterIP != nil {
		service.Spec.ClusterIP = *clusterIP
	}

	if nodeport := codegen.Spec.ServiceSpec.CodeGenServicePort.NodePort; nodeport != nil {
		service.Spec.Ports[0].NodePort = *nodeport
	}

	if loadbalancerIP := codegen.Spec.ServiceSpec.LoadBalancerIP; loadbalancerIP != nil {
		service.Spec.LoadBalancerIP = *loadbalancerIP
	}

	if loadbalancerClass := codegen.Spec.ServiceSpec.LoadBalancerClass; loadbalancerClass != nil {
		service.Spec.LoadBalancerClass = loadbalancerClass
	}

	if externalTrafficPolicy := codegen.Spec.ServiceSpec.ExternalTrafficPolicy; externalTrafficPolicy != nil {
		service.Spec.ExternalTrafficPolicy = *externalTrafficPolicy
	}

	if customServiceAnnotations := codegen.Spec.ServiceSpec.ServiceCustomAnnotations; customServiceAnnotations != nil {
		service.ObjectMeta.Annotations = customServiceAnnotations
	}

	return service
}

func deploymentForCodeGen(codegen *apicodegenv1beta2.CodeGen) *appsv1.Deployment {
	labels := codegen.Spec.Labels
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      codegen.Name,
			Namespace: codegen.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &codegen.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: codegen.Spec.PodSpec.TerminationGracePeriodSeconds,
					Containers: []corev1.Container{{
						Image: *codegen.Spec.PodSpec.CodeGenContainerSpec.Image,
						Name:  CONTAINERNAME,
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
						}},
						LivenessProbe:  &ContainerProbe,
						ReadinessProbe: &ContainerProbe,
					}},
				},
			},
		},
	}

	if serviceAccountName := codegen.Spec.PodSpec.ServiceAccountName; serviceAccountName != nil {
		deployment.Spec.Template.Spec.ServiceAccountName = *serviceAccountName
	}

	if imagePullPolicy := codegen.Spec.PodSpec.CodeGenContainerSpec.ImagePullPolicy; imagePullPolicy != nil {
		deployment.Spec.Template.Spec.Containers[0].ImagePullPolicy = *imagePullPolicy
	}

	if imagePullSecret := codegen.Spec.PodSpec.CodeGenContainerSpec.ImagePullSecret; imagePullSecret != nil {
		deployment.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{
			Name: *imagePullSecret,
		}}
	}

	if resources := codegen.Spec.PodSpec.CodeGenContainerSpec.Resources; resources != nil {
		deployment.Spec.Template.Spec.Containers[0].Resources = *resources
	}

	if securityContext := codegen.Spec.PodSpec.SecurityContext; securityContext != nil {
		deployment.Spec.Template.Spec.SecurityContext = securityContext
	}

	if schedulerName := codegen.Spec.PodSpec.SchedulerName; schedulerName != nil {
		deployment.Spec.Template.Spec.SchedulerName = *schedulerName
	}

	if priorityClassName := codegen.Spec.PodSpec.PriorityClassName; priorityClassName != nil {
		deployment.Spec.Template.Spec.PriorityClassName = *priorityClassName
	}

	if licenseSource := codegen.Spec.LicenseSpec.CodeGenLicenseSourceType; strings.Compare(licenseSource, "LicenseBlob") == 0 {
		deployment.Spec.Template.Spec.Containers[0].Env = []corev1.EnvVar{{
			Name:  LICENSEBLOBKEY,
			Value: codegen.Spec.LicenseSpec.CodeGenLicenseSourceValue,
		}}
	} else {
		deployment.Spec.Template.Spec.Containers[0].Env = []corev1.EnvVar{{
			Name:  LICENSEPATHKEY,
			Value: LICENSEPATHVALUE,
		}}

		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{{
			ReadOnly:  true,
			Name:      LICENSEVOLUME,
			MountPath: LICENSEPATHVALUE,
		}}

		deployment.Spec.Template.Spec.Volumes = []corev1.Volume{
			*apimaticLicenseVolumeSource(codegen),
		}
	}

	deployment.Spec.Template.Spec.Affinity = &corev1.Affinity{}

	deployment.Spec.Template.Spec.Affinity.PodAntiAffinity = &corev1.PodAntiAffinity{
		PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{{
			Weight: 100,
			PodAffinityTerm: corev1.PodAffinityTerm{
				TopologyKey: "anti-affinity",
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: labels,
				},
			},
		}},
	}

	if podPlacement := codegen.Spec.CodeGenPodPlacementSpec; podPlacement != nil {
		if nodeAffinity := podPlacement.NodeAffinity; nodeAffinity != nil {
			deployment.Spec.Template.Spec.Affinity.NodeAffinity = nodeAffinity
		}

		if podAffinity := podPlacement.PodAffinity; podAffinity != nil {
			deployment.Spec.Template.Spec.Affinity.PodAffinity = podAffinity
		}

		if nodeName := podPlacement.NodeName; nodeName != nil {
			deployment.Spec.Template.Spec.NodeName = *nodeName
		}

		if tolerations := podPlacement.Tolerations; tolerations != nil {
			deployment.Spec.Template.Spec.Tolerations = []corev1.Toleration{}
			deployment.Spec.Template.Spec.Tolerations = append(deployment.Spec.Template.Spec.Tolerations, tolerations...)
		}

		if nodeSelector := podPlacement.NodeSelector; nodeSelector != nil {
			deployment.Spec.Template.Spec.NodeSelector = nodeSelector
		}
	}

	return deployment
}

func apimaticLicenseVolumeSource(codegen *apicodegenv1beta2.CodeGen) *corev1.Volume {
	licenseSource := &corev1.Volume{
		Name: LICENSEVOLUME,
	}
	if strings.Compare(codegen.Spec.LicenseSpec.CodeGenLicenseSourceType, "ConfigMap") == 0 {
		licenseSource.VolumeSource = corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: codegen.Spec.LicenseSpec.CodeGenLicenseSourceValue,
				},
			},
		}
	} else {
		licenseSource.VolumeSource = corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: codegen.Spec.LicenseSpec.CodeGenLicenseSourceValue,
			},
		}
	}

	return licenseSource
}

func updateResource(codegen *apicodegenv1beta2.CodeGen, clusterResource GenericResource) (GenericResource, bool, error) {
	kind := reflect.ValueOf(clusterResource).Elem().Type().String()
	if strings.Compare(kind, "v1.Service") == 0 {
		if serviceJsonData, err := json.Marshal(clusterResource); err == nil {
			service := corev1.Service{}
			if err := json.Unmarshal([]byte(serviceJsonData), &service); err == nil {
				update := updateService(codegen, &service)
				return &service, update, nil
			} else {
				return clusterResource, true, err
			}
		}
	}

	if strings.Compare(kind, "v1.Deployment") == 0 {
		if deploymentJsonData, err := json.Marshal(clusterResource); err == nil {
			deployment := appsv1.Deployment{}
			if err := json.Unmarshal([]byte(deploymentJsonData), &deployment); err == nil {
				update := updateDeployment(&deployment)
				return &deployment, update, nil
			} else {
				return clusterResource, true, err
			}
		}
	}

	return clusterResource, false, nil

}

func updateService(codegen *apicodegenv1beta2.CodeGen, service *corev1.Service) bool {
	actualAnnotations := service.GetAnnotations()
	desiredAnnotions := codegen.Spec.ServiceSpec.ServiceCustomAnnotations

	if !reflect.DeepEqual(actualAnnotations, desiredAnnotions) {
		service.Annotations = desiredAnnotions
		return true
	}

	return false
}

func updateDeployment(deployment *appsv1.Deployment) bool {
	if envVariableName := deployment.Spec.Template.Spec.Containers[0].Env[0].Name; strings.Compare(LICENSEBLOBKEY, envVariableName) == 0 {
		if volumes := deployment.Spec.Template.Spec.Volumes; volumes != nil {
			deployment.Spec.Template.Spec.Volumes = nil
			deployment.Spec.Template.Spec.Containers[0].VolumeMounts = nil

			return true
		}
	}

	return false
}

func (r *CodeGenReconciler) setupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apicodegenv1beta2.CodeGen{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
