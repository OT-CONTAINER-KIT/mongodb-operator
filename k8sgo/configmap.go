package k8sgo

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReadData extracts the contents of the Data field in a given config map
func ReadData(namespace string, cmName string) (map[string]string, error) {
	configmap, err := generateK8sClient().CoreV1().ConfigMaps(namespace).Get(context.TODO(), cmName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return configmap.Data, nil
}

func generateHash(source []byte) string {
	hash := sha1.New()
	hash.Write(source)
	return hex.EncodeToString(hash.Sum(nil))
}

// CheckConfigMapExist is a method to check configmap exists
//func CheckConfigMapExist(namespace string, cmName string) bool {
//	config, err := generateK8sClient().CoreV1().ConfigMaps(namespace).Get(context.TODO(), cmName, metav1.GetOptions{})
//	if err != nil {
//		return false
//	}
//	return true
//}
