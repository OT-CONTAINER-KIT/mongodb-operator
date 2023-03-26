package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	opstreelabsinv1alpha1 "mongodb-operator/api/v1alpha1"
	k8sgo "mongodb-operator/k8sgo"
	mongoc "mongodb-operator/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MongoDBUserReconciler reconciles a MongoDBUser object
type MongoDBUserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=opstreelabs.in,resources=mongodbusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=opstreelabs.in,resources=mongodbusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=opstreelabs.in,resources=mongodbusers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MongoDBUser object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *MongoDBUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// _ = r.Log.WithValues("mongodbuser", req.NamespacedName)

	mongodbUser := &opstreelabsinv1alpha1.MongoDBUser{}
	mongodb := &opstreelabsinv1alpha1.MongoDB{}
	mongodbCluster := &opstreelabsinv1alpha1.MongoDBCluster{}

	err := r.Client.Get(context.TODO(), req.NamespacedName, mongodbUser)

	var mongoClient *mongo.Client

	// The MongoDBUser resource may have been deleted, so just return
	// and let the controller reconcile again later
	if err != nil {

		if errors.IsNotFound(err) {
			return ctrl.Result{RequeueAfter: time.Second * 10}, nil
		}
		return ctrl.Result{RequeueAfter: time.Second * 10}, err
	}
	var params mongoc.MongoDBParameters

	if mongodbUser.Spec.Type == "standalone" {
		err = r.Client.Get(context.TODO(), req.NamespacedName, mongodb)

		if err != nil {
 
			if errors.IsNotFound(err) {
				return ctrl.Result{RequeueAfter: time.Second * 10}, nil
			}
			return ctrl.Result{RequeueAfter: time.Second * 10}, err
		}
		passwordParams := k8sgo.SecretsParameters{Name: mongodb.ObjectMeta.Name, Namespace: mongodb.Namespace, SecretName: *mongodb.Spec.MongoDBSecurity.SecretRef.Name, SecretKey: *mongodb.Spec.MongoDBSecurity.SecretRef.Key}
		password := k8sgo.GetMongoDBPassword(passwordParams)
		// mongoURL := "mongodb://" + mongodb.Spec.MongoDBSecurity.MongoDBAdminUser + ":", password, "@"
		serviceName := fmt.Sprintf("%s-%s.%s", mongodb.ObjectMeta.Name, "standalone", mongodb.Namespace)
		mongoURL := fmt.Sprintf("mongodb://%s:%s@%s:27017/", mongodb.Spec.MongoDBSecurity.MongoDBAdminUser, password, serviceName)

		params = mongoc.MongoDBParameters{

			MongoURL:  mongoURL,
			SetupType: "standalone",
			Namespace: mongodbUser.Namespace,
			Name:      mongodbUser.Name,
			Password:  mongodbUser.Spec.Password,
			UserName:  &mongodbUser.Spec.User,
		}
		mongoClient = mongoc.InitiateMongoClient(params)

	} else if mongodbUser.Spec.Type == "replicaSet" {
		err = r.Client.Get(context.TODO(), req.NamespacedName, mongodbCluster)

		if err != nil {

			if errors.IsNotFound(err) {
				return ctrl.Result{RequeueAfter: time.Second * 10}, nil
			}
			return ctrl.Result{RequeueAfter: time.Second * 10}, err
		}
		passwordParams := k8sgo.SecretsParameters{Name: mongodbCluster.ObjectMeta.Name, Namespace: mongodbCluster.Namespace, SecretName: *mongodbCluster.Spec.MongoDBSecurity.SecretRef.Name, SecretKey: *mongodbCluster.Spec.MongoDBSecurity.SecretRef.Key}
		password := k8sgo.GetMongoDBPassword(passwordParams)
		

		mongoParams := mongoc.MongoDBParameters{
		Namespace: mongodbCluster.Namespace,
		Name:      mongodbCluster.ObjectMeta.Name,
		UserName:  &mongodbUser.Spec.User,
		SetupType: "cluster",
	}

		mongoURL := []string{"mongodb://", mongodbCluster.Spec.MongoDBSecurity.MongoDBAdminUser, ":", password, "@"}
	for node := 0; node < int(*mongodbCluster.Spec.MongoDBClusterSize); node++ {
		if node != int(*mongodbCluster.Spec.MongoDBClusterSize) {
			mongoURL = append(mongoURL, fmt.Sprintf("%s,", mongoc.GetMongoNodeInfo(mongoParams, node)))
		} else {
			mongoURL = append(mongoURL, mongoc.GetMongoNodeInfo(mongoParams, node))
		}
	}
	mongoURL = append(mongoURL, fmt.Sprintf("/?replicaSet=%s", mongodbCluster.ObjectMeta.Name))
	mongoString := strings.Join(mongoURL, "")

    
		params = mongoc.MongoDBParameters{

			MongoURL:  mongoString,
			SetupType: "cluster",
			Namespace: mongodbUser.Namespace,
			Name:      mongodbUser.Name,
			Password:  mongodbUser.Spec.Password,
			UserName:  &mongodbUser.Spec.User,
		}
		mongoClient = mongoc.InitiateMongoClient(params)
	}

	if err != nil {
		return ctrl.Result{}, err
	}
	defer mongoClient.Disconnect(ctx)

	// Create, delete, or update the MongoDB user based on the desired state
	if mongodbUser.DeletionTimestamp != nil {
		err = r.deleteMongoDBUser(ctx, mongoClient, mongodbUser)
		if err != nil {
			return ctrl.Result{}, err
		}
		// The object is being deleted, so don't requeue
		return ctrl.Result{}, nil
	} else if mongodbUser.Status.Created {
		err = r.updateMongoDBUser(ctx, mongoClient, mongodbUser)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else {
		err = r.createMongoDBUser(ctx, mongoClient, mongodbUser)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MongoDBUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opstreelabsinv1alpha1.MongoDBUser{}).
		Complete(r)
}

func (r *MongoDBUserReconciler) createMongoDBUser(ctx context.Context, client *mongo.Client, mongodbUser *opstreelabsinv1alpha1.MongoDBUser) error {
	db := client.Database(mongodbUser.Spec.Database)
	if db == nil {
		return fmt.Errorf("failed to get database %s", mongodbUser.Spec.Database)
	}

	// Create the MongoDB user
	result := db.RunCommand(ctx, bson.D{
		{Key: "createUser", Value: &mongodbUser.Spec.User},
		{Key: "pwd", Value: mongodbUser.Spec.Password},
		{Key: "roles", Value: bson.A{
			bson.D{{Key: "role", Value: mongodbUser.Spec.Role}, {Key: "db", Value: mongodbUser.Spec.Database}},
		}},
	})
	if err := result.Err(); err != nil {
		return err
	}

	// Update the MongoDBUser status
	mongodbUser.Status.Created = true
	if err := r.Status().Update(ctx, mongodbUser); err != nil {
		return err
	}

	return nil
}

func (r *MongoDBUserReconciler) updateMongoDBUser(ctx context.Context, client *mongo.Client, mongodbUser *opstreelabsinv1alpha1.MongoDBUser) error {
	db := client.Database(mongodbUser.Spec.Database)
	if db == nil {
		return fmt.Errorf("failed to get database %s", mongodbUser.Spec.Database)

	}

	// Update the MongoDB user
	result := db.RunCommand(ctx, bson.D{
		{Key: "updateUser", Value: mongodbUser.Spec.User},
		{Key: "pwd", Value: mongodbUser.Spec.Password},
		{Key: "roles", Value: bson.A{
			bson.D{{Key: "role", Value: mongodbUser.Spec.Role}, {Key: "db", Value: mongodbUser.Spec.Database}},
		}},
	})
	if result.Err() != nil {
		return result.Err()
	}
	return nil
}

func (r *MongoDBUserReconciler) deleteMongoDBUser(ctx context.Context, client *mongo.Client, mongodbUser *opstreelabsinv1alpha1.MongoDBUser) error {
	db := client.Database(mongodbUser.Spec.Database)
	if db == nil {
		return fmt.Errorf("failed to get database %s", mongodbUser.Spec.Database)

	}

	// Delete the MongoDB user
	result := db.RunCommand(ctx, bson.D{
		{Key: "dropUser", Value: mongodbUser.Spec.User},
	})
	if result.Err() != nil {
		return result.Err()
	}

	// Update the MongoDBUser status
	mongodbUser.Status.Created = false
	err := r.Status().Update(ctx, mongodbUser)
	if err != nil {
		return err
	}

	return nil
}
