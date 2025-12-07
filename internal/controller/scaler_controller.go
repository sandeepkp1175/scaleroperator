/*
Copyright 2025 Sandeep Pattnaik.

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

package controller

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	apiv1alpha1 "github.com/sandeepkp1175/scaler-operator/api/v1alpha1"
)

// ScalerReconciler reconciles a Scaler object
type ScalerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=api.sandeeppattnaik.online,resources=scalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=api.sandeeppattnaik.online,resources=scalers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=api.sandeeppattnaik.online,resources=scalers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Scaler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *ScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// TODO(user): your logic here

	logf.Log.Info("Reconcile function called for Scaler", "NamespacedName", req.NamespacedName)

	log := ctrl.LoggerFrom(ctx).WithValues("scaler", req.NamespacedName)
	_ = log
	log.Info("Scaler resource reconciled successfully")
	scaler := &apiv1alpha1.Scaler{}
	error := r.Get(ctx, req.NamespacedName, scaler)

	if error != nil {
		log.Error(error, "Failed to get Scaler resource")
		return ctrl.Result{}, client.IgnoreNotFound(error)
	} else {
		log.Info("Successfully fetched Scaler resource", "Scaler.Spec", scaler.Spec)
	}

	startTime := scaler.Spec.StartTime
	endTime := scaler.Spec.EndTime
	desiredReplicas := scaler.Spec.Replicas

	currentHour := int32(time.Now().Hour())

	log.Info("Current Hour", "Hour", currentHour)

	if currentHour >= startTime && currentHour < endTime {

		for _, deploy := range scaler.Spec.Deployments {
			deployment := &appsv1.Deployment{}

			error := r.Get(ctx, types.NamespacedName{Namespace: deploy.Namespace, Name: deploy.Name}, deployment)

			if error != nil {
				log.Error(error, "Failed to get Deployment", "Deployment.Namespace", deploy.Namespace, "Deployment.Name", deploy.Name)
				return ctrl.Result{}, error
			}

			if *deployment.Spec.Replicas != int32(desiredReplicas) {
				*deployment.Spec.Replicas = int32(desiredReplicas)
				r.Update(ctx, deployment)
			}

		} // end for loop

		log.Info("Within scaling window, scaling to desired replicas", "DesiredReplicas", desiredReplicas)
	} else {
		log.Info("Outside scaling window, no scaling action taken")
	}

	log.Info("Scaler Spec Details", "StartTime", startTime, "EndTime", endTime, "DesiredReplicas", desiredReplicas)

	return ctrl.Result{RequeueAfter: time.Duration(30 * time.Second)}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.Scaler{}).
		Named("scaler").
		Complete(r)
}
