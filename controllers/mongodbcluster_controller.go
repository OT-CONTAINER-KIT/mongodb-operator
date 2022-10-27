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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"mongodb-operator/controllers/watch"
	"mongodb-operator/k8sgo/status"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"

	opstreelabsinv1alpha1 "mongodb-operator/api/v1alpha1"
	"mongodb-operator/k8sgo"
	types "mongodb-operator/k8sgo/type"
)

// MongoDBClusterReconciler reconciles a MongoDBCluster object
type MongoDBClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	//log    *zap.SugaredLogger
	secretWatcher    *watch.ResourceWatcher
	configMapWatcher *watch.ResourceWatcher
}

func NewReconciler(mgr manager.Manager) *MongoDBClusterReconciler {
	secretWatcher := watch.New()
	configMapWatcher := watch.New()

	return &MongoDBClusterReconciler{
		Client:           mgr.GetClient(),
		Scheme:           mgr.GetScheme(),
		secretWatcher:    &secretWatcher,
		configMapWatcher: &configMapWatcher,
	}
}

//+kubebuilder:rbac:groups=opstreelabs.in,resources=mongodbclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=opstreelabs.in,resources=mongodbclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=opstreelabs.in,resources=mongodbclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups="policy",resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete

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

	isValid, err := k8sgo.ValidateTLSConfig(instance)

	if err != nil {
		return status.Update(r.Client.Status(), instance,
			statusOptions().
				withMessage(Error, fmt.Sprintf("Error validating TLS config: %s", err)).
				withFailedState(),
		)
	}

	if !isValid {
		return status.Update(r.Client.Status(), instance,
			statusOptions().
				withMessage(Info, "TLS config is not yet valid, retrying in 10 seconds").
				withPendingState(10),
		)
	}
	if err := k8sgo.EnsureTLSResources(instance); err != nil {
		return status.Update(r.Client.Status(), instance,
			statusOptions().
				withMessage(Error, fmt.Sprintf("Error ensuring TLS resources: %s", err)).
				withFailedState(),
		)
	}

	err = k8sgo.CreateMongoClusterSetup(instance)
	if err != nil {
		if err.Error() == "expanding" {
			return status.Update(r.Client.Status(), instance, statusOptions().
				withMessage(Info, "Expanding pvc").
				withExpandingState(5),
			)
		}
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

	if instance.Status.State == "" {
		return status.Update(r.Client.Status(), instance, statusOptions().
			withMessage(Info, "Creating cluster").
			withCreatingState(10),
		)
	}

	if int(mongoDBSTS.Status.ReadyReplicas) != int(*instance.Spec.MongoDBClusterSize) {
		if instance.Status.State != types.Pending && instance.Status.State != types.Creating {
			return status.Update(r.Client.Status(), instance, statusOptions().
				withMessage(Info, "Updating cluster").
				withPendingState(10),
			)
		}
		return ctrl.Result{RequeueAfter: time.Second * 10}, nil
	}

	if err := k8sgo.CheckMongoClusterState(instance); err != nil {
		return status.Update(r.Client.Status(), instance, statusOptions().
			withMessage(Warn, fmt.Sprintf("Error with connecting mongodb: %s", err)).
			withPendingState(5),
		)
	}

	k8sgo.Log.Info("MongoDB Cluster is healthy")

	if err := k8sgo.CheckMongoDBClusterMonitoringUser(instance); err != nil {
		return status.Update(r.Client.Status(), instance, statusOptions().
			withMessage(Warn, fmt.Sprintf("Error with checking monitor user in mongodb: %s", err)).
			withPendingState(5),
		)
	}

	return status.Update(r.Client.Status(), instance, statusOptions().
		withMessage(Info, "done").
		withRunningState(),
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *MongoDBClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{MaxConcurrentReconciles: 3}).
		For(&opstreelabsinv1alpha1.MongoDBCluster{}).
		Watches(&source.Kind{Type: &corev1.Secret{}}, r.secretWatcher).
		Watches(&source.Kind{Type: &corev1.ConfigMap{}}, r.configMapWatcher).
		Complete(r)
}
