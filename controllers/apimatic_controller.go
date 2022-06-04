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
	"context"
	"reflect"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apicodegenv1beta2 "github.com/apimatic/apimatic-codegen-operator/api/v1beta2"
)

// CodeGenReconciler reconciles a CodeGen object
type CodeGenReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.apimatic.io,resources=codegens,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.apimatic.io,resources=codegens/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.apimatic.io,resources=codegens/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;patch;delete;update
//+kubebuilder:rbac:groups=core,resources=services/status,verbs=get
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CodeGen object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *CodeGenReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// TODO(user): your logic here
	// Fetch the APIMatic instance
	codegen := &apicodegenv1beta2.CodeGen{}
	err := r.Get(ctx, req.NamespacedName, codegen)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("CodeGen resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request
		log.Error(err, "Failed to get CodeGen")
		return ctrl.Result{}, err
	}

	// generateResources returns list of resources to check.
	for _, wantedResource := range generateResources(codegen) {
		kind := reflect.ValueOf(wantedResource).Elem().Type().String()
		if err := ctrl.SetControllerReference(codegen, wantedResource, r.Scheme); err != nil {
			log.Error(err, "Failed to set controller reference", "Name", codegen.Name, "NameSpace", codegen.Namespace, "Kind", kind)
			return ctrl.Result{}, err
		}
		// Fetch this resource from cluster, if any.
		clusterResource := wantedResource.DeepCopyObject().(GenericResource)
		err := r.Get(ctx, req.NamespacedName, clusterResource)

		if err != nil && errors.IsNotFound(err) {
			// Create wanted resource. Next client.Get will at least fetch wantedResource and will eventually
			// reflect the object as it is in the cluster.
			if err := r.Create(ctx, wantedResource); err != nil {
				log.Error(err, "Error creating a new resource", "Name", codegen.Name, "NameSpace", codegen.Namespace, "Kind", kind)
				return ctrl.Result{}, err
			}
			log.Info("Creating resource", "Name", codegen.Name, "NameSpace", codegen.Namespace, "Kind", kind)
			return ctrl.Result{Requeue: true}, nil
		}
		// Patch cluster resource to wanted resource, if needed.
		if !SubsetEqual(wantedResource, clusterResource) {
			log.Info("Applying patch", "Name", codegen.Name, "NameSpace", codegen.Namespace, "Kind", kind)
			// patch.Merge uses the raw object as a merge patch, without modifications.
			if err := r.Patch(ctx, wantedResource, client.Merge); err != nil {
				log.Error(err, "Error patching resource", "Name", codegen.Name, "NameSpace", codegen.Namespace, "Kind", kind)
				return ctrl.Result{}, err
			}
			log.Info("Patching resource", "Name", codegen.Name, "NameSpace", codegen.Namespace, "Kind", kind)
			return ctrl.Result{Requeue: true}, nil
		}
		if resource, update, err := updateResource(codegen, clusterResource); update {
			if err != nil {
				log.Error(err, "Error updating resource specific fields not managed by patch", "Name", codegen.Name, "NameSpace", codegen.Namespace, "Kind", kind)
				return ctrl.Result{}, err
			}
			if err := r.Update(ctx, resource); err != nil {
				log.Error(err, "Error updating resource specific fields not managed by patch", "Name", codegen.Name, "NameSpace", codegen.Namespace, "Kind", kind)
				return ctrl.Result{}, err
			}
			log.Info("Updating resource fields not properly managed by patch", "Name", codegen.Name, "NameSpace", codegen.Namespace, "Kind", kind)
			return ctrl.Result{Requeue: true}, nil
		}
	}

	codegenForStatus := &apicodegenv1beta2.CodeGen{}
	if err := r.Get(ctx, req.NamespacedName, codegenForStatus); err != nil {
		log.Error(err, "Error getting codegenForStatus", "Name", req.Name, "NameSpace", req.NamespacedName)
		return ctrl.Result{}, err
	}
	status := apicodegenv1beta2.CodeGenStatus{
		Replicas: codegenForStatus.Spec.Replicas,
	}
	if !reflect.DeepEqual(status, codegen.Status) {
		codegenForStatus.Status = status
		log.Info("Updating CodeGen Status", "Name", codegenForStatus.Name, "NameSpace", codegenForStatus.Namespace)
		if err := r.Status().Update(ctx, codegenForStatus); err != nil {
			log.Error(err, "Failed to update CodeGen Status", "Name", codegenForStatus.Name, "NameSpace", codegenForStatus.Namespace)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CodeGenReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return r.setupWithManager(mgr)
}
