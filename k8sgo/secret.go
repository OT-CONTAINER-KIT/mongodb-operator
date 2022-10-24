package k8sgo

import (
	"context"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// secretsParameters is an interface for secret input
type secretsParameters struct {
	Name        string
	OwnerDef    metav1.OwnerReference
	Data        string
	Namespace   string
	Labels      map[string]string
	Annotations map[string]string
	SecretsMeta metav1.ObjectMeta
	SecretName  string
	SecretKey   string
}

// CreateSecret is a method to create secret
func CreateSecret(params secretsParameters, key string) error {
	secretDef := generateSecret(params, key)
	logger := logGenerator(params.Name, params.Namespace, "Secret")
	_, err := generateK8sClient().CoreV1().Secrets(params.Namespace).Create(context.TODO(), secretDef, metav1.CreateOptions{})
	if err != nil {
		logger.Error(err, "MongoDB secret creation is failed")
		return err
	}
	logger.Info("MongoDB secret creation is successful")
	return nil
}

// generateSecret is a method that will generate a secret interface
func generateSecret(params secretsParameters, key string) *corev1.Secret {
	data := []byte(params.Data)
	secret := &corev1.Secret{
		TypeMeta:   generateMetaInformation("Secret", "v1"),
		ObjectMeta: params.SecretsMeta,
		Data: map[string][]byte{
			key: data,
		},
	}
	AddOwnerRefToObject(secret, params.OwnerDef)
	return secret
}

// getMongoDBPassword method will return the mongodb password
func getMongoDBPassword(params secretsParameters) string {
	logger := logGenerator(params.Name, params.Namespace, "Secret")
	secretName, err := generateK8sClient().CoreV1().Secrets(params.Namespace).Get(context.TODO(), params.SecretName, metav1.GetOptions{})
	if err != nil {
		logger.Error(err, "Failed in getting existing secret for mongodb admin")
	}
	value := string(secretName.Data[params.SecretKey])
	return value
}

//nolint:gosimple
// CheckSecretExist is a method to check secret exists
func CheckSecretExist(namespace string, secret string) bool {
	_, err := generateK8sClient().CoreV1().Secrets(namespace).Get(context.TODO(), secret, metav1.GetOptions{})
	if err != nil {
		return false
	}
	return true
}

// ReadStringData reads the StringData field of the secret with the given objectKey
func ReadStringData(namespace string, secretName string) (map[string]string, error) {
	secret, err := generateK8sClient().CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return dataToStringData(secret.Data), nil
}

func dataToStringData(data map[string][]byte) map[string]string {
	stringData := make(map[string]string)
	for k, v := range data {
		stringData[k] = string(v)
	}
	return stringData
}

func ReadKey(secretName string, key string, data map[string]string) (string, error) {
	if val, ok := data[key]; ok {
		return val, nil
	}
	return "", errors.Errorf(`key "%s" not present in the Secret %s`, key, secretName)
}
