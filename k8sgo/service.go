package k8sgo

import (
	"context"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	mongoDBPort           = 27017
	mongoDBMonitoringPort = 9216
)

// serviceParameters is a structure for service inputs
type serviceParameters struct {
	ServiceMeta     metav1.ObjectMeta
	OwnerDef        metav1.OwnerReference
	Labels          map[string]string
	Annotations     map[string]string
	Namespace       string
	HeadlessService bool
	Port            int32
	PortName        string
}

// CreateOrUpdateService method will create or update MongoDB service
func CreateOrUpdateService(params serviceParameters) error {
	logger := logGenerator(params.ServiceMeta.Name, params.Namespace, "Service")
	serviceDef := generateServiceDef(params)
	storedService, err := getService(params.Namespace, params.ServiceMeta.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(serviceDef); err != nil {
				logger.Error(err, "Unable to patch MongoDB service with compare annotations")
			}
			return createService(params.Namespace, serviceDef)
		}
		return err
	}
	return patchService(storedService, serviceDef, params.Namespace)
}

// patchService will patch Kubernetes service
func patchService(storedService *corev1.Service, newService *corev1.Service, namespace string) error {
	logger := logGenerator(storedService.Name, namespace, "Service")
	// adding meta fields
	newService.ResourceVersion = storedService.ResourceVersion
	newService.CreationTimestamp = storedService.CreationTimestamp
	newService.ManagedFields = storedService.ManagedFields
	newService.Spec.ClusterIP = storedService.Spec.ClusterIP

	patchResult, err := patch.DefaultPatchMaker.Calculate(storedService, newService,
		patch.IgnoreStatusFields(),
		patch.IgnoreField("kind"),
		patch.IgnoreField("apiVersion"),
	)
	if err != nil {
		logger.Error(err, "Unable to patch MongoDB service with comparison object")
		return err
	}
	if !patchResult.IsEmpty() {
		for key, value := range storedService.Annotations {
			if _, present := newService.Annotations[key]; !present {
				newService.Annotations[key] = value
			}
		}
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(newService); err != nil {
			logger.Error(err, "Unable to patch MongoDB service with comparison object")
			return err
		}
		logger.Info("Syncing MongoDB service with defined properties")
		return updateService(namespace, newService)
	}
	logger.Info("MongoDB service is already in-sync")
	return nil
}

// createService is a method to create service
func createService(namespace string, service *corev1.Service) error {
	logger := logGenerator(service.Name, namespace, "Service")
	_, err := generateK8sClient().CoreV1().Services(namespace).Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		logger.Error(err, "MongoDB service creation is failed")
		return err
	}
	logger.Info("MongoDB service creation is successful")
	return nil
}

// updateService is a method to update service
func updateService(namespace string, service *corev1.Service) error {
	logger := logGenerator(service.Name, namespace, "Service")
	_, err := generateK8sClient().CoreV1().Services(namespace).Update(context.TODO(), service, metav1.UpdateOptions{})
	if err != nil {
		logger.Error(err, "MongoDB service updation is failed")
		return err
	}
	logger.Info("MongoDB service updation is successful")
	return nil
}

// getService is a method to get service
func getService(namespace string, service string) (*corev1.Service, error) {
	logger := logGenerator(service, namespace, "Service")
	serviceInfo, err := generateK8sClient().CoreV1().Services(namespace).Get(context.TODO(), service, metav1.GetOptions{})
	if err != nil {
		logger.Info("MongoDB service get action is failed")
		return nil, err
	}
	logger.Info("MongoDB service get action is successful")
	return serviceInfo, nil
}

// generateServiceDef is a method to generate service definition
func generateServiceDef(params serviceParameters) *corev1.Service {
	service := &corev1.Service{
		TypeMeta:   generateMetaInformation("Service", "core/v1"),
		ObjectMeta: params.ServiceMeta,
		Spec: corev1.ServiceSpec{
			Selector: params.Labels,
			Ports: []corev1.ServicePort{
				{
					Name:       params.PortName,
					Port:       params.Port,
					TargetPort: intstr.FromInt(int(params.Port)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
	if params.HeadlessService {
		service.Spec.ClusterIP = "None"
	}
	return service
}
