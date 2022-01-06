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
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"

	opstreelabsinv1alpha1 "mongodb-operator/api/v1alpha1"
	"mongodb-operator/k8sgo"
)

// MongoDBClusterReconciler reconciles a MongoDBCluster object
type MongoDBClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=opstreelabs.in,resources=mongodbclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=opstreelabs.in,resources=mongodbclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=opstreelabs.in,resources=mongodbclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
func (r *MongoDBClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	instance := &opstreelabsinv1alpha1.MongoDBCluster{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{RequeueAfter: time.Second * 10}, nil
		}
		return ctrl.Result{RequeueAfter: time.Second * 10}, err
	}
	if err := controllerutil.SetControllerReference(instance, instance, r.Scheme); err != nil {
		return ctrl.Result{RequeueAfter: time.Second * 10}, err
	}
	if !k8sgo.CheckSecretExist(instance.Namespace, fmt.Sprintf("%s-%s", instance.ObjectMeta.Name, "cluster-monitoring")) {
		err = k8sgo.CreateMongoClusterMonitoringSecret(instance)
		if err != nil {
			return ctrl.Result{RequeueAfter: time.Second * 10}, err
		}
	}
	err = k8sgo.CreateMongoClusterSetup(instance)
	if err != nil {
		return ctrl.Result{RequeueAfter: time.Second * 10}, err
	}
	err = k8sgo.CreateMongoClusterMonitoringService(instance)
	if err != nil {
		return ctrl.Result{RequeueAfter: time.Second * 10}, err
	}
	err = k8sgo.CreateMongoClusterService(instance)
	if err != nil {
		return ctrl.Result{RequeueAfter: time.Second * 10}, err
	}
	mongoDBSTS, err := k8sgo.GetStateFulSet(instance.Namespace, fmt.Sprintf("%s-%s", instance.ObjectMeta.Name, "cluster"))
	if err != nil {
		return ctrl.Result{RequeueAfter: time.Second * 10}, err
	}
	if int(mongoDBSTS.Status.ReadyReplicas) != int(*instance.Spec.MongoDBClusterSize) {
		return ctrl.Result{RequeueAfter: time.Second * 60}, nil
	} else {
		state, err := k8sgo.CheckMongoClusterStateInitialized(instance)
		if err != nil || !state {
			err = k8sgo.InitializeMongoDBCluster(instance)
			if err != nil {
				return ctrl.Result{RequeueAfter: time.Second * 10}, err
			}
			if !k8sgo.CheckMongoDBClusterMonitoringUser(instance) {
				err = k8sgo.CreateMongoDBClusterMonitoringUser(instance)
				if err != nil {
					return ctrl.Result{RequeueAfter: time.Second * 10}, err
				}
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MongoDBClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opstreelabsinv1alpha1.MongoDBCluster{}).
		Complete(r)
}
