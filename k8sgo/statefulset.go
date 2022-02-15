package k8sgo

import (
	"context"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	appsv1 "k8s.io/api/apps/v1"
)

// statefulSetParameters is the input struct for MongoDB statefulset
type statefulSetParameters struct {
	StatefulSetMeta   metav1.ObjectMeta
	OwnerDef          metav1.OwnerReference
	Namespace         string
	ContainerParams   containerParameters
	Labels            map[string]string
	Annotations       map[string]string
	Replicas          *int32
	PVCParameters     pvcParameters
	ExtraVolumes      *[]corev1.Volume
	ImagePullSecret   *string
	Affinity          *corev1.Affinity
	NodeSelector      map[string]string
	Tolerations       *[]corev1.Toleration
	PriorityClassName string
}

// pvcParameters is the structure for MongoDB PVC
type pvcParameters struct {
	Name             string
	Namespace        string
	Labels           map[string]string
	Annotations      map[string]string
	AccessModes      []corev1.PersistentVolumeAccessMode
	StorageClassName *string
	StorageSize      string
}

// CreateOrUpdateStateFul method will create or update StatefulSet
func CreateOrUpdateStateFul(params statefulSetParameters) error {
	logger := logGenerator(params.StatefulSetMeta.Name, params.Namespace, "StatefulSet")
	storedStateful, err := GetStateFulSet(params.Namespace, params.StatefulSetMeta.Name)
	statefulSetDef := generateStatefulSetDef(params)
	if err != nil {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(statefulSetDef); err != nil {
			logger.Error(err, "Unable to patch redis statefulset with comparison object")
			return err
		}
		if errors.IsNotFound(err) {
			return createStateFulSet(params.Namespace, statefulSetDef)
		}
		return err
	}
	return patchStateFulSet(storedStateful, statefulSetDef, params.Namespace)
}

// patchStateFulSet will patch Statefulset
func patchStateFulSet(storedStateful *appsv1.StatefulSet, newStateful *appsv1.StatefulSet, namespace string) error {
	logger := logGenerator(storedStateful.Name, namespace, "StatefulSet")
	// adding meta information
	newStateful.ResourceVersion = storedStateful.ResourceVersion
	newStateful.CreationTimestamp = storedStateful.CreationTimestamp
	newStateful.ManagedFields = storedStateful.ManagedFields
	patchResult, err := patch.DefaultPatchMaker.Calculate(storedStateful, newStateful,
		patch.IgnoreStatusFields(),
		patch.IgnoreVolumeClaimTemplateTypeMetaAndStatus(),
		patch.IgnoreField("kind"),
		patch.IgnoreField("apiVersion"),
		patch.IgnoreField("metadata"),
	)
	if err != nil {
		logger.Error(err, "Unable to patch mongodb statefulset with comparison object")
		return err
	}
	if !patchResult.IsEmpty() {
		logger.Info("Changes in statefulset Detected, Updating...", "patch", string(patchResult.Patch))
		for key, value := range storedStateful.Annotations {
			if _, present := newStateful.Annotations[key]; !present {
				newStateful.Annotations[key] = value
			}
		}
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(newStateful); err != nil {
			logger.Error(err, "Unable to patch mongodb statefulset with comparison object")
			return err
		}
		return updateStateFulSet(namespace, newStateful)
	}
	logger.Info("Reconciliation Complete, no Changes required.")
	return nil
}

// createStateFulSet is a method to create statefulset in Kubernetes
func createStateFulSet(namespace string, stateful *appsv1.StatefulSet) error {
	logger := logGenerator(stateful.Name, namespace, "StatefulSet")
	_, err := generateK8sClient().AppsV1().StatefulSets(namespace).Create(context.TODO(), stateful, metav1.CreateOptions{})
	if err != nil {
		logger.Error(err, "MongoDB Statefulset creation failed")
		return err
	}
	logger.Info("MongoDB Statefulset successfully created")
	return nil
}

// updateStateFulSet is a method to update statefulset in Kubernetes
func updateStateFulSet(namespace string, stateful *appsv1.StatefulSet) error {
	logger := logGenerator(stateful.Name, namespace, "StatefulSet")
	_, err := generateK8sClient().AppsV1().StatefulSets(namespace).Update(context.TODO(), stateful, metav1.UpdateOptions{})
	if err != nil {
		logger.Error(err, "MongoDB Statefulset update failed")
		return err
	}
	logger.Info("MongoDB Statefulset successfully updated")
	return nil
}

// GetStateFulSet is a method to get statefulset in Kubernetes
func GetStateFulSet(namespace string, stateful string) (*appsv1.StatefulSet, error) {
	logger := logGenerator(stateful, namespace, "StatefulSet")
	statefulInfo, err := generateK8sClient().AppsV1().StatefulSets(namespace).Get(context.TODO(), stateful, metav1.GetOptions{})
	if err != nil {
		logger.Info("MongoDB Statefulset get action failed")
		return nil, err
	}
	logger.Info("MongoDB Statefulset get action was successful")
	return statefulInfo, err
}

// generateStatefulSetDef is a method to generate statefulset definition
func generateStatefulSetDef(params statefulSetParameters) *appsv1.StatefulSet {
	statefulset := &appsv1.StatefulSet{
		TypeMeta:   generateMetaInformation("StatefulSet", "apps/v1"),
		ObjectMeta: params.StatefulSetMeta,
		Spec: appsv1.StatefulSetSpec{
			Selector:    LabelSelectors(params.Labels),
			ServiceName: params.StatefulSetMeta.Name,
			Replicas:    params.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: params.Labels},
				Spec: corev1.PodSpec{
					Containers:        generateContainerDef(params.StatefulSetMeta.Name, params.ContainerParams),
					NodeSelector:      params.NodeSelector,
					Affinity:          params.Affinity,
					PriorityClassName: params.PriorityClassName,
				},
			},
		},
	}

	if params.Tolerations != nil {
		statefulset.Spec.Template.Spec.Tolerations = *params.Tolerations
	}
	if params.ContainerParams.PersistenceEnabled != nil && *params.ContainerParams.PersistenceEnabled {
		statefulset.Spec.VolumeClaimTemplates = append(statefulset.Spec.VolumeClaimTemplates, generatePersistentVolumeTemplate(params.PVCParameters))
	}
	if params.ExtraVolumes != nil {
		statefulset.Spec.Template.Spec.Volumes = *params.ExtraVolumes
	}
	if params.ImagePullSecret != nil {
		statefulset.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{Name: *params.ImagePullSecret}}
	}
	AddOwnerRefToObject(statefulset, params.OwnerDef)
	return statefulset
}

// generatePersistentVolumeTemplate is a method to create the persistent volume claim template
func generatePersistentVolumeTemplate(params pvcParameters) corev1.PersistentVolumeClaim {
	return corev1.PersistentVolumeClaim{
		TypeMeta:   generateMetaInformation("PersistentVolumeClaim", "v1"),
		ObjectMeta: generateObjectMetaInformation(params.Name, params.Namespace, params.Labels, params.Annotations),
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: params.AccessModes,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceName(corev1.ResourceStorage): resource.MustParse(params.StorageSize),
				},
			},
			StorageClassName: params.StorageClassName,
		},
	}
}

// logGenerator is a method to generate logging interfacce
func logGenerator(name, namespace, resourceType string) logr.Logger {
	reqLogger := log.WithValues("Namespace", namespace, "Name", name, "Resource Type", resourceType)
	return reqLogger
}
