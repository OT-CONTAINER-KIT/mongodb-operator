package k8sgo

import (
	"context"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	policyv1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// PodDisruptionParameters is an input parameter structure for Pod disruption budget
type PodDisruptionParameters struct {
	PDBMeta        metav1.ObjectMeta
	OwnerDef       metav1.OwnerReference
	Labels         map[string]string
	Namespace      string
	MinAvailable   *int32
	MaxUnavailable *int32
}

// CreateOrUpdatePodDisruption method will create or update MongoDB PodDisruptionBudgets
func CreateOrUpdatePodDisruption(params PodDisruptionParameters) error {
	logger := logGenerator(params.PDBMeta.Name, params.Namespace, "PodDisruptionBudget")
	pdbDef := generatePodDisruption(params)
	storedPDB, err := getPodDisruption(params.Namespace, params.PDBMeta.Name)
	if err != nil {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(pdbDef); err != nil {
			logger.Error(err, "Unable to patch MongoDB PodDisruptionBudget with comparison object")
			return err
		}
		if errors.IsNotFound(err) {
			return createPodDisruption(params.Namespace, pdbDef)
		}
		return err
	}
	return patchPodDisruption(storedPDB, pdbDef, params.Namespace)
}

// patchPodDisruption will patch MongoDB Kubernetes PodDisruptionBudgets
func patchPodDisruption(storedPdb *policyv1.PodDisruptionBudget, newPdb *policyv1.PodDisruptionBudget, namespace string) error {
	logger := logGenerator(newPdb.Name, namespace, "PodDisruptionBudget")
	newPdb.ResourceVersion = storedPdb.ResourceVersion
	newPdb.CreationTimestamp = storedPdb.CreationTimestamp
	newPdb.ManagedFields = storedPdb.ManagedFields

	storedPdb.Kind = "PodDisruptionBudget"
	storedPdb.APIVersion = "policy/v1"

	patchResult, err := patch.DefaultPatchMaker.Calculate(storedPdb, newPdb,
		patch.IgnorePDBSelector(),
		patch.IgnoreStatusFields(),
	)
	if err != nil {
		logger.Error(err, "Unable to patch MongoDB PodDisruption with comparison object")
		return err
	}
	if !patchResult.IsEmpty() {
		logger.Info("Changes in PodDisruptionBudget Detected, Updating...",
			"patch", string(patchResult.Patch),
			"Current", string(patchResult.Current),
			"Original", string(patchResult.Original),
			"Modified", string(patchResult.Modified),
		)
		for key, value := range storedPdb.Annotations {
			if _, present := newPdb.Annotations[key]; !present {
				newPdb.Annotations[key] = value
			}
		}
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(newPdb); err != nil {
			logger.Error(err, "Unable to patch MongoDB PodDisruptionBudget with comparison object")
			return err
		}
		return updatePodDisruption(namespace, newPdb)
	}
	logger.Info("PodDisruptionBudget is reconciled, nothing to change")
	return nil
}

// updatePodDisruption is a method to create Pod disruption budget
func updatePodDisruption(namespace string, pdb *policyv1.PodDisruptionBudget) error {
	logger := logGenerator(pdb.Name, namespace, "PodDisruptionBudget")
	_, err := generateK8sClient().PolicyV1beta1().PodDisruptionBudgets(namespace).Update(context.TODO(), pdb, metav1.UpdateOptions{})
	if err != nil {
		logger.Error(err, "MongoDB PodDisruptionBudget update failed")
		return err
	}
	logger.Info("MongoDB PodDisruptionBudget update was successful", "PDB.Spec", pdb.Spec)
	return nil
}

// createPodDisruption is a method to create Pod disruption budget
func createPodDisruption(namespace string, pdb *policyv1.PodDisruptionBudget) error {
	logger := logGenerator(pdb.Name, namespace, "PodDisruptionBudget")
	_, err := generateK8sClient().PolicyV1beta1().PodDisruptionBudgets(namespace).Create(context.TODO(), pdb, metav1.CreateOptions{})
	if err != nil {
		logger.Error(err, "MongoDB PodDisruptionBudget creation failed")
		return err
	}
	logger.Info("MongoDB PodDisruptionBudget creation was successful", "PDB.Spec", pdb.Spec)
	return nil
}

// getPodDisruption is a method to get Pod disruption budget
func getPodDisruption(namespace, name string) (*policyv1.PodDisruptionBudget, error) {
	logger := logGenerator(name, namespace, "PodDisruptionBudget")
	pdbInfo, err := generateK8sClient().PolicyV1beta1().PodDisruptionBudgets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logger.Info("Unable to get pod disruption budget")
		return nil, err
	}
	logger.Info("MongoDB PodDisruptionBudget get action was successful")
	return pdbInfo, err
}

// generatePodDisruption is a method to generate Pod disruption budget definiton
func generatePodDisruption(params PodDisruptionParameters) *policyv1.PodDisruptionBudget {
	pdbTemplate := &policyv1.PodDisruptionBudget{
		TypeMeta:   generateMetaInformation("PodDisruptionBudget", "policy/v1beta1"),
		ObjectMeta: params.PDBMeta,
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: params.Labels,
			},
			MaxUnavailable: &intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: int32(*params.MaxUnavailable),
			},
		},
	}
	if params.MinAvailable != nil {
	    pdbTemplate.Spec.MinAvailable = &intstr.IntOrString{
	        Type:   intstr.Int,
	        IntVal: int32(*params.MinAvailable),
	    }
	}
	if params.MaxUnavailable != nil {
	    pdbTemplate.Spec.MaxUnavailable = &intstr.IntOrString{
	        Type:   intstr.Int,
	        IntVal: int32(*params.MaxUnavailable),
	    }
	}
	AddOwnerRefToObject(pdbTemplate, params.OwnerDef)
	return pdbTemplate
}
