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

package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"mongodb-operator/k8sgo"

	opstreelabsinv1alpha1 "mongodb-operator/api/v1alpha1"
)

// MongoDBReconciler reconciles a MongoDB object
type MongoDBReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=opstreelabs.in,resources=mongodbs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=opstreelabs.in,resources=mongodbs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=opstreelabs.in,resources=mongodbs/finalizers,verbs=update
//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps;events;services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
func (r *MongoDBReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// reqLogger := r.Log.WithValues("Request.Namespace", req.NamespacedName)
	instance := &opstreelabsinv1alpha1.MongoDB{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if err := controllerutil.SetControllerReference(instance, instance, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	// reqLogger.Info("Reconciling Opstree MongoDB controller")
	err = k8sgo.CreateMongoStandaloneSetup(instance)
	if err != nil {
		return ctrl.Result{}, err
	}
	err = k8sgo.CreateMongoStandaloneService(instance)
	if err != nil {
		return ctrl.Result{}, err
	}
	// reqLogger.Info("Will reconcile mongodb operator in again 10 seconds")
	return ctrl.Result{RequeueAfter: time.Second * 10}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MongoDBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opstreelabsinv1alpha1.MongoDB{}).
		Complete(r)
}
