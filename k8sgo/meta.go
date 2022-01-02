package k8sgo

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mongodbv1alpha1 "mongodb-operator/api/v1alpha1"
)

// generateMetaInformation generates the meta information
func generateMetaInformation(resourceKind string, apiVersion string) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       resourceKind,
		APIVersion: apiVersion,
	}
}

// generateObjectMetaInformation generates the object meta information
func generateObjectMetaInformation(name string, namespace string, labels map[string]string, annotations map[string]string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:        name,
		Namespace:   namespace,
		Labels:      labels,
		Annotations: annotations,
	}
}

// AddOwnerRefToObject adds the owner references to object
func AddOwnerRefToObject(obj metav1.Object, ownerRef metav1.OwnerReference) {
	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), ownerRef))
}

// mongoAsOwner generates and returns object refernece
func mongoAsOwner(cr *mongodbv1alpha1.MongoDB) metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: cr.APIVersion,
		Kind:       cr.Kind,
		Name:       cr.Name,
		UID:        cr.UID,
		Controller: &trueVar,
	}
}

// mongoClusterAsOwner generates and returns object refernece
func mongoClusterAsOwner(cr *mongodbv1alpha1.MongoDBCluster) metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: cr.APIVersion,
		Kind:       cr.Kind,
		Name:       cr.Name,
		UID:        cr.UID,
		Controller: &trueVar,
	}
}

// LabelSelectors generates object for label selection
func LabelSelectors(labels map[string]string) *metav1.LabelSelector {
	return &metav1.LabelSelector{MatchLabels: labels}
}

// generateAnnotations generates and returns annotations
func generateAnnotations() map[string]string {
	return map[string]string{
		"mongodb.opstreelabs.in": "true",
		"prometheus.io/scrape":   "true",
		"prometheus.io/port":     "9216",
	}
}
