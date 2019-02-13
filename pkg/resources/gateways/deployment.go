package gateways

import (
	"fmt"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) deployment(gw string, owner *istiov1beta1.Config) runtime.Object {
	return &appsv1.Deployment{
		ObjectMeta: templates.ObjectMeta(gatewayName(gw), labelSelector(gw), owner),
		Spec: appsv1.DeploymentSpec{
			Replicas: util.IntPointer(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labelSelector(gw),
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labelSelector(gw),
					Annotations: templates.DefaultDeployAnnotations(),
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: serviceAccountName(gw),
					Containers: []apiv1.Container{
						{
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
								"--zipkinAddress", fmt.Sprintf("zipkin.%s:9411", owner.Namespace),
								"--proxyAdminPort", "15000",
								"--controlPlaneAuthPolicy", "NONE",
								"--discoveryAddress", fmt.Sprintf("istio-pilot.%s:8080", owner.Namespace),
							},
							Ports: r.ports(gw),
							Env: append(templates.IstioProxyEnv(), apiv1.EnvVar{
								Name: "ISTIO_META_POD_NAME",
								ValueFrom: &apiv1.EnvVarSource{
									FieldRef: &apiv1.ObjectFieldSelector{

										FieldPath: "metadata.name",
									},
								},
							}),
							Resources: templates.DefaultResources(),
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
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "istio-certs",
							VolumeSource: apiv1.VolumeSource{
								Secret: &apiv1.SecretVolumeSource{
									SecretName: fmt.Sprintf("istio.%s", serviceAccountName(gw)),
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

func (r *Reconciler) ports(gw string) []apiv1.ContainerPort {
	switch gw {
	case "ingressgateway":
		return []apiv1.ContainerPort{
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
		return []apiv1.ContainerPort{
			{ContainerPort: 80},
			{ContainerPort: 443},
			{ContainerPort: 15090, Protocol: apiv1.ProtocolTCP, Name: "http-envoy-prom"},
		}
	}
	return nil
}
