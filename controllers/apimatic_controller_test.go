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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	apicodegenv1beta2 "github.com/apimatic/apimatic-codegen-operator/api/v1beta2"
)

var _ = Describe("Controller", func() {

	Context("When processing CodeGen instance in k8s API Server with minimal valid specifications", func() {
		const (
			CodeGenName      = "test-controller"
			CodeGenNameSpace = "default"
			timeout          = time.Second * 60
			duration         = time.Second * 5
			interval         = time.Millisecond * 250

			LicenseSourceType  = "LicenseBlob"
			LicenseSourceValue = "12345"
		)
		It("Should be created successfully", func() {
			By("Creating a new CodeGen instance with minimal specifications")
			log := log.FromContext(ctx)
			codegen := &apicodegenv1beta2.CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNameSpace,
				},
				Spec: apicodegenv1beta2.CodeGenSpec{
					LicenseSpec: apicodegenv1beta2.CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  LicenseSourceType,
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
				},
			}

			log.Info("Creating CodeGen resource and triggering reconcile function")
			Expect(k8sClient.Create(ctx, codegen)).Should(Succeed())

			// Wait for the reconciler loop to finish before accessing the updated deployment and service resources
			time.Sleep(time.Second * 20)

			codegenLookupKey := types.NamespacedName{Name: CodeGenName, Namespace: CodeGenNameSpace}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, codegenLookupKey, codegen)
				return err == nil
			}, timeout, duration).Should(BeTrue())

			Expect(codegen.Spec.Replicas).Should(Equal(int32(1)))
			Expect(*codegen.Spec.PodSpec.TerminationGracePeriodSeconds).Should(Equal(int64(30)))
			Expect(codegen.Spec.PodSpec.ServiceAccountName).Should(BeNil())
			Expect(codegen.Spec.PodSpec.SecurityContext).Should(BeNil())
			Expect(codegen.Spec.PodSpec.SchedulerName).Should(BeNil())
			Expect(codegen.Spec.PodSpec.PriorityClassName).Should(BeNil())
			Expect(*codegen.Spec.PodSpec.CodeGenContainerSpec.Image).Should(Equal(string(apicodegenv1beta2.DEFAULTIMAGENAME)))

			Expect(codegen.Spec.LicenseSpec.CodeGenLicenseSourceType).Should(Equal(string(LicenseSourceType)))
			Expect(codegen.Spec.LicenseSpec.CodeGenLicenseSourceValue).Should(Equal(string(LicenseSourceValue)))

			Expect(codegen.Spec.ServiceSpec.Type).Should(Equal(corev1.ServiceType(corev1.ServiceTypeClusterIP)))
			Expect(codegen.Spec.ServiceSpec.ClusterIP).Should(BeNil())
			Expect(codegen.Spec.ServiceSpec.ExternalTrafficPolicy).Should(BeNil())
			Expect(codegen.Spec.ServiceSpec.CodeGenServicePort.Name).Should(Equal(string("codegen")))
			Expect(codegen.Spec.ServiceSpec.ServiceCustomAnnotations).Should(BeNil())

			Expect(codegen.Spec.Labels).ShouldNot(BeNil())
			Expect(len(codegen.Spec.Labels)).Should(Equal(int(2)))
			val, isKeyExists := codegen.Spec.Labels[apicodegenv1beta2.DEFAULTLABELKEY1]
			val2, isKeyExists2 := codegen.Spec.Labels[apicodegenv1beta2.DEFAULTLABELKEY2]
			Expect(isKeyExists).Should(Equal(bool(true)))
			Expect(isKeyExists2).Should(Equal(bool(true)))
			Expect(val).Should(Equal(string(apicodegenv1beta2.DEFAULTLABELVALUE1)))
			Expect(val2).Should(Equal(string(CodeGenName)))

			Expect(codegen.Status.Replicas).Should(Equal(int32(1)))

			createdDeployment := &appsv1.Deployment{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, codegenLookupKey, createdDeployment)
				return err == nil
			}, timeout, duration).Should(BeTrue())
			Expect(*createdDeployment.Spec.Replicas).Should(Equal(int32(1)))
			Expect(createdDeployment.ObjectMeta.OwnerReferences[len(createdDeployment.ObjectMeta.OwnerReferences)-1].Name).Should(Equal(string(CodeGenName)))

			createdService := &corev1.Service{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, codegenLookupKey, createdService)
				return err == nil
			}, timeout, duration).Should(BeTrue())
			Expect(createdService.Spec.Type).Should(Equal(corev1.ServiceTypeClusterIP))
			Expect(createdService.ObjectMeta.OwnerReferences[len(createdService.ObjectMeta.OwnerReferences)-1].Name).Should(Equal(string(CodeGenName)))

			// Cleanup
			log.Info("Deleting CodeGen resource", "Name", codegenLookupKey.Name, "Namespace", codegenLookupKey.Namespace)
			Expect(k8sClient.Delete(ctx, codegen)).Should(Succeed())
			// We have to manually delete service and deployment resources during test as garbage collection is done by Kubelet
			// which is not available in TestEnv. It only deploys API server and etcd
			Expect(k8sClient.Delete(ctx, createdService)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, createdDeployment)).Should(Succeed())
		})
	})

	Context("When processing CodeGen instance in k8s API Server with new fields", func() {
		const (
			CodeGenName      = "test-controller2"
			CodeGenNameSpace = "default"
			timeout          = time.Second * 60
			duration         = time.Second * 5
			interval         = time.Millisecond * 250

			LicenseSourceValue = "12345"

			key1   = "key1"
			key2   = "key2"
			value1 = "value1"
			value2 = "value2"
		)
		It("Should be created successfully and be updated successfully with new  fields", func() {
			By("Creating a new CodeGen instance with specifications and then updating fields")
			loadbalancerIP := "10.1.1.1"
			nodePort := int32(31000)
			name := "name1"
			annotations := map[string]string{
				key1: value1,
				key2: value2,
			}
			imageName := "testImage"
			pullPolicy := corev1.PullAlways
			pullSecret := "testSecret"
			terminationGracePeriod := int64(15)
			serviceAccountName := "example1.com"
			schedulerName := "testScheduler"
			priorityClassName := "woooo.edu"
			runAsNonRoot := true
			nodeSelector := map[string]string{
				key1: value1,
				key2: value2,
			}
			nodeName := "nodename.my"
			tolerations := []corev1.Toleration{{
				Key:      key1,
				Operator: "Equal",
				Value:    value1,
			}}
			nodeAffinity := corev1.NodeAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{{
					Weight: int32(100),
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key:      key1,
							Operator: "In",
							Values:   []string{"a", "b", "c"},
						}},
					},
				}},
			}
			podAffinity := corev1.PodAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{{
					Weight: int32(50),
					PodAffinityTerm: corev1.PodAffinityTerm{
						TopologyKey: key1,
					},
				}},
			}
			labels := map[string]string{
				key1: value1,
				key2: value2,
			}

			log := log.FromContext(ctx)
			codegen := &apicodegenv1beta2.CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNameSpace,
				},
				Spec: apicodegenv1beta2.CodeGenSpec{
					Labels: labels,
					ServiceSpec: &apicodegenv1beta2.CodeGenServiceSpec{
						Type:                     corev1.ServiceTypeLoadBalancer,
						LoadBalancerIP:           &loadbalancerIP,
						ServiceCustomAnnotations: annotations,
						CodeGenServicePort: &apicodegenv1beta2.CodeGenServicePort{
							NodePort: &nodePort,
							Name:     name,
						},
					},
					PodSpec: &apicodegenv1beta2.CodeGenPodSpec{
						CodeGenContainerSpec: &apicodegenv1beta2.CodeGenContainerSpec{
							Image:           &imageName,
							ImagePullPolicy: &pullPolicy,
							ImagePullSecret: &pullSecret,
						},
						TerminationGracePeriodSeconds: &terminationGracePeriod,
						ServiceAccountName:            &serviceAccountName,
						SchedulerName:                 &schedulerName,
						PriorityClassName:             &priorityClassName,
						SecurityContext: &corev1.PodSecurityContext{
							RunAsNonRoot: &runAsNonRoot,
						},
					},
					LicenseSpec: apicodegenv1beta2.CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  "ConfigMap",
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
					CodeGenPodPlacementSpec: &apicodegenv1beta2.CodeGenPodPlacementSpec{
						NodeSelector: nodeSelector,
						NodeName:     &nodeName,
						Tolerations:  tolerations,
						NodeAffinity: &nodeAffinity,
						PodAffinity:  &podAffinity,
					},
				},
			}

			log.Info("Creating CodeGen resource and triggering reconcile function")
			Expect(k8sClient.Create(ctx, codegen)).Should(Succeed())

			// Wait for the reconciler loop to finish before accessing the updated deployment and service resources
			time.Sleep(time.Second * 20)

			codegenLookupKey := types.NamespacedName{Name: CodeGenName, Namespace: CodeGenNameSpace}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, codegenLookupKey, codegen)
				return err == nil
			}, timeout, duration).Should(BeTrue())

			createdDeployment := &appsv1.Deployment{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, codegenLookupKey, createdDeployment)
				return err == nil
			}, timeout, duration).Should(BeTrue())
			Expect(*createdDeployment.Spec.Replicas).Should(Equal(int32(1)))
			Expect(createdDeployment.Labels).Should(Equal(map[string]string(labels)))
			Expect(createdDeployment.Spec.Selector.MatchLabels).Should(Equal(map[string]string(labels)))
			Expect(createdDeployment.Spec.Template.Labels).Should(Equal(map[string]string(labels)))
			Expect(*createdDeployment.Spec.Template.Spec.TerminationGracePeriodSeconds).Should(Equal(int64(terminationGracePeriod)))
			Expect(createdDeployment.Spec.Template.Spec.ServiceAccountName).Should(Equal(string(serviceAccountName)))
			Expect(createdDeployment.Spec.Template.Spec.SchedulerName).Should(Equal(string(schedulerName)))
			Expect(createdDeployment.Spec.Template.Spec.PriorityClassName).Should(Equal(string(priorityClassName)))
			Expect(*createdDeployment.Spec.Template.Spec.SecurityContext.RunAsNonRoot).Should(Equal(runAsNonRoot))
			Expect(createdDeployment.Spec.Template.Spec.NodeName).Should(Equal(string(nodeName)))
			Expect(createdDeployment.Spec.Template.Spec.NodeSelector).Should(Equal(map[string]string(nodeSelector)))
			Expect(createdDeployment.Spec.Template.Spec.Tolerations).Should(Equal([]corev1.Toleration(tolerations)))
			Expect(*createdDeployment.Spec.Template.Spec.Affinity.NodeAffinity).Should(Equal(corev1.NodeAffinity(nodeAffinity)))
			Expect(*createdDeployment.Spec.Template.Spec.Affinity.PodAffinity).Should(Equal(corev1.PodAffinity(podAffinity)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Image).Should(Equal(string(imageName)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy).Should(Equal(corev1.PullPolicy(pullPolicy)))
			Expect(createdDeployment.Spec.Template.Spec.ImagePullSecrets[0].Name).Should(Equal(string(pullSecret)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].VolumeMounts).ShouldNot(BeNil())
			Expect(createdDeployment.Spec.Template.Spec.Volumes).ShouldNot(BeNil())
			Expect(createdDeployment.ObjectMeta.OwnerReferences[len(createdDeployment.ObjectMeta.OwnerReferences)-1].Name).Should(Equal(string(CodeGenName)))

			createdService := &corev1.Service{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, codegenLookupKey, createdService)
				return err == nil
			}, timeout, duration).Should(BeTrue())
			Expect(createdService.Labels).Should(Equal(map[string]string(labels)))
			Expect(createdService.Spec.Type).Should(Equal(corev1.ServiceTypeLoadBalancer))
			Expect(createdService.Spec.Selector).Should(Equal(map[string]string(labels)))
			Expect(createdService.Spec.LoadBalancerIP).Should(Equal(string(loadbalancerIP)))
			Expect(createdService.Spec.Ports[0].NodePort).Should(Equal(int32(nodePort)))
			Expect(createdService.Spec.Ports[0].Name).Should(Equal(string(name)))
			Expect(createdService.Annotations[key1]).Should(Equal(value1))
			Expect(createdService.Annotations[key2]).Should(Equal(value2))
			Expect(createdService.ObjectMeta.OwnerReferences[len(createdService.ObjectMeta.OwnerReferences)-1].Name).Should(Equal(string(CodeGenName)))

			log.Info("Updating specs")
			// We GET the updated CodeGen resource from API Server with latest version to avoid 409 conflict error
			Expect(k8sClient.Get(ctx, codegenLookupKey, codegen)).Should(Succeed())
			loadbalancerIP2 := "10.1.1.2"
			nodePort2 := int32(31001)
			name2 := "name2"
			imageName2 := "testImage2"
			pullPolicy2 := corev1.PullNever
			pullSecret2 := "pullSecret2"
			terminationGracePeriod2 := int64(15)
			serviceAccountName2 := "example2.com"
			priorityClassName2 := "yayyy.com"
			schedulerName2 := "scheduler2"
			runAsNonRoot2 := false
			nodeSelector2 := map[string]string{
				key1: value2,
				key2: value1,
			}
			nodeName2 := "nodename2.my"
			tolerations2 := []corev1.Toleration{{
				Key:      key1,
				Operator: "Equal",
				Value:    value2,
			}}
			nodeAffinity2 := corev1.NodeAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{{
					Weight: int32(50),
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key:      key2,
							Operator: "In",
							Values:   []string{"a2", "b2", "c2"},
						}},
					},
				}},
			}
			podAffinity2 := corev1.PodAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{{
					Weight: int32(20),
					PodAffinityTerm: corev1.PodAffinityTerm{
						TopologyKey: key2,
					},
				}},
			}
			annotations2 := map[string]string{
				value1: key1,
			}
			codegen.Spec.Replicas = 4
			codegen.Spec.ServiceSpec.LoadBalancerIP = &loadbalancerIP2
			codegen.Spec.ServiceSpec.CodeGenServicePort.Name = name2
			codegen.Spec.ServiceSpec.CodeGenServicePort.NodePort = &nodePort2
			codegen.Spec.ServiceSpec.ServiceCustomAnnotations = annotations2
			codegen.Spec.PodSpec.CodeGenContainerSpec.Image = &imageName2
			codegen.Spec.PodSpec.CodeGenContainerSpec.ImagePullPolicy = &pullPolicy2
			codegen.Spec.PodSpec.CodeGenContainerSpec.ImagePullSecret = &pullSecret2
			codegen.Spec.PodSpec.TerminationGracePeriodSeconds = &terminationGracePeriod2
			codegen.Spec.PodSpec.ServiceAccountName = &serviceAccountName2
			codegen.Spec.PodSpec.SchedulerName = &schedulerName2
			codegen.Spec.PodSpec.PriorityClassName = &priorityClassName2
			codegen.Spec.PodSpec.SecurityContext = &corev1.PodSecurityContext{
				RunAsNonRoot: &runAsNonRoot2,
			}
			codegen.Spec.CodeGenPodPlacementSpec.NodeAffinity = &nodeAffinity2
			codegen.Spec.CodeGenPodPlacementSpec.NodeName = &nodeName2
			codegen.Spec.CodeGenPodPlacementSpec.NodeSelector = nodeSelector2
			codegen.Spec.CodeGenPodPlacementSpec.PodAffinity = &podAffinity2
			codegen.Spec.CodeGenPodPlacementSpec.Tolerations = tolerations2
			codegen.Spec.LicenseSpec.CodeGenLicenseSourceType = "LicenseBlob"

			Expect(k8sClient.Update(ctx, codegen)).Should(Succeed())

			// Wait for the reconciler loop to finish before accessing the updated deployment and service resources
			time.Sleep(time.Second * 20)

			Eventually(func() bool {
				err := k8sClient.Get(ctx, codegenLookupKey, codegen)
				return err == nil
			}, timeout, duration).Should(BeTrue())

			Expect(codegen.Status.Replicas).Should(Equal(int32(4)))

			Eventually(func() bool {
				err := k8sClient.Get(ctx, codegenLookupKey, createdDeployment)
				return err == nil
			}, timeout, duration).Should(BeTrue())
			Expect(*createdDeployment.Spec.Replicas).Should(Equal(int32(4)))
			Expect(*createdDeployment.Spec.Template.Spec.TerminationGracePeriodSeconds).Should(Equal(int64(terminationGracePeriod2)))
			Expect(createdDeployment.Spec.Template.Spec.ServiceAccountName).Should(Equal(string(serviceAccountName2)))
			Expect(createdDeployment.Spec.Template.Spec.SchedulerName).Should(Equal(string(schedulerName2)))
			Expect(createdDeployment.Spec.Template.Spec.PriorityClassName).Should(Equal(string(priorityClassName2)))
			Expect(*createdDeployment.Spec.Template.Spec.SecurityContext.RunAsNonRoot).Should(Equal(runAsNonRoot2))
			Expect(createdDeployment.Spec.Template.Spec.NodeName).Should(Equal(string(nodeName2)))
			Expect(createdDeployment.Spec.Template.Spec.NodeSelector).Should(Equal(map[string]string(nodeSelector2)))
			Expect(createdDeployment.Spec.Template.Spec.Tolerations).Should(Equal([]corev1.Toleration(tolerations2)))
			Expect(*createdDeployment.Spec.Template.Spec.Affinity.NodeAffinity).Should(Equal(corev1.NodeAffinity(nodeAffinity2)))
			Expect(*createdDeployment.Spec.Template.Spec.Affinity.PodAffinity).Should(Equal(corev1.PodAffinity(podAffinity2)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Image).Should(Equal(string(imageName2)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy).Should(Equal(corev1.PullPolicy(pullPolicy2)))
			Expect(createdDeployment.Spec.Template.Spec.ImagePullSecrets[0].Name).Should(Equal(string(pullSecret2)))
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].VolumeMounts).Should(BeNil())
			Expect(createdDeployment.Spec.Template.Spec.Volumes).Should(BeNil())
			Expect(createdDeployment.ObjectMeta.OwnerReferences[len(createdDeployment.ObjectMeta.OwnerReferences)-1].Name).Should(Equal(string(CodeGenName)))

			Eventually(func() bool {
				err := k8sClient.Get(ctx, codegenLookupKey, createdService)
				return err == nil
			}, timeout, duration).Should(BeTrue())

			_, exist := createdService.Annotations[key1]
			_, exist2 := createdService.Annotations[key2]

			Expect(exist).Should(Equal(bool(false)))
			Expect(exist2).Should(Equal(bool(false)))
			Expect(len(createdService.Annotations)).Should(Equal(int(1)))
			Expect(createdService.Annotations[value1]).Should(Equal(string(key1)))

			Expect(createdService.Spec.LoadBalancerIP).Should(Equal(string(loadbalancerIP2)))
			Expect(createdService.Spec.Ports[0].NodePort).Should(Equal(int32(nodePort2)))
			Expect(createdService.Spec.Ports[0].Name).Should(Equal(string(name2)))
			Expect(createdService.ObjectMeta.OwnerReferences[len(createdService.ObjectMeta.OwnerReferences)-1].Name).Should(Equal(string(CodeGenName)))

			// Cleanup
			log.Info("Deleting CodeGen resource", "Name", codegenLookupKey.Name, "Namespace", codegenLookupKey.Namespace)
			Expect(k8sClient.Delete(ctx, codegen)).Should(Succeed())
			// We have to manually delete service and deployment resources during test as garbage collection is done by Kubelet
			// which is not available in TestEnv. It only deploys API server and etcd
			Expect(k8sClient.Delete(ctx, createdService)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, createdDeployment)).Should(Succeed())
		})
	})

})
