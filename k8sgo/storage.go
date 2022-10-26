package k8sgo

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

const waitLimit = 2 * 60 * 60

func canExpandPVC(oldDataStorage resource.Quantity, newDataStorage *resource.Quantity) bool {
	zap.S().Info("oldDataStorage: ", oldDataStorage, ",newDataStorage: ", newDataStorage)
	dataChanged := true
	if newDataStorage.Cmp(oldDataStorage) != 1 {
		zap.S().Info("Can not expand, new pvc is not larger than old pvc")
		dataChanged = false
	}

	return dataChanged
}

func doExpandPVC(ctx context.Context, pvc *v1.PersistentVolumeClaim, sts appsv1.StatefulSet) error {
	name := pvc.Name
	Log.Info("expand PVC【", name, "】 start")
	pvc.Spec.Resources.Requests = sts.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests

	if _, err := generateK8sClient().CoreV1().PersistentVolumeClaims(sts.Namespace).Update(
		ctx,
		pvc,
		metav1.UpdateOptions{},
	); err != nil {
		return err
	}
	if err := retry(time.Second*2, time.Duration(waitLimit)*time.Second, func() (bool, error) {
		// Check the pvc status.
		var currentPVC corev1.PersistentVolumeClaim

		if _, err2 := generateK8sClient().CoreV1().PersistentVolumeClaims(sts.Namespace).Get(ctx, name, metav1.GetOptions{}); err2 != nil {
			return true, err2
		}
		var conditons = currentPVC.Status.Conditions
		capacity := currentPVC.Status.Capacity
		// Notice: When expanding not start, or been completed, conditons is nil
		if conditons == nil {
			// If change storage request when replicas are creating, should check the currentPVC.Status.Capacity.
			// for example:
			// Pod0 has created successful,but Pod1 is creating. then change PVC from 20Gi to 30Gi .
			// Pod0's PVC need to expand, but Pod1's PVC has created as 30Gi, so need to skip it.

			if equality.Semantic.DeepEqual(capacity, pvc.Spec.Resources.Requests) {
				Log.Info("Executing expand PVC【", name, "】 completed")
				return true, nil
			}
			Log.Info("Executing expand PVC【", name, "】 not start")
			return false, nil
		}
		status := conditons[0].Type
		Log.Info("Executing expand PVC【", name, "】, storage 【", capacity.Storage(), "】, status 【", status, "】")
		if status == "FileSystemResizePending" {
			return true, nil
		}
		return false, nil
	}); err != nil {
		return err
	}

	return nil
}

// retry runs func "f" every "in" time until "limit" is reached.
// it also doesn't have an extra tail wait after the limit is reached
// and f func runs first time instantly
func retry(in, limit time.Duration, f func() (bool, error)) error {
	fdone, err := f()
	if err != nil {
		return err
	}
	if fdone {
		return nil
	}

	done := time.NewTimer(limit)
	defer done.Stop()
	tk := time.NewTicker(in)
	defer tk.Stop()

	for {
		select {
		case <-done.C:
			return fmt.Errorf("reach pod wait limit")
		case <-tk.C:
			fdone, err := f()
			if err != nil {
				return err
			}
			if fdone {
				return nil
			}
		}
	}
}
