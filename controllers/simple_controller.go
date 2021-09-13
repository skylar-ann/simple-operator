package controllers

/*

Copyright 2021.

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

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	simplev1alpha1 "github.worldpay.com/Atlas/simple-operator/api/v1alpha1"
)

// SimpleReconciler reconciles a Simple object
type SimpleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=simple.atlas.fis.dev,resources=simples,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=simple.atlas.fis.dev,resources=simples/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=simple.atlas.fis.dev,resources=simples/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Simple object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *SimpleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// your logic here
	logger := ctrl.LoggerFrom(ctx)

	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: req.Name, Namespace: req.Namespace}}
	cm.Data = make(map[string]string)

	simple := &simplev1alpha1.Simple{}
	err := r.Get(ctx, req.NamespacedName, simple)
	if err != nil {
		return ctrl.Result{}, nil
	}

	logger.Info("Simple object", "simple", simple)

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, cm, func() error {
		cm.Data = make(map[string]string)
		if simple.Spec.Foo != "" {
			cm.Data["something.conf"] = fmt.Sprintf("setting = %s", simple.Spec.Foo)
		}

		if err := ctrl.SetControllerReference(simple, cm, r.Scheme); err != nil {
			logger.Error(err, "Failed to set controller ref")
			return err
		}
		return nil
	})

	if err != nil {
		logger.Info("failed to reconcile ConfigMap")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SimpleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&simplev1alpha1.Simple{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
