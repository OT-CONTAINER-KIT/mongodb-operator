package k8sgo

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/iamabhishek-dubey/k8s-objectmatcher/patch"
	apiErrors "github.com/pkg/errors"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	opstreelabsinv1alpha1 "mongodb-operator/api/v1alpha1"
	types "mongodb-operator/k8sgo/type"
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
	AdditionalConfig  *string
	SecurityContext   *corev1.PodSecurityContext
	TLS               bool
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
func CreateOrUpdateStateFul(params statefulSetParameters, cluster *opstreelabsinv1alpha1.MongoDBCluster, standalone *opstreelabsinv1alpha1.MongoDB) error {
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

	oldStorage := storedStateful.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests.Storage()
	newStorage := statefulSetDef.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests.Storage()
	if canExpandPVC(*oldStorage, newStorage) {
		if cluster.Status.State != types.Expanding {
			return apiErrors.Errorf("expanding")
		}
		zap.S().Info("canExpandPVC true")
		// delete sts
		policy := metav1.DeletePropagationOrphan
		if err := generateK8sClient().AppsV1().StatefulSets(params.Namespace).Delete(context.TODO(), storedStateful.Name, metav1.DeleteOptions{PropagationPolicy: &policy}); err != nil {
			return err
		}

		// expand pvc
		if err := dealWithExpandingPVC(context.TODO(), *storedStateful); err != nil {
			return err
		}

		// build sts , recreate sts
		if err := createStateFulSet(params.Namespace, statefulSetDef); err != nil {
			return err
		}

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
		patch.IgnorePersistenVolumeFields(),
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

func checkExpandPVC() {

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
					SecurityContext:   params.SecurityContext,
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
	if params.AdditionalConfig != nil {
		statefulset.Spec.Template.Spec.Volumes = getAdditionalConfig(params)
	}
	if params.ImagePullSecret != nil {
		statefulset.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{Name: *params.ImagePullSecret}}
	}
	if params.TLS {
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, getVolumeFromSecret(tlsCAVolumeName, params.StatefulSetMeta.Name+"-ca-certificate")...)
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, getVolumeFromSecret(tlsCertVolumeName, params.StatefulSetMeta.Name+"-server-certificate-key")...)
	}

	AddOwnerRefToObject(statefulset, params.OwnerDef)
	return statefulset
}

// generatePersistentVolumeTemplate is a method to create the persistent volume claim template
func generatePersistentVolumeTemplate(params pvcParameters) corev1.PersistentVolumeClaim {
	return corev1.PersistentVolumeClaim{
		TypeMeta:   generateMetaInformation("PersistentVolumeClaim", "v1"),
		ObjectMeta: metav1.ObjectMeta{Name: params.Name},
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

// getAdditionalConfig will return the MongoDB additional configuration
func getAdditionalConfig(params statefulSetParameters) []corev1.Volume {
	return []corev1.Volume{
		{
			Name: "external-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: *params.AdditionalConfig,
					},
				},
			},
		},
	}
}

// logGenerator is a method to generate logging interfacce
func logGenerator(name, namespace, resourceType string) logr.Logger {
	reqLogger := Log.WithValues("Namespace", namespace, "Name", name, "Resource Type", resourceType)
	return reqLogger
}
