package k8sgo

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func dealWithExpandingPVC(ctx context.Context, sts appsv1.StatefulSet) error {
	// get pods
	podList, err := generateK8sClient().CoreV1().Pods(sts.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.FormatLabels(sts.Labels),
	})
	if err != nil {
		return err
	}

	for _, item := range podList.Items {
		name := item.Name
		// get pvc
		pvc, err := generateK8sClient().CoreV1().PersistentVolumeClaims(sts.Namespace).Get(ctx, sts.Name+"-"+name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		// delete pod
		err1 := generateK8sClient().CoreV1().Pods(sts.Namespace).Delete(ctx, item.Name, metav1.DeleteOptions{})
		if err1 != nil {
			return err1
		}

		// execute expanding pvc
		err2 := doExpandPVC(ctx, pvc, sts)
		if err2 != nil {
			return err2
		}

		// rebuild pod
		err3 := doRebuildPod(ctx, &item, sts.Namespace)
		if err3 != nil {
			return err3
		}

	}

	return nil
}

func doRebuildPod(ctx context.Context, pod *corev1.Pod, namespace string) error {
	newPod := pod
	newPod.Annotations = nil
	newPod.ResourceVersion = ""
	newPod.UID = ""
	newPod.DeletionTimestamp = nil
	newPod.OwnerReferences = nil
	newPod.Status = corev1.PodStatus{}

	_, err := generateK8sClient().CoreV1().Pods(namespace).Create(ctx, newPod, metav1.CreateOptions{})
	if err != nil {
		Log.Info("Create failed ", "name", newPod.Name, "err", err)
		return err
	}

	err2 := retry(time.Second*2, time.Duration(waitLimit)*time.Second, func() (bool, error) {
		currentPod, err3 := generateK8sClient().CoreV1().Pods(namespace).Get(ctx, pod.Name, metav1.GetOptions{})
		if err3 != nil {
			return false, client.IgnoreNotFound(err3)
		}
		if currentPod.Status.Phase != "Running" {
			Log.Info("CurrentPod is not running yet", "name", currentPod.Name)
			return false, nil
		}
		for _, c := range currentPod.Status.ContainerStatuses {
			if !c.Ready {
				Log.Info("currentPod's image is not ready yet", "podName", currentPod.Name, "imageName", c.Image)
				return false, nil
			}
		}
		return true, nil
	})

	return err2
}
