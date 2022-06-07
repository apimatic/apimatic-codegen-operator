package controllers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	apicodegenv1beta2 "github.com/apimatic/apimatic-codegen-operator/api/v1beta2"
)

var _ = Describe("DeployBuilder", func() {
	Context("When providing CodeGen instance as argument to DeployBuilder using LicenseBlob value as LicenseSourceType", func() {
		const (
			CodeGenName        = "test-deployBuilder"
			CodeGenNamespace   = "default"
			LicenseSourceType  = "LicenseBlob"
			LicenseSourceValue = "12345"
			key1               = "key1"
			key2               = "key2"
			value1             = "value1"
			value2             = "value2"
			tolerationKey      = "testTolerationKey"
			nodeAffinityWt     = int32(100)
			podAffinityWt      = int32(50)
		)
		It("Should produce Service and Deployment resource as intended", func() {
			serviceType := corev1.ServiceTypeLoadBalancer
			clusterIp := "10.1.1.1"
			externalTrafficPolicyType := corev1.ServiceExternalTrafficPolicyTypeCluster
			nodePort := int32(30000)
			loadBalancerIP := "10.1.1.2"
			loadBalancerClass := "testloadbalancerclass"
			servicePortName := "testName"
			imageName := "testImage"
			imagePullPolicy := corev1.PullIfNotPresent
			imagePullSecret := "pullSecret"
			terminationGracePeriodInSeconds := int64(30)
			serviceAccountName := "testAccount"
			schedulerName := "testScheduler"
			priorityClassName := "testPriorityClass"
			replicas := int32(3)
			securityContext := corev1.PodSecurityContext{
				SupplementalGroups: []int64{
					56,
				},
			}
			resources := corev1.ResourceRequirements{
				Limits:   corev1.ResourceList{},
				Requests: corev1.ResourceList{},
			}
			resources.Limits[corev1.ResourceCPU] = resource.Quantity{
				Format: resource.DecimalExponent,
			}
			resources.Requests[corev1.ResourceCPU] = resource.Quantity{
				Format: resource.DecimalSI,
			}

			nodeName := "testNodeName"
			nodeSelector := make(map[string]string)
			nodeSelector[key1] = value1
			nodeSelector[key2] = value2
			tolerations := []corev1.Toleration{{
				Key: tolerationKey,
			}}
			nodeAffinity := corev1.NodeAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{{
					Weight: nodeAffinityWt,
				}},
			}
			podAffinity := corev1.PodAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{{
					Weight: podAffinityWt,
				}},
			}

			serviceCustomAnnotations := make(map[string]string)
			serviceCustomAnnotations[key1] = value1
			serviceCustomAnnotations[key2] = value2
			labels := make(map[string]string)
			labels[key1] = value1
			labels[key2] = value2

			codegen := &apicodegenv1beta2.CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNamespace,
				},
				Spec: apicodegenv1beta2.CodeGenSpec{
					Replicas: replicas,
					Labels:   labels,
					ServiceSpec: &apicodegenv1beta2.CodeGenServiceSpec{
						Type:                  serviceType,
						ClusterIP:             &clusterIp,
						ExternalTrafficPolicy: &externalTrafficPolicyType,
						CodeGenServicePort: &apicodegenv1beta2.CodeGenServicePort{
							Name:     servicePortName,
							NodePort: &nodePort,
						},
						LoadBalancerIP:           &loadBalancerIP,
						LoadBalancerClass:        &loadBalancerClass,
						ServiceCustomAnnotations: serviceCustomAnnotations,
					},
					PodSpec: &apicodegenv1beta2.CodeGenPodSpec{
						CodeGenContainerSpec: &apicodegenv1beta2.CodeGenContainerSpec{
							Image:           &imageName,
							ImagePullPolicy: &imagePullPolicy,
							ImagePullSecret: &imagePullSecret,
							Resources:       &resources,
						},
						TerminationGracePeriodSeconds: &terminationGracePeriodInSeconds,
						ServiceAccountName:            &serviceAccountName,
						SecurityContext:               &securityContext,
						SchedulerName:                 &schedulerName,
						PriorityClassName:             &priorityClassName,
					},
					LicenseSpec: apicodegenv1beta2.CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  LicenseSourceType,
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
					CodeGenPodPlacementSpec: &apicodegenv1beta2.CodeGenPodPlacementSpec{
						NodeName:     &nodeName,
						NodeSelector: nodeSelector,
						Tolerations:  tolerations,
						NodeAffinity: &nodeAffinity,
						PodAffinity:  &podAffinity,
					},
				},
			}

			genericResources := generateResources(codegen)
			Expect(len(genericResources)).Should(Equal(int(2)))

			var createdService *corev1.Service = serviceForCodeGen(codegen)
			Expect(createdService).ShouldNot(BeNil())
			Expect(createdService.Spec.Type).Should(Equal(corev1.ServiceType(serviceType)))
			Expect(createdService.Spec.Ports[0].Name).Should(Equal(string(servicePortName)))
			Expect(createdService.Spec.Ports[0].NodePort).Should(Equal(int32(nodePort)))
			Expect(createdService.Spec.Ports[0].Port).Should(Equal(int32(8080)))
			Expect(createdService.Spec.Ports[0].TargetPort).Should(Equal(intstr.FromInt(8080)))
			Expect(createdService.Spec.ExternalTrafficPolicy).Should(Equal(corev1.ServiceExternalTrafficPolicyType(externalTrafficPolicyType)))
			Expect(createdService.Spec.ClusterIP).Should(Equal(string(clusterIp)))
			Expect(createdService.Spec.LoadBalancerIP).Should(Equal(string(loadBalancerIP)))
			Expect(*createdService.Spec.LoadBalancerClass).Should(Equal(string(loadBalancerClass)))
			Expect(createdService.Name).Should(Equal(string(CodeGenName)))
			Expect(createdService.Namespace).Should(Equal(string(CodeGenNamespace)))
			Expect(createdService.Annotations[key1]).Should(Equal(string(value1)))
			Expect(createdService.Annotations[key2]).Should(Equal(string(value2)))
			Expect(createdService.Labels[key1]).Should(Equal(string(value1)))
			Expect(createdService.Labels[key2]).Should(Equal(string(value2)))
			Expect(createdService.Spec.Selector[key1]).Should(Equal(string(value1)))
			Expect(createdService.Spec.Selector[key2]).Should(Equal(string(value2)))

			var createdDeployment *appsv1.Deployment = deploymentForCodeGen(codegen)
			Expect(createdDeployment).ShouldNot(BeNil())
			Expect(createdDeployment.Name).Should(Equal(string(CodeGenName)))
			Expect(createdDeployment.Namespace).Should(Equal(string(CodeGenNamespace)))
			Expect(createdDeployment.Labels[key1]).Should(Equal(string(value1)))
			Expect(createdDeployment.Labels[key2]).Should(Equal(string(value2)))
			Expect(*createdDeployment.Spec.Replicas).Should(Equal(int32(replicas)))
			Expect(createdDeployment.Spec.Selector.MatchLabels[key1]).Should(Equal(string(value1)))
			Expect(createdDeployment.Spec.Selector.MatchLabels[key2]).Should(Equal(string(value2)))
			Expect(createdDeployment.Spec.Template.Labels[key1]).Should(Equal(string(value1)))
			Expect(createdDeployment.Spec.Template.Labels[key2]).Should(Equal(string(value2)))
			Expect(createdDeployment.Spec.Template.Spec.ServiceAccountName).Should(Equal(string(serviceAccountName)))
			Expect(*createdDeployment.Spec.Template.Spec.TerminationGracePeriodSeconds).Should(Equal(int64(terminationGracePeriodInSeconds)))
			Expect(createdDeployment.Spec.Template.Spec.ImagePullSecrets[0].Name).Should(Equal(string(imagePullSecret)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Image).Should(Equal(string(imageName)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy).Should(Equal(corev1.PullPolicy(imagePullPolicy)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Env[0].Name).Should(Equal(string(LICENSEBLOBKEY)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Env[0].Value).Should(Equal(string(LicenseSourceValue)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).Should(Equal(int32(8080)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU].Format).Should(Equal(resource.Format(resource.DecimalExponent)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU].Format).Should(Equal(resource.Format(resource.DecimalSI)))
			Expect(*createdDeployment.Spec.Template.Spec.Containers[0].LivenessProbe).Should(Equal(corev1.Probe(ContainerProbe)))
			Expect(*createdDeployment.Spec.Template.Spec.Containers[0].ReadinessProbe).Should(Equal(corev1.Probe(ContainerProbe)))
			Expect(createdDeployment.Spec.Template.Spec.SecurityContext.SupplementalGroups[0]).Should(Equal(int64(56)))
			Expect(createdDeployment.Spec.Template.Spec.SchedulerName).Should(Equal(string(schedulerName)))
			Expect(createdDeployment.Spec.Template.Spec.PriorityClassName).Should(Equal(string(priorityClassName)))
			Expect(createdDeployment.Spec.Template.Spec.NodeName).Should(Equal(string(nodeName)))
			Expect(createdDeployment.Spec.Template.Spec.Tolerations[0].Key).Should(Equal(string(tolerationKey)))
			Expect(createdDeployment.Spec.Template.Spec.NodeSelector[key1]).Should(Equal(string(value1)))
			Expect(createdDeployment.Spec.Template.Spec.NodeSelector[key2]).Should(Equal(string(value2)))
			Expect(createdDeployment.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].Weight).Should(Equal(int32(nodeAffinityWt)))
			Expect(createdDeployment.Spec.Template.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].Weight).Should(Equal(int32(podAffinityWt)))
		})
	})

	Context("When providing CodeGen instance as argument to DeployBuilder using ConfigMap as LicenseSourceType value", func() {
		const (
			CodeGenName        = "test-deployBuilder"
			CodeGenNamespace   = "default"
			LicenseSourceType  = "ConfigMap"
			LicenseSourceValue = "testconfigmap"
			key1               = "key1"
			key2               = "key2"
			value1             = "value1"
			value2             = "value2"
			tolerationKey      = "testTolerationKey"
			nodeAffinityWt     = int32(100)
			podAffinityWt      = int32(50)
		)
		It("Should produce Service and Deployment resource as intended", func() {
			serviceType := corev1.ServiceTypeLoadBalancer
			clusterIp := "10.1.1.1"
			externalTrafficPolicyType := corev1.ServiceExternalTrafficPolicyTypeCluster
			nodePort := int32(30000)
			loadBalancerIP := "10.1.1.2"
			loadBalancerClass := "testloadbalancerclass"
			servicePortName := "testName"
			imageName := "testImage"
			imagePullPolicy := corev1.PullIfNotPresent
			imagePullSecret := "pullSecret"
			terminationGracePeriodInSeconds := int64(30)
			serviceAccountName := "testAccount"
			schedulerName := "testScheduler"
			priorityClassName := "testPriorityClass"
			replicas := int32(3)
			securityContext := corev1.PodSecurityContext{
				SupplementalGroups: []int64{
					56,
				},
			}
			resources := corev1.ResourceRequirements{
				Limits:   corev1.ResourceList{},
				Requests: corev1.ResourceList{},
			}
			resources.Limits[corev1.ResourceCPU] = resource.Quantity{
				Format: resource.DecimalExponent,
			}
			resources.Requests[corev1.ResourceCPU] = resource.Quantity{
				Format: resource.DecimalSI,
			}
			nodeName := "testNodeName"
			nodeSelector := make(map[string]string)
			nodeSelector[key1] = value1
			nodeSelector[key2] = value2
			tolerations := []corev1.Toleration{{
				Key: tolerationKey,
			}}
			nodeAffinity := corev1.NodeAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{{
					Weight: nodeAffinityWt,
				}},
			}
			podAffinity := corev1.PodAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{{
					Weight: podAffinityWt,
				}},
			}
			serviceCustomAnnotations := make(map[string]string)
			serviceCustomAnnotations[key1] = value1
			serviceCustomAnnotations[key2] = value2
			labels := make(map[string]string)
			labels[key1] = value1
			labels[key2] = value2

			volumeSource := corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: LicenseSourceValue,
					},
				},
			}

			codegen := &apicodegenv1beta2.CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNamespace,
				},
				Spec: apicodegenv1beta2.CodeGenSpec{
					Replicas: replicas,
					Labels:   labels,
					ServiceSpec: &apicodegenv1beta2.CodeGenServiceSpec{
						Type:                  serviceType,
						ClusterIP:             &clusterIp,
						ExternalTrafficPolicy: &externalTrafficPolicyType,
						CodeGenServicePort: &apicodegenv1beta2.CodeGenServicePort{
							Name:     servicePortName,
							NodePort: &nodePort,
						},
						LoadBalancerIP:           &loadBalancerIP,
						LoadBalancerClass:        &loadBalancerClass,
						ServiceCustomAnnotations: serviceCustomAnnotations,
					},
					PodSpec: &apicodegenv1beta2.CodeGenPodSpec{
						CodeGenContainerSpec: &apicodegenv1beta2.CodeGenContainerSpec{
							Image:           &imageName,
							ImagePullPolicy: &imagePullPolicy,
							ImagePullSecret: &imagePullSecret,
							Resources:       &resources,
						},
						TerminationGracePeriodSeconds: &terminationGracePeriodInSeconds,
						ServiceAccountName:            &serviceAccountName,
						SecurityContext:               &securityContext,
						SchedulerName:                 &schedulerName,
						PriorityClassName:             &priorityClassName,
					},
					LicenseSpec: apicodegenv1beta2.CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  LicenseSourceType,
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
					CodeGenPodPlacementSpec: &apicodegenv1beta2.CodeGenPodPlacementSpec{
						NodeName:     &nodeName,
						NodeSelector: nodeSelector,
						Tolerations:  tolerations,
						NodeAffinity: &nodeAffinity,
						PodAffinity:  &podAffinity,
					},
				},
			}

			genericResources := generateResources(codegen)
			Expect(len(genericResources)).Should(Equal(int(2)))

			var createdService *corev1.Service = serviceForCodeGen(codegen)
			Expect(createdService).ShouldNot(BeNil())
			Expect(createdService.Spec.Type).Should(Equal(corev1.ServiceType(serviceType)))
			Expect(createdService.Spec.Ports[0].Name).Should(Equal(string(servicePortName)))
			Expect(createdService.Spec.Ports[0].NodePort).Should(Equal(int32(nodePort)))
			Expect(createdService.Spec.Ports[0].Port).Should(Equal(int32(8080)))
			Expect(createdService.Spec.Ports[0].TargetPort).Should(Equal(intstr.FromInt(8080)))
			Expect(createdService.Spec.ExternalTrafficPolicy).Should(Equal(corev1.ServiceExternalTrafficPolicyType(externalTrafficPolicyType)))
			Expect(createdService.Spec.ClusterIP).Should(Equal(string(clusterIp)))
			Expect(createdService.Spec.LoadBalancerIP).Should(Equal(string(loadBalancerIP)))
			Expect(*createdService.Spec.LoadBalancerClass).Should(Equal(string(loadBalancerClass)))
			Expect(createdService.Name).Should(Equal(string(CodeGenName)))
			Expect(createdService.Namespace).Should(Equal(string(CodeGenNamespace)))
			Expect(createdService.Annotations[key1]).Should(Equal(string(value1)))
			Expect(createdService.Annotations[key2]).Should(Equal(string(value2)))
			Expect(createdService.Labels[key1]).Should(Equal(string(value1)))
			Expect(createdService.Labels[key2]).Should(Equal(string(value2)))
			Expect(createdService.Spec.Selector[key1]).Should(Equal(string(value1)))
			Expect(createdService.Spec.Selector[key2]).Should(Equal(string(value2)))

			var createdDeployment *appsv1.Deployment = deploymentForCodeGen(codegen)
			Expect(createdDeployment).ShouldNot(BeNil())
			Expect(createdDeployment.Name).Should(Equal(string(CodeGenName)))
			Expect(createdDeployment.Namespace).Should(Equal(string(CodeGenNamespace)))
			Expect(createdDeployment.Labels[key1]).Should(Equal(string(value1)))
			Expect(createdDeployment.Labels[key2]).Should(Equal(string(value2)))
			Expect(*createdDeployment.Spec.Replicas).Should(Equal(int32(replicas)))
			Expect(createdDeployment.Spec.Selector.MatchLabels[key1]).Should(Equal(string(value1)))
			Expect(createdDeployment.Spec.Selector.MatchLabels[key2]).Should(Equal(string(value2)))
			Expect(createdDeployment.Spec.Template.Labels[key1]).Should(Equal(string(value1)))
			Expect(createdDeployment.Spec.Template.Labels[key2]).Should(Equal(string(value2)))
			Expect(createdDeployment.Spec.Template.Spec.ServiceAccountName).Should(Equal(string(serviceAccountName)))
			Expect(*createdDeployment.Spec.Template.Spec.TerminationGracePeriodSeconds).Should(Equal(int64(terminationGracePeriodInSeconds)))
			Expect(createdDeployment.Spec.Template.Spec.ImagePullSecrets[0].Name).Should(Equal(string(imagePullSecret)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Image).Should(Equal(string(imageName)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy).Should(Equal(corev1.PullPolicy(imagePullPolicy)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Env[0].Name).Should(Equal(string(LICENSEPATHKEY)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Env[0].Value).Should(Equal(string(LICENSEPATHVALUE)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).Should(Equal(int32(8080)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU].Format).Should(Equal(resource.Format(resource.DecimalExponent)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU].Format).Should(Equal(resource.Format(resource.DecimalSI)))
			Expect(*createdDeployment.Spec.Template.Spec.Containers[0].LivenessProbe).Should(Equal(corev1.Probe(ContainerProbe)))
			Expect(*createdDeployment.Spec.Template.Spec.Containers[0].ReadinessProbe).Should(Equal(corev1.Probe(ContainerProbe)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name).Should(Equal(string(LICENSEVOLUME)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath).Should(Equal(string(LICENSEPATHVALUE)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].VolumeMounts[0].ReadOnly).Should(Equal(bool(true)))
			Expect(createdDeployment.Spec.Template.Spec.Volumes[0].Name).Should(Equal(string(LICENSEVOLUME)))
			Expect(createdDeployment.Spec.Template.Spec.Volumes[0].VolumeSource).Should(Equal(corev1.VolumeSource(volumeSource)))
			Expect(createdDeployment.Spec.Template.Spec.SecurityContext.SupplementalGroups[0]).Should(Equal(int64(56)))
			Expect(createdDeployment.Spec.Template.Spec.SchedulerName).Should(Equal(string(schedulerName)))
			Expect(createdDeployment.Spec.Template.Spec.PriorityClassName).Should(Equal(string(priorityClassName)))
			Expect(createdDeployment.Spec.Template.Spec.NodeName).Should(Equal(string(nodeName)))
			Expect(createdDeployment.Spec.Template.Spec.Tolerations[0].Key).Should(Equal(string(tolerationKey)))
			Expect(createdDeployment.Spec.Template.Spec.NodeSelector[key1]).Should(Equal(string(value1)))
			Expect(createdDeployment.Spec.Template.Spec.NodeSelector[key2]).Should(Equal(string(value2)))
			Expect(createdDeployment.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].Weight).Should(Equal(int32(nodeAffinityWt)))
			Expect(createdDeployment.Spec.Template.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].Weight).Should(Equal(int32(podAffinityWt)))
		})
	})

	Context("When providing CodeGen instance as argument to DeployBuilder using ConfigSecret value as LicenseSourceType", func() {
		const (
			CodeGenName        = "test-deployBuilder"
			CodeGenNamespace   = "default"
			LicenseSourceType  = "ConfigSecret"
			LicenseSourceValue = "testconfigsecret"
			key1               = "key1"
			key2               = "key2"
			value1             = "value1"
			value2             = "value2"
			tolerationKey      = "testTolerationKey"
			nodeAffinityWt     = int32(100)
			podAffinityWt      = int32(50)
		)
		It("Should produce Service and Deployment resource as intended", func() {
			serviceType := corev1.ServiceTypeLoadBalancer
			clusterIp := "10.1.1.1"
			externalTrafficPolicyType := corev1.ServiceExternalTrafficPolicyTypeCluster
			nodePort := int32(30000)
			loadBalancerIP := "10.1.1.2"
			loadBalancerClass := "testloadbalancerclass"
			servicePortName := "testName"
			imageName := "testImage"
			imagePullPolicy := corev1.PullIfNotPresent
			imagePullSecret := "pullSecret"
			terminationGracePeriodInSeconds := int64(30)
			serviceAccountName := "testAccount"
			schedulerName := "testScheduler"
			priorityClassName := "testPriorityClass"
			replicas := int32(3)
			securityContext := corev1.PodSecurityContext{
				SupplementalGroups: []int64{
					56,
				},
			}
			resources := corev1.ResourceRequirements{
				Limits:   corev1.ResourceList{},
				Requests: corev1.ResourceList{},
			}
			resources.Limits[corev1.ResourceCPU] = resource.Quantity{
				Format: resource.DecimalExponent,
			}
			resources.Requests[corev1.ResourceCPU] = resource.Quantity{
				Format: resource.DecimalSI,
			}
			nodeName := "testNodeName"
			nodeSelector := make(map[string]string)
			nodeSelector[key1] = value1
			nodeSelector[key2] = value2
			tolerations := []corev1.Toleration{{
				Key: tolerationKey,
			}}
			nodeAffinity := corev1.NodeAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{{
					Weight: nodeAffinityWt,
				}},
			}
			podAffinity := corev1.PodAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{{
					Weight: podAffinityWt,
				}},
			}
			serviceCustomAnnotations := make(map[string]string)
			serviceCustomAnnotations[key1] = value1
			serviceCustomAnnotations[key2] = value2
			labels := make(map[string]string)
			labels[key1] = value1
			labels[key2] = value2

			volumeSource := corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: LicenseSourceValue,
				},
			}

			codegen := &apicodegenv1beta2.CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNamespace,
				},
				Spec: apicodegenv1beta2.CodeGenSpec{
					Replicas: replicas,
					Labels:   labels,
					ServiceSpec: &apicodegenv1beta2.CodeGenServiceSpec{
						Type:                  serviceType,
						ClusterIP:             &clusterIp,
						ExternalTrafficPolicy: &externalTrafficPolicyType,
						CodeGenServicePort: &apicodegenv1beta2.CodeGenServicePort{
							Name:     servicePortName,
							NodePort: &nodePort,
						},
						LoadBalancerIP:           &loadBalancerIP,
						LoadBalancerClass:        &loadBalancerClass,
						ServiceCustomAnnotations: serviceCustomAnnotations,
					},
					PodSpec: &apicodegenv1beta2.CodeGenPodSpec{
						CodeGenContainerSpec: &apicodegenv1beta2.CodeGenContainerSpec{
							Image:           &imageName,
							ImagePullPolicy: &imagePullPolicy,
							ImagePullSecret: &imagePullSecret,
							Resources:       &resources,
						},
						TerminationGracePeriodSeconds: &terminationGracePeriodInSeconds,
						ServiceAccountName:            &serviceAccountName,
						SecurityContext:               &securityContext,
						SchedulerName:                 &schedulerName,
						PriorityClassName:             &priorityClassName,
					},
					LicenseSpec: apicodegenv1beta2.CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  LicenseSourceType,
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
					CodeGenPodPlacementSpec: &apicodegenv1beta2.CodeGenPodPlacementSpec{
						NodeName:     &nodeName,
						NodeSelector: nodeSelector,
						Tolerations:  tolerations,
						NodeAffinity: &nodeAffinity,
						PodAffinity:  &podAffinity,
					},
				},
			}

			genericResources := generateResources(codegen)
			Expect(len(genericResources)).Should(Equal(int(2)))

			var createdService *corev1.Service = serviceForCodeGen(codegen)
			Expect(createdService).ShouldNot(BeNil())
			Expect(createdService.Spec.Type).Should(Equal(corev1.ServiceType(serviceType)))
			Expect(createdService.Spec.Ports[0].Name).Should(Equal(string(servicePortName)))
			Expect(createdService.Spec.Ports[0].NodePort).Should(Equal(int32(nodePort)))
			Expect(createdService.Spec.Ports[0].Port).Should(Equal(int32(8080)))
			Expect(createdService.Spec.Ports[0].TargetPort).Should(Equal(intstr.FromInt(8080)))
			Expect(createdService.Spec.ExternalTrafficPolicy).Should(Equal(corev1.ServiceExternalTrafficPolicyType(externalTrafficPolicyType)))
			Expect(createdService.Spec.ClusterIP).Should(Equal(string(clusterIp)))
			Expect(createdService.Spec.LoadBalancerIP).Should(Equal(string(loadBalancerIP)))
			Expect(*createdService.Spec.LoadBalancerClass).Should(Equal(string(loadBalancerClass)))
			Expect(createdService.Name).Should(Equal(string(CodeGenName)))
			Expect(createdService.Namespace).Should(Equal(string(CodeGenNamespace)))
			Expect(createdService.Annotations[key1]).Should(Equal(string(value1)))
			Expect(createdService.Annotations[key2]).Should(Equal(string(value2)))
			Expect(createdService.Labels[key1]).Should(Equal(string(value1)))
			Expect(createdService.Labels[key2]).Should(Equal(string(value2)))
			Expect(createdService.Spec.Selector[key1]).Should(Equal(string(value1)))
			Expect(createdService.Spec.Selector[key2]).Should(Equal(string(value2)))

			var createdDeployment *appsv1.Deployment = deploymentForCodeGen(codegen)
			Expect(createdDeployment).ShouldNot(BeNil())
			Expect(createdDeployment.Name).Should(Equal(string(CodeGenName)))
			Expect(createdDeployment.Namespace).Should(Equal(string(CodeGenNamespace)))
			Expect(createdDeployment.Labels[key1]).Should(Equal(string(value1)))
			Expect(createdDeployment.Labels[key2]).Should(Equal(string(value2)))
			Expect(*createdDeployment.Spec.Replicas).Should(Equal(int32(replicas)))
			Expect(createdDeployment.Spec.Selector.MatchLabels[key1]).Should(Equal(string(value1)))
			Expect(createdDeployment.Spec.Selector.MatchLabels[key2]).Should(Equal(string(value2)))
			Expect(createdDeployment.Spec.Template.Labels[key1]).Should(Equal(string(value1)))
			Expect(createdDeployment.Spec.Template.Labels[key2]).Should(Equal(string(value2)))
			Expect(createdDeployment.Spec.Template.Spec.ServiceAccountName).Should(Equal(string(serviceAccountName)))
			Expect(*createdDeployment.Spec.Template.Spec.TerminationGracePeriodSeconds).Should(Equal(int64(terminationGracePeriodInSeconds)))
			Expect(createdDeployment.Spec.Template.Spec.ImagePullSecrets[0].Name).Should(Equal(string(imagePullSecret)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Image).Should(Equal(string(imageName)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy).Should(Equal(corev1.PullPolicy(imagePullPolicy)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Env[0].Name).Should(Equal(string(LICENSEPATHKEY)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Env[0].Value).Should(Equal(string(LICENSEPATHVALUE)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).Should(Equal(int32(8080)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU].Format).Should(Equal(resource.Format(resource.DecimalExponent)))
			Expect(*createdDeployment.Spec.Template.Spec.Containers[0].LivenessProbe).Should(Equal(corev1.Probe(ContainerProbe)))
			Expect(*createdDeployment.Spec.Template.Spec.Containers[0].ReadinessProbe).Should(Equal(corev1.Probe(ContainerProbe)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU].Format).Should(Equal(resource.Format(resource.DecimalSI)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name).Should(Equal(string(LICENSEVOLUME)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath).Should(Equal(string(LICENSEPATHVALUE)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].VolumeMounts[0].ReadOnly).Should(Equal(bool(true)))
			Expect(createdDeployment.Spec.Template.Spec.Volumes[0].Name).Should(Equal(string(LICENSEVOLUME)))
			Expect(createdDeployment.Spec.Template.Spec.Volumes[0].VolumeSource).Should(Equal(corev1.VolumeSource(volumeSource)))
			Expect(createdDeployment.Spec.Template.Spec.SecurityContext.SupplementalGroups[0]).Should(Equal(int64(56)))
			Expect(createdDeployment.Spec.Template.Spec.SchedulerName).Should(Equal(string(schedulerName)))
			Expect(createdDeployment.Spec.Template.Spec.PriorityClassName).Should(Equal(string(priorityClassName)))
			Expect(createdDeployment.Spec.Template.Spec.NodeName).Should(Equal(string(nodeName)))
			Expect(createdDeployment.Spec.Template.Spec.Tolerations[0].Key).Should(Equal(string(tolerationKey)))
			Expect(createdDeployment.Spec.Template.Spec.NodeSelector[key1]).Should(Equal(string(value1)))
			Expect(createdDeployment.Spec.Template.Spec.NodeSelector[key2]).Should(Equal(string(value2)))
			Expect(createdDeployment.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].Weight).Should(Equal(int32(nodeAffinityWt)))
			Expect(createdDeployment.Spec.Template.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].Weight).Should(Equal(int32(podAffinityWt)))
		})
	})

})
