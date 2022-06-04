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
	"context"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Mutation Webhook", func() {
	const (
		CodeGenName      = "test-codegen-mutation-webhook"
		CodeGenNameSpace = "default"
		timeout          = time.Second * 20
		duration         = time.Second * 5
		interval         = time.Millisecond * 250

		LicenseSourceType  = "LicenseBlob"
		LicenseSourceValue = "12345"

		ServicePort = 8080
	)

	AfterEach(func() {
		codegenLookupKey := types.NamespacedName{Name: CodeGenName, Namespace: CodeGenNameSpace}
		codegenCR := &CodeGen{}
		err := k8sClient.Get(context.Background(), codegenLookupKey, codegenCR)

		if err == nil {
			Expect(k8sClient.Delete(context.Background(), codegenCR)).Should(Succeed())
		}
	})

	Context("When uploading CodeGen instance to k8s API Server with minimal valid specifications", func() {
		It("Should be created successfully", func() {
			By("Creating a new CodeGen instance with specifications")
			ctx := context.Background()

			codegen := &CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNameSpace,
				},
				Spec: CodeGenSpec{
					LicenseSpec: CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  LicenseSourceType,
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
				},
			}

			Expect(k8sClient.Create(ctx, codegen)).Should(Succeed())

			codegenLookupKey := types.NamespacedName{Name: CodeGenName, Namespace: CodeGenNameSpace}
			createdCodeGenCR := &CodeGen{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, codegenLookupKey, createdCodeGenCR)
				return err == nil
			}, timeout, duration).Should(BeTrue())

			Expect(createdCodeGenCR.Spec.Replicas).Should(Equal(int32(1)))
			Expect(*createdCodeGenCR.Spec.PodSpec.TerminationGracePeriodSeconds).Should(Equal(int64(30)))
			Expect(createdCodeGenCR.Spec.PodSpec.ServiceAccountName).Should(BeNil())
			Expect(createdCodeGenCR.Spec.PodSpec.SecurityContext).Should(BeNil())
			Expect(createdCodeGenCR.Spec.PodSpec.SchedulerName).Should(BeNil())
			Expect(createdCodeGenCR.Spec.PodSpec.PriorityClassName).Should(BeNil())
			Expect(*createdCodeGenCR.Spec.PodSpec.CodeGenContainerSpec.Image).Should(Equal(string(DEFAULTIMAGENAME)))

			Expect(createdCodeGenCR.Spec.LicenseSpec.CodeGenLicenseSourceType).Should(Equal(string(LicenseSourceType)))
			Expect(createdCodeGenCR.Spec.LicenseSpec.CodeGenLicenseSourceValue).Should(Equal(string(LicenseSourceValue)))

			Expect(createdCodeGenCR.Spec.ServiceSpec.Type).Should(Equal(corev1.ServiceType(corev1.ServiceTypeClusterIP)))
			Expect(createdCodeGenCR.Spec.ServiceSpec.ClusterIP).Should(BeNil())
			Expect(createdCodeGenCR.Spec.ServiceSpec.ExternalTrafficPolicy).Should(BeNil())
			Expect(createdCodeGenCR.Spec.ServiceSpec.CodeGenServicePort.Name).Should(Equal(string("codegen")))
			Expect(createdCodeGenCR.Spec.ServiceSpec.ServiceCustomAnnotations).Should(BeNil())
		})
	})
})

var _ = Describe("Validation webhook", func() {
	const (
		CodeGenName        = "test-codegen-validation-webhook"
		CodeGenNameSpace   = "default"
		LicenseSourceType  = "LicenseBlob"
		LicenseSourceValue = "12345"
		timeout            = time.Second * 60
		duration           = time.Second * 5
		interval           = time.Millisecond * 250
	)

	AfterEach(func() {
		codegenLookupKey := types.NamespacedName{Name: CodeGenName, Namespace: CodeGenNameSpace}
		codegenCR := &CodeGen{}
		err := k8sClient.Get(context.Background(), codegenLookupKey, codegenCR)

		if err == nil {
			Expect(k8sClient.Delete(context.Background(), codegenCR)).Should(Succeed())
		}
	})

	Context("When creating an CodeGen instance with invalid PodSecurityContext", func() {
		It("Should give a validation error response", func() {
			By("Creating a new CodeGen instance with specifications")
			ctx := context.Background()
			userName := "testUser"
			userId := int64(-1)
			groupId := int64(-1)
			f5GroupId := int64(-4)
			f5GroupChangePolicy := corev1.PodFSGroupChangePolicy("brad")
			codegen := &CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNameSpace,
				},
				Spec: CodeGenSpec{
					PodSpec: &CodeGenPodSpec{
						SecurityContext: &corev1.PodSecurityContext{
							RunAsUser:  &userId,
							RunAsGroup: &groupId,
							SupplementalGroups: []int64{
								int64(-2),
								int64(-3),
							},
							FSGroup: &f5GroupId,
							Sysctls: []corev1.Sysctl{{
								Name:  "first",
								Value: "1st",
							}, {
								Name:  "first",
								Value: "1st",
							}, {
								Value: "2nd",
							}, {
								Name:  strings.Repeat("repeat", 50),
								Value: "third",
							}},
							FSGroupChangePolicy: &f5GroupChangePolicy,
							SeccompProfile: &corev1.SeccompProfile{
								Type: "",
							},
							WindowsOptions: &corev1.WindowsSecurityContextOptions{
								RunAsUserName: &userName,
							},
						},
					},
					LicenseSpec: CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  LicenseSourceType,
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
				},
			}

			codegen.Default()
			err := codegen.ValidateCreate()
			Expect(err).ShouldNot(BeNil())
			Expect(k8sClient.Create(ctx, codegen)).ShouldNot(Succeed())
		})
	})

	Context("When creating a CodeGen instance with invalid ResourceRequirements", func() {
		It("Should give a validation error response", func() {
			By("Creating a new CodeGen instance with specifications")
			ctx := context.Background()
			codegen := &CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNameSpace,
				},
				Spec: CodeGenSpec{
					PodSpec: &CodeGenPodSpec{
						CodeGenContainerSpec: &CodeGenContainerSpec{
							Resources: &corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:              resource.MustParse("-501m"),
									corev1.ResourceMemory:           resource.MustParse("5Gi"),
									corev1.ResourceName("badValue"): resource.MustParse("8m"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("501m"),
									corev1.ResourceMemory: resource.MustParse("6Gi"),
								},
							},
						},
					},
					LicenseSpec: CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  LicenseSourceType,
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
				},
			}

			codegen.Default()
			err := codegen.ValidateCreate()
			Expect(err).ShouldNot(BeNil())
			Expect(k8sClient.Create(ctx, codegen)).ShouldNot(Succeed())
		})
	})

	Context("When creating a CodeGen instance with invalid PodPlacementSpec", func() {
		It("Should give a validation error response", func() {
			By("Creating a new CodeGen instance with specifications")
			tolerationSeconds := int64(23)
			ctx := context.Background()
			codegen := &CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNameSpace,
				},
				Spec: CodeGenSpec{
					CodeGenPodPlacementSpec: &CodeGenPodPlacementSpec{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{{
									MatchExpressions: []corev1.NodeSelectorRequirement{{
										Key:      "a",
										Operator: "Invalid",
										Values:   []string{"x", "y", "z"},
									}, {
										Key:      "a2",
										Operator: "In",
									}, {
										Key:      "a3",
										Operator: "Exists",
										Values:   []string{"x", "y"},
									}, {
										Key:      "a4",
										Operator: "Gt",
										Values:   []string{"4", "5"},
									}},
								}},
							},
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{{
								Weight: int32(120),
							}},
						},
						PodAffinity: &corev1.PodAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{{
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"1": "2",
									},
									MatchExpressions: []metav1.LabelSelectorRequirement{{
										Key:      "x",
										Operator: "Yummy",
										Values:   []string{"1", "2"},
									}, {
										Key:      "g",
										Operator: "Exists",
										Values:   []string{"k"},
									}, {
										Key:      "rrr",
										Operator: "In",
									}, {
										Key:      "_",
										Operator: "NotIn",
									}, {
										Key:      "rrr",
										Operator: "DoesNotExist",
										Values:   []string{"k"},
									}},
								},
							}},
						},
						NodeSelector: map[string]string{
							"/hhhh": "brad",
						},
						Tolerations: []corev1.Toleration{{
							Key:               "3",
							Operator:          "random",
							Value:             "blech",
							TolerationSeconds: &tolerationSeconds,
							Effect:            corev1.TaintEffectNoSchedule,
						}},
					},
					LicenseSpec: CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  LicenseSourceType,
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
				},
			}

			codegen.Default()
			err := codegen.ValidateCreate()
			Expect(err).ShouldNot(BeNil())
			Expect(k8sClient.Create(ctx, codegen)).ShouldNot(Succeed())
		})
	})

	Context("When creating a CodeGen instance with LoadBalancerIP set to an invalid value and Service set to type LoadBalancer", func() {
		It("Should give a validation error response", func() {
			By("Creating a new CodeGen instance with specifications")
			ctx := context.Background()
			var loadBalancerIP string = "10.10.10"
			codegen := &CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNameSpace,
				},
				Spec: CodeGenSpec{
					Replicas: 3,
					ServiceSpec: &CodeGenServiceSpec{
						Type:           corev1.ServiceTypeLoadBalancer,
						LoadBalancerIP: &loadBalancerIP,
					},
					LicenseSpec: CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  LicenseSourceType,
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
				},
			}

			codegen.Default()
			err := codegen.ValidateCreate()
			Expect(err).ShouldNot(BeNil())
			Expect(k8sClient.Create(ctx, codegen)).ShouldNot(Succeed())
		})
	})

	Context("When creating a CodeGen instance with LoadBalancerIP set and Service not of type LoadBalancer", func() {
		It("Should give a validation error response", func() {
			By("Creating a new CodeGen instance with specifications")
			ctx := context.Background()
			var loadBalancerIP string = "10.10.10.10"
			codegen := &CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNameSpace,
				},
				Spec: CodeGenSpec{
					Replicas: 3,
					ServiceSpec: &CodeGenServiceSpec{
						Type:           corev1.ServiceTypeNodePort,
						LoadBalancerIP: &loadBalancerIP,
					},
					LicenseSpec: CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  LicenseSourceType,
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
				},
			}

			codegen.Default()
			err := codegen.ValidateCreate()
			Expect(err).ShouldNot(BeNil())
			Expect(k8sClient.Create(ctx, codegen)).ShouldNot(Succeed())
		})
	})

	Context("When creating a CodeGen instance with ExternalTrafficPolicy set and Service not of type LoadBalancer or NodePort", func() {
		It("Should give a validation error response", func() {
			By("Creating a new CodeGen instance with specifications")
			ctx := context.Background()
			var externalTrafficePolicy corev1.ServiceExternalTrafficPolicyType = corev1.ServiceExternalTrafficPolicyTypeLocal
			codegen := &CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNameSpace,
				},
				Spec: CodeGenSpec{
					Replicas: 3,
					ServiceSpec: &CodeGenServiceSpec{
						Type:                  corev1.ServiceTypeClusterIP,
						ExternalTrafficPolicy: &externalTrafficePolicy,
					},
					LicenseSpec: CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  LicenseSourceType,
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
				},
			}

			codegen.Default()
			err := codegen.ValidateCreate()
			Expect(err).ShouldNot(BeNil())
			Expect(k8sClient.Create(ctx, codegen)).ShouldNot(Succeed())
		})
	})

	Context("When creating a CodeGen instance with nodeport set in APIMaticServicePort and Service not of type LoadBalancer or NodePort", func() {
		It("Should give a validation error response", func() {
			By("Creating a new CodeGen instance with specifications")
			ctx := context.Background()
			var nodePort int32 = 30000
			codegen := &CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNameSpace,
				},
				Spec: CodeGenSpec{
					Replicas: 3,
					ServiceSpec: &CodeGenServiceSpec{
						Type: corev1.ServiceTypeClusterIP,
						CodeGenServicePort: &CodeGenServicePort{
							NodePort: &nodePort,
						},
					},
					LicenseSpec: CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  LicenseSourceType,
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
				},
			}

			codegen.Default()
			err := codegen.ValidateCreate()
			Expect(err).ShouldNot(BeNil())
			Expect(k8sClient.Create(ctx, codegen)).ShouldNot(Succeed())
		})
	})

	Context("When creating a CodeGen instance with clusterIP set to 'None' and Service Type ClusterIP", func() {
		It("Should not give a validation error response", func() {
			By("Creating a new CodeGen instance with specifications")
			ctx := context.Background()
			var clusterIP string = "None"
			codegen := &CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNameSpace,
				},
				Spec: CodeGenSpec{
					Replicas: 3,
					ServiceSpec: &CodeGenServiceSpec{
						ClusterIP: &clusterIP,
						Type:      corev1.ServiceTypeClusterIP,
					},
					LicenseSpec: CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  LicenseSourceType,
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
				},
			}
			codegen.Default()
			err := codegen.ValidateCreate()
			Expect(err).Should(BeNil())
			Expect(k8sClient.Create(ctx, codegen)).Should(Succeed())
		})
	})

	Context("When creating a CodeGen instance with clusterIP set to 'None' and Type set to LoadBalancer", func() {
		It("Should give a validation error response", func() {
			By("Creating a new CodeGen instance with specifications")
			ctx := context.Background()
			var clusterIP string = "None"
			codegen := &CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNameSpace,
				},
				Spec: CodeGenSpec{
					Replicas: 3,
					ServiceSpec: &CodeGenServiceSpec{
						ClusterIP: &clusterIP,
						Type:      corev1.ServiceTypeLoadBalancer,
					},
					LicenseSpec: CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  LicenseSourceType,
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
				},
			}
			codegen.Default()
			err := codegen.ValidateCreate()
			Expect(err).ShouldNot(BeNil())
			Expect(k8sClient.Create(ctx, codegen)).ShouldNot(Succeed())
		})
	})

	Context("When creating an CodeGen instance with clusterIP set to valid IP address", func() {
		It("Should not give a validation error response", func() {
			By("Creating a new CodeGen instance with specifications")
			ctx := context.Background()
			var clusterIP string = "1.2.3.4"
			codegen := &CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNameSpace,
				},
				Spec: CodeGenSpec{
					Replicas: 3,
					ServiceSpec: &CodeGenServiceSpec{
						ClusterIP: &clusterIP,
					},
					LicenseSpec: CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  LicenseSourceType,
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
				},
			}
			codegen.Default()
			err := codegen.ValidateCreate()
			Expect(err).Should(BeNil())
			Expect(k8sClient.Create(ctx, codegen)).Should(Succeed())
		})
	})

	Context("When creating a CodeGen instance with clusterIP set to invalid IP address", func() {
		It("Should give a validation error response", func() {
			By("Creating a new CodeGen instance with specifications")
			ctx := context.Background()
			var clusterIP string = "1.2.34"
			codegen := &CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNameSpace,
				},
				Spec: CodeGenSpec{
					Replicas: 3,
					ServiceSpec: &CodeGenServiceSpec{
						ClusterIP: &clusterIP,
					},
					LicenseSpec: CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  LicenseSourceType,
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
				},
			}
			codegen.Default()
			err := codegen.ValidateCreate()
			Expect(err).ShouldNot(BeNil())
			Expect(k8sClient.Create(ctx, codegen)).ShouldNot(Succeed())
		})
	})

	Context("When creating a CodeGen instance with service type ClusterIP and NodePort set", func() {
		It("Should give a validation error", func() {
			ctx := context.Background()
			var nodePort int32 = 31000
			codegen := &CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNameSpace,
				},
				Spec: CodeGenSpec{
					Replicas: 3,
					ServiceSpec: &CodeGenServiceSpec{
						CodeGenServicePort: &CodeGenServicePort{
							NodePort: &nodePort,
						},
					},
					LicenseSpec: CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  LicenseSourceType,
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
				},
			}

			codegen.Default()
			err := codegen.ValidateCreate()
			Expect(err).ShouldNot(BeNil())
			Expect(k8sClient.Create(ctx, codegen)).ShouldNot(Succeed())
		})
	})

	Context("When creating a CodeGen instance with service type ClusterIP and LoadBalancerClass set", func() {
		It("Should give a validation error", func() {
			ctx := context.Background()
			var loadBalancerClass string = "test"
			codegen := &CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNameSpace,
				},
				Spec: CodeGenSpec{
					Replicas: 3,
					ServiceSpec: &CodeGenServiceSpec{
						Type:              corev1.ServiceTypeClusterIP,
						LoadBalancerClass: &loadBalancerClass,
					},
					LicenseSpec: CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  LicenseSourceType,
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
				},
			}

			codegen.Default()
			err := codegen.ValidateCreate()
			Expect(err).ShouldNot(BeNil())
			Expect(k8sClient.Create(ctx, codegen)).ShouldNot(Succeed())
		})
	})

	Context("When updating a CodeGen instance by changing clusterIP", func() {
		It("Should give a validation error response", func() {
			By("Creating a new CodeGen instance with specifications")
			ctx := context.Background()
			var clusterIP string = "1.2.3.4"
			codegen := &CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNameSpace,
				},
				Spec: CodeGenSpec{
					Replicas: 3,
					ServiceSpec: &CodeGenServiceSpec{
						ClusterIP: &clusterIP,
					},
					LicenseSpec: CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  LicenseSourceType,
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
				},
			}
			codegen.Default()
			err := codegen.ValidateCreate()
			Expect(err).Should(BeNil())
			Expect(k8sClient.Create(ctx, codegen)).Should(Succeed())

			codegenLookupKey := types.NamespacedName{Name: CodeGenName, Namespace: CodeGenNameSpace}
			codegenAPIMaticCR := &CodeGen{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, codegenLookupKey, codegenAPIMaticCR)
				return err == nil
			}, timeout, duration).Should(BeTrue())

			var newClusterIP string = "None"
			codegenNew := &CodeGen{}
			codegenAPIMaticCR.DeepCopyInto(codegenNew)
			codegenNew.Spec.ServiceSpec.ClusterIP = &newClusterIP

			err = codegenNew.ValidateUpdate(codegenAPIMaticCR)
			Expect(err).ShouldNot(BeNil())
			Expect(k8sClient.Update(ctx, codegenNew)).ShouldNot(Succeed())
		})
	})

	Context("When updating a CodeGen instance by changing loadBalancerClass field", func() {
		It("Should give a validation error response", func() {
			By("Creating a new CodeGen instance with specifications")
			ctx := context.Background()
			var loadBalancerClass string = "testLbClass"
			codegen := &CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNameSpace,
				},
				Spec: CodeGenSpec{
					Replicas: 3,
					ServiceSpec: &CodeGenServiceSpec{
						LoadBalancerClass: &loadBalancerClass,
						Type:              corev1.ServiceTypeLoadBalancer,
					},
					LicenseSpec: CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  LicenseSourceType,
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
				},
			}
			codegen.Default()
			err := codegen.ValidateCreate()
			Expect(err).Should(BeNil())
			Expect(k8sClient.Create(ctx, codegen)).Should(Succeed())

			codegenLookupKey := types.NamespacedName{Name: CodeGenName, Namespace: CodeGenNameSpace}
			codegenCR := &CodeGen{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, codegenLookupKey, codegenCR)
				return err == nil
			}, timeout, duration).Should(BeTrue())

			var loadBalancerClass2 string = "testLbClass2"
			codegenNew := &CodeGen{}
			codegenCR.DeepCopyInto(codegenNew)
			codegenNew.Spec.ServiceSpec.LoadBalancerClass = &loadBalancerClass2

			err = codegenNew.ValidateUpdate(codegenCR)
			Expect(err).ShouldNot(BeNil())
			Expect(k8sClient.Update(ctx, codegenNew)).ShouldNot(Succeed())
		})
	})

	Context("When updating a CodeGen instance by changing labels field", func() {
		It("Should give a validation error response", func() {
			By("Creating a new CodeGen instance with specifications")
			ctx := context.Background()
			labels := map[string]string{
				"a": "b",
				"c": "d",
			}
			codegen := &CodeGen{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps.apimatic.io/v1beta2",
					Kind:       "CodeGen",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CodeGenName,
					Namespace: CodeGenNameSpace,
				},
				Spec: CodeGenSpec{
					Labels: labels,
					LicenseSpec: CodeGenLicenseSpec{
						CodeGenLicenseSourceType:  LicenseSourceType,
						CodeGenLicenseSourceValue: LicenseSourceValue,
					},
				},
			}
			codegen.Default()
			err := codegen.ValidateCreate()
			Expect(err).Should(BeNil())
			Expect(k8sClient.Create(ctx, codegen)).Should(Succeed())

			codegenLookupKey := types.NamespacedName{Name: CodeGenName, Namespace: CodeGenNameSpace}
			codegenCR := &CodeGen{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, codegenLookupKey, codegenCR)
				return err == nil
			}, timeout, duration).Should(BeTrue())

			codegenNew := &CodeGen{}
			codegenCR.DeepCopyInto(codegenNew)
			codegenNew.Spec.Labels["a"] = "k"

			err = codegenNew.ValidateUpdate(codegenCR)
			Expect(err).ShouldNot(BeNil())
			Expect(k8sClient.Update(ctx, codegenNew)).ShouldNot(Succeed())
		})
	})

})
