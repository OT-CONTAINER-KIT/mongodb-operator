package k8sgo

import (
	"crypto/sha256"
	"fmt"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	opstreelabsinv1alpha1 "mongodb-operator/api/v1alpha1"
	"strings"
)

const (
	tlsCAMountPath             = "/var/lib/tls/ca/"
	tlsCACertName              = "ca.crt"
	tlsOperatorSecretMountPath = "/var/lib/tls/server/" //nolint
	tlsSecretCertName          = "tls.crt"              //nolint
	tlsSecretKeyName           = "tls.key"
	tlsSecretPemName           = "tls.pem"
	tlsCAVolumeName            = "tls-ca"
	tlsCertVolumeName          = "tls-secret"
)

func ValidateTLSConfig(instance *opstreelabsinv1alpha1.MongoDBCluster) (bool, error) {
	if !instance.Spec.Security.TLS.Enabled {
		return true, nil
	}

	log.Info("Ensuring TLS is correctly configured")

	// Ensure CA cert is configured
	_, err := getCaCrt(instance)

	if err != nil {
		if apiErrors.IsNotFound(err) {
			log.Error(err, "CA resource not found")
			return false, nil
		}

		return false, err
	}

	// Ensure Secret exists
	_, err = ReadStringData(instance.Namespace, instance.Spec.Security.TLS.CertificateKeySecret.Name)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			log.Error(err, "CertificateKeySecret not found")
			return false, nil
		}

		return false, err
	}

	// validate whether the secret contains "tls.crt" and "tls.key", or it contains "tls.pem"
	// if it contains all three, then the pem entry should be equal to the concatenation of crt and key
	_, err = getPemOrConcatenatedCrtAndKey(instance, instance.Spec.Security.TLS.CertificateKeySecret.Name)
	if err != nil {
		return false, err
	}

	log.Info("Successfully validated TLS config")

	return true, nil
}

func getCaCrt(instance *opstreelabsinv1alpha1.MongoDBCluster) (string, error) {
	var caData map[string]string
	var err error
	if instance.Spec.Security.TLS.CaCertificateSecret != nil {
		caData, err = ReadStringData(instance.Namespace, instance.Spec.Security.TLS.CaCertificateSecret.Name)
	} else {
		caData, err = ReadData(instance.Namespace, instance.Spec.Security.TLS.CaConfigMap.Name)
	}
	if err != nil {
		return "", err
	}

	if cert, ok := caData[tlsCACertName]; !ok || cert == "" {
		return "", errors.New("CA certificate resource should have a CA certificate in field ")
	} else {
		return cert, nil
	}
}

// getPemOrConcatenatedCrtAndKey will get the final PEM to write to the secret.
// This is either the tls.pem entry in the given secret, or the concatenation
// of tls.crt and tls.key
// It performs a basic validation on the entries.
func getPemOrConcatenatedCrtAndKey(instance *opstreelabsinv1alpha1.MongoDBCluster, secretName string) (string, error) {
	data, err := ReadStringData(instance.Namespace, secretName)
	if err != nil {
		return "", err
	}
	certKey := getCertAndKey(data, secretName)
	pem := getPem(data, secretName)
	if certKey == "" && pem == "" {
		return "", fmt.Errorf(`neither "%s" nor the pair "%s"/"%s" were present in the TLS secret`, tlsSecretPemName, tlsSecretCertName, tlsSecretKeyName)
	}
	if certKey == "" {
		return pem, nil
	}
	if pem == "" {
		return certKey, nil
	}
	if certKey != pem {
		return "", fmt.Errorf(`if all of "%s", "%s" and "%s" are present in the secret, the entry for "%s" must be equal to the concatenation of "%s" with "%s"`, tlsSecretCertName, tlsSecretKeyName, tlsSecretPemName, tlsSecretPemName, tlsSecretCertName, tlsSecretKeyName)
	}
	return certKey, nil
}

// getCertAndKey will fetch the certificate and key from the user-provided Secret.
func getCertAndKey(data map[string]string, secretName string) string {
	cert, err := ReadKey(secretName, tlsSecretCertName, data)
	if err != nil {
		return ""
	}
	key, err := ReadKey(secretName, tlsSecretKeyName, data)
	if err != nil {
		return ""
	}
	return combineCertificateAndKey(cert, key)
}

func combineCertificateAndKey(cert, key string) string {
	trimmedCert := strings.TrimRight(cert, "\n")
	trimmedKey := strings.TrimRight(key, "\n")
	return fmt.Sprintf("%s\n%s", trimmedCert, trimmedKey)
}

// getPem will fetch the pem from the user-provided secret
func getPem(data map[string]string, secretName string) string {
	pem, err := ReadKey(secretName, tlsSecretPemName, data)
	if err != nil {
		return ""
	}
	return pem
}

// ensureTLSResources creates any required TLS resources that the MongoDBCommunity
// requires for TLS configuration.
func EnsureTLSResources(instance *opstreelabsinv1alpha1.MongoDBCluster) error {
	if !instance.Spec.Security.TLS.Enabled {
		return nil
	}
	// the TLS secret needs to be created beforehand, as both the StatefulSet and AutomationConfig
	// require the contents.

	log.Info("TLS is enabled, creating/updating CA secret")
	if err := ensureCASecret(instance); err != nil {
		return errors.Errorf("could not ensure CA secret: %s", err)
	}

	log.Info("TLS is enabled, creating/updating TLS secret")
	if err := ensureTLSSecret(instance); err != nil {
		return errors.Errorf("could not ensure TLS secret: %s", err)
	}

	return nil
}

// ensureCASecret will create or update the operator managed Secret containing
// the CA certficate from the user provided Secret or ConfigMap.
func ensureCASecret(instance *opstreelabsinv1alpha1.MongoDBCluster) error {
	cert, err := getCaCrt(instance)
	if err != nil {
		return err
	}

	//caFileName := tlsOperatorSecretFileName(cert)
	labels := map[string]string{
		"app":           instance.Name,
		"mongodb_setup": "cluster",
		"role":          "cluster",
	}
	params := secretsParameters{
		SecretsMeta: generateObjectMetaInformation(instance.Name, instance.Namespace, labels, generateAnnotations()),
		OwnerDef:    mongoClusterAsOwner(instance),
		Namespace:   instance.Namespace,
		Labels:      labels,
		Annotations: generateAnnotations(),
		Name:        instance.Name + "-ca-certificate",
		Data:        cert,
	}

	return CreateSecret(params, tlsCACertName)
}

// ensureTLSSecret will create or update the operator-managed Secret containing
// the concatenated certificate and key from the user-provided Secret.
func ensureTLSSecret(instance *opstreelabsinv1alpha1.MongoDBCluster) error {
	certKey, err := getPemOrConcatenatedCrtAndKey(instance, instance.Spec.Security.TLS.CertificateKeySecret.Name)
	if err != nil {
		return err
	}
	// Calculate file name from certificate and key
	//fileName := tlsOperatorSecretFileName(certKey)
	labels := map[string]string{
		"app":           instance.Name,
		"mongodb_setup": "cluster",
		"role":          "cluster",
	}
	params := secretsParameters{
		SecretsMeta: generateObjectMetaInformation(instance.Name, instance.Namespace, labels, generateAnnotations()),
		OwnerDef:    mongoClusterAsOwner(instance),
		Namespace:   instance.Namespace,
		Labels:      labels,
		Annotations: generateAnnotations(),
		Name:        instance.Name + "-server-certificate-key",
		Data:        certKey,
	}

	return CreateSecret(params, tlsSecretCertName)
}

// tlsOperatorSecretFileName calculates the file name to use for the mounted
// certificate-key file. The name is based on the hash of the combined cert and key.
// If the certificate or key changes, the file path changes as well which will trigger
// the agent to perform a restart.
// The user-provided secret is being watched and will trigger a reconciliation
// on changes. This enables the operator to automatically handle cert rotations.
func tlsOperatorSecretFileName(certKey string) string {
	hash := sha256.Sum256([]byte(certKey))
	return fmt.Sprintf("%x.pem", hash)
}

func getVolumeFromSecret(volumeName, secretName string) []corev1.Volume {
	permission := int32(416)
	return []corev1.Volume{
		{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  secretName,
					DefaultMode: &permission,
				},
			},
		},
	}
}
