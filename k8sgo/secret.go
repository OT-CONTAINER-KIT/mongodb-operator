package k8sgo

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// secretsParameters is an interface for secret input
type secretsParameters struct {
	Name        string
	OwnerDef    metav1.OwnerReference
	Password    string
	Namespace   string
	Labels      map[string]string
	Annotations map[string]string
	SecretsMeta metav1.ObjectMeta
	SecretName  string
}

// CreateSecret is a method to create secret
func CreateSecret(params secretsParameters) error {
	secretDef := generateSecret(params)
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
func generateSecret(params secretsParameters) *corev1.Secret {
	password := []byte(params.Password)
	secret := &corev1.Secret{
		TypeMeta:   generateMetaInformation("Secret", "v1"),
		ObjectMeta: params.SecretsMeta,
		Data: map[string][]byte{
			"password": password,
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
	value := string(secretName.Data["password"])
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
