package istio

import (
	"fmt"

	istiov1alpha1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1alpha1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/util"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	appsv1 "k8s.io/api/apps/v1"
	autoscalev2beta1 "k8s.io/api/autoscaling/v2beta1"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileIstio) ReconcileGateways(log logr.Logger, istio *istiov1alpha1.Istio) error {

	gatewayResources := make(map[string]runtime.Object)

	for _, gw := range []string{"ingressgateway", "egressgateway"} {
		gatewaySa := &apiv1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("istio-%s-service-account", gw),
				Namespace: istio.Namespace,
				Labels: map[string]string{
					"app": gw,
				},
			},
		}
		controllerutil.SetControllerReference(istio, gatewaySa, r.scheme)
		gatewayResources[gatewaySa.Name] = gatewaySa

		gatewayCr := &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("istio-%s-cluster-role", gw),
				Labels: map[string]string{
					"app": "istio",
				},
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{"extensions"},
					Resources: []string{"thirdpartyresources", "virtualservices", "destinationrules", "gateways"},
					Verbs:     []string{"get", "watch", "list", "update"},
				},
			},
		}
		controllerutil.SetControllerReference(istio, gatewayCr, r.scheme)
		gatewayResources[gatewayCr.Name] = gatewayCr

		gatewayCrb := &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("istio-%s-cluster-role-binding", gw),
			},
			RoleRef: rbacv1.RoleRef{
				Kind:     "ClusterRole",
				APIGroup: "rbac.authorization.k8s.io",
				Name:     fmt.Sprintf("istio-%s-cluster-role", gw),
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      fmt.Sprintf("istio-%s-service-account", gw),
					Namespace: istio.Namespace,
				},
			},
		}
		controllerutil.SetControllerReference(istio, gatewayCrb, r.scheme)
		gatewayResources[gatewayCrb.Name] = gatewayCrb

		gatewayDeploy := gatewayDeployment(gw, istio.Namespace)
		controllerutil.SetControllerReference(istio, gatewayDeploy, r.scheme)
		gatewayResources[gatewayDeploy.Name] = gatewayDeploy

		gatewaySvc := &apiv1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("istio-%s", gw),
				Namespace: istio.Namespace,
				Labels: map[string]string{
					"app":   fmt.Sprintf("istio-%s", gw),
					"istio": gw,
				},
			},
			Spec: apiv1.ServiceSpec{
				Type:           apiv1.ServiceTypeLoadBalancer,
				LoadBalancerIP: "",
				Ports:          servicePorts(gw),
				Selector: map[string]string{
					"app":   fmt.Sprintf("istio-%s", gw),
					"istio": gw,
				},
			},
		}
		controllerutil.SetControllerReference(istio, gatewaySvc, r.scheme)
		gatewayResources[gatewaySvc.Name] = gatewaySvc

		gatewayAutoscaler := &autoscalev2beta1.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("istio-%s-autoscaler", gw),
				Namespace: istio.Namespace,
			},
			Spec: autoscalev2beta1.HorizontalPodAutoscalerSpec{
				MaxReplicas: 5,
				MinReplicas: util.IntPointer(1),
				ScaleTargetRef: autoscalev2beta1.CrossVersionObjectReference{
					Name:       fmt.Sprintf("istio-%s-deployment", gw),
					Kind:       "Deployment",
					APIVersion: "apps/v1",
				},
				Metrics: targetAvgCpuUtil80(),
			},
		}
		controllerutil.SetControllerReference(istio, gatewayAutoscaler, r.scheme)
		gatewayResources[gatewayAutoscaler.Name] = gatewayAutoscaler
	}

	for name, res := range gatewayResources {
		err := k8sutil.ReconcileResource(log, r.client, istio.Namespace, name, res)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", res.GetObjectKind().GroupVersionKind().Kind, "name", name)
		}
	}

	return nil
}

func gatewayDeployment(gw, ns string) *appsv1.Deployment {
	istioProxy := apiv1.Container{
		Name:            "istio-proxy",
		Image:           "docker.io/istio/proxyv2:1.0.5",
		ImagePullPolicy: apiv1.PullIfNotPresent,
		Args: []string{
			"proxy",
			"router",
			"-v", "2",
			"--discoveryRefreshDelay", "1s",
			"--drainDuration", "45s",
			"--parentShutdownDuration", "1m0s",
			"--connectTimeout", "10s",
			"--serviceCluster", fmt.Sprintf("istio-%s", gw),
			"--zipkinAddress", fmt.Sprintf("zipkin.%s:9411", ns),
			"--proxyAdminPort", "15000",
			"--controlPlaneAuthPolicy", "NONE",
			"--discoveryAddress", fmt.Sprintf("istio-pilot.%s:8080", ns),
		},
		Env: append(istioProxyEnv(), apiv1.EnvVar{
			Name: "ISTIO_META_POD_NAME",
			ValueFrom: &apiv1.EnvVarSource{
				FieldRef: &apiv1.ObjectFieldSelector{

					FieldPath: "metadata.name",
				},
			},
		}),
		Resources: defaultResources(),
		VolumeMounts: []apiv1.VolumeMount{
			{
				Name:      "istio-certs",
				MountPath: "/etc/certs",
				ReadOnly:  true,
			},
			{
				Name:      fmt.Sprintf("%s-certs", gw),
				MountPath: fmt.Sprintf("/etc/istio/%s-certs", gw),
				ReadOnly:  true,
			},
			{
				Name:      fmt.Sprintf("%s-ca-certs", gw),
				MountPath: fmt.Sprintf("/etc/istio/%s-ca-certs", gw),
				ReadOnly:  true,
			},
		},
	}

	switch gw {
	case "ingressgateway":
		istioProxy.Ports = []apiv1.ContainerPort{
			{ContainerPort: 80},
			{ContainerPort: 443},
			{ContainerPort: 31400},
			{ContainerPort: 15011},
			{ContainerPort: 8060},
			{ContainerPort: 853},
			{ContainerPort: 15030},
			{ContainerPort: 15031},
			{ContainerPort: 15090, Protocol: apiv1.ProtocolTCP, Name: "http-envoy-prom"},
		}
	case "egressgateway":
		istioProxy.Ports = []apiv1.ContainerPort{
			{ContainerPort: 80},
			{ContainerPort: 443},
			{ContainerPort: 15090, Protocol: apiv1.ProtocolTCP, Name: "http-envoy-prom"},
		}
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("istio-%s-deployment", gw),
			Namespace: ns,
			Labels: map[string]string{
				"app":   fmt.Sprintf("istio-%s", gw),
				"istio": gw,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: util.IntPointer(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":   fmt.Sprintf("istio-%s", gw),
					"istio": gw,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":   fmt.Sprintf("istio-%s", gw),
						"istio": gw,
					},
					Annotations: defaultDeployAnnotations(),
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: fmt.Sprintf("istio-%s-service-account", gw),
					Containers: []apiv1.Container{
						istioProxy,
					},
					Volumes: []apiv1.Volume{
						{
							Name: "istio-certs",
							VolumeSource: apiv1.VolumeSource{
								Secret: &apiv1.SecretVolumeSource{
									SecretName: fmt.Sprintf("istio.istio-%s-service-account", gw),
									Optional:   util.BoolPointer(true),
								},
							},
						},
						{
							Name: fmt.Sprintf("%s-certs", gw),
							VolumeSource: apiv1.VolumeSource{
								Secret: &apiv1.SecretVolumeSource{
									SecretName: fmt.Sprintf("istio-%s-certs", gw),
									Optional:   util.BoolPointer(true),
								},
							},
						},
						{
							Name: fmt.Sprintf("%s-ca-certs", gw),
							VolumeSource: apiv1.VolumeSource{
								Secret: &apiv1.SecretVolumeSource{
									SecretName: fmt.Sprintf("istio-%s-ca-certs", gw),
									Optional:   util.BoolPointer(true),
								},
							},
						},
					},
					Affinity: &apiv1.Affinity{},
				},
			},
		},
	}
}

func servicePorts(gw string) []apiv1.ServicePort {
	switch gw {
	case "ingressgateway":
		return []apiv1.ServicePort{
			{Port: 80, TargetPort: intstr.FromInt(80), Name: "http2", NodePort: 31380},
			{Port: 443, Name: "https", NodePort: 31390},
			{Port: 31400, Name: "tcp", NodePort: 31400},
			{Port: 15011, TargetPort: intstr.FromInt(15011), Name: "tcp-pilot-grpc-tls"},
			{Port: 8060, TargetPort: intstr.FromInt(8060), Name: "tcp-citadel-grpc-tls"},
			{Port: 853, TargetPort: intstr.FromInt(853), Name: "tcp-dns-tls"},
			{Port: 15030, TargetPort: intstr.FromInt(15030), Name: "http2-prometheus"},
			{Port: 15031, TargetPort: intstr.FromInt(15031), Name: "http2-grafana"},
		}
	case "egressgateway":
		return []apiv1.ServicePort{
			{Port: 80, Name: "http2"},
			{Port: 443, Name: "https"},
		}
	}
	return []apiv1.ServicePort{}
}
