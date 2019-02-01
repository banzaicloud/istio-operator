package istio

import (
	"fmt"

	istiov1alpha1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1alpha1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/util"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	appsv1 "k8s.io/api/apps/v1"
	autoscalev2beta1 "k8s.io/api/autoscaling/v2beta1"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	networkingv1alpha3 "istio.io/api/pkg/kube/apis/networking/v1alpha3"
	"istio.io/api/networking/v1alpha3"
)

func (r *ReconcileIstio) ReconcilePilot(log logr.Logger, istio *istiov1alpha1.Istio) error {

	pilotResources := make(map[string]runtime.Object)

	pilotSa := &apiv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-pilot-service-account",
			Namespace: istio.Namespace,
			Labels: map[string]string{
				"app": "istio-pilot",
			},
		},
	}
	controllerutil.SetControllerReference(istio, pilotSa, r.scheme)
	pilotResources[pilotSa.Name] = pilotSa

	pilotCr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "istio-pilot-cluster-role",
			Labels: map[string]string{
				"app": "istio-pilot",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"config.istio.io"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"rbac.istio.io"},
				Resources: []string{"*"},
				Verbs:     []string{"get", "watch", "list"},
			},
			{
				APIGroups: []string{"networking.istio.io"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"authentication.istio.io"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"apiextensions.k8s.io"},
				Resources: []string{"customresourcedefinitions"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"extensions"},
				Resources: []string{"thirdpartyresources", "thirdpartyresources.extensions", "ingresses", "ingresses/status"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
				Verbs:     []string{"create", "get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"endpoints", "pods", "services"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"namespaces", "nodes", "secrets"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}
	controllerutil.SetControllerReference(istio, pilotCr, r.scheme)
	pilotResources[pilotCr.Name] = pilotCr

	pilotCrb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "istio-pilot-cluster-role-binding",
			Labels: map[string]string{
				"app": "istio-pilot",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     pilotCr.Name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      pilotSa.Name,
				Namespace: istio.Namespace,
			},
		},
	}
	controllerutil.SetControllerReference(istio, pilotCrb, r.scheme)
	pilotResources[pilotCrb.Name] = pilotCrb

	meshConfig, err := meshConfig(istio.Namespace)
	if err != nil {
		return emperror.With(err)
	}
	istioCm := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio",
			Namespace: istio.Namespace,
			Labels: map[string]string{
				"app": "istio",
			},
		},
		Data: map[string]string{
			"mesh": meshConfig,
		},
	}
	controllerutil.SetControllerReference(istio, istioCm, r.scheme)
	pilotResources[istioCm.Name] = istioCm

	pilotDeploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-pilot-deployment",
			Namespace: istio.Namespace,
			Labels: map[string]string{
				"app":   "istio-pilot",
				"istio": "pilot",
			},
			// annotations:
			//    checksum/config-volume: {{ template "istio.configmap.checksum" . }}
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: util.IntPointer(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"istio": "pilot",
					"app":   "pilot",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"istio": "pilot",
						"app":   "pilot",
					},
					Annotations: defaultDeployAnnotations(),
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: pilotSa.Name,
					Containers: []apiv1.Container{
						{
							Name:            "discovery",
							Image:           "docker.io/istio/pilot:1.0.5",
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Args: []string{
								"discovery",
							},
							Ports: []apiv1.ContainerPort{
								{ContainerPort: 8080},
								{ContainerPort: 15010},
							},
							ReadinessProbe: &apiv1.Probe{
								Handler: apiv1.Handler{
									HTTPGet: &apiv1.HTTPGetAction{
										Path: "/ready",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       30,
								TimeoutSeconds:      5,
							},
							Env: []apiv1.EnvVar{
								{
									Name: "POD_NAME",
									ValueFrom: &apiv1.EnvVarSource{
										FieldRef: &apiv1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.name",
										},
									},
								},
								{
									Name: "POD_NAMESPACE",
									ValueFrom: &apiv1.EnvVarSource{
										FieldRef: &apiv1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.namespace",
										},
									},
								},
								{Name: "PILOT_CACHE_SQUASH", Value: "5"},
								{Name: "PILOT_PUSH_THROTTLE_COUNT", Value: "100"},
								{Name: "GODEBUG", Value: "gctrace=2"},
								{Name: "PILOT_TRACE_SAMPLING", Value: "1.0"},
							},
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("500m"),
									apiv1.ResourceMemory: resource.MustParse("2048Mi"),
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "config-volume",
									MountPath: "/etc/istio/config",
								},
								{
									Name:      "istio-certs",
									MountPath: "/etc/certs",
									ReadOnly:  true,
								},
							},
						},
						{
							Name:            "istio-proxy",
							Image:           "docker.io/istio/proxyv2:1.0.5",
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Ports: []apiv1.ContainerPort{
								{ContainerPort: 15003},
								{ContainerPort: 15005},
								{ContainerPort: 15007},
								{ContainerPort: 15011},
							},
							Args: []string{
								"proxy",
								"--serviceCluster",
								"istio-pilot",
								"--templateFile",
								"/etc/istio/proxy/envoy_pilot.yaml.tmpl",
								"--controlPlaneAuthPolicy",
								"NONE",
							},
							Env:       istioProxyEnv(),
							Resources: defaultResources(),
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "istio-certs",
									MountPath: "/etc/certs",
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
									SecretName: fmt.Sprintf("istio.%s", pilotSa.Name),
									Optional:   util.BoolPointer(true),
								},
							},
						},
						{
							Name: "config-volume",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: "istio",
									},
								},
							},
						},
					},
					Affinity: &apiv1.Affinity{},
				},
			},
		},
	}
	controllerutil.SetControllerReference(istio, pilotDeploy, r.scheme)
	pilotResources[pilotDeploy.Name] = pilotDeploy

	pilotSvc := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-pilot",
			Namespace: istio.Namespace,
			Labels: map[string]string{
				"app": "istio-pilot",
			},
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{Name: "grpc-xds", Port: 15010},
				{Name: "https-xds", Port: 15011},
				{Name: "http-legacy-discovery", Port: 8080},
				{Name: "http-monitoring", Port: 9093},
			},
			Selector: map[string]string{
				"istio": "pilot",
			},
		},
	}
	controllerutil.SetControllerReference(istio, pilotSvc, r.scheme)
	pilotResources[pilotSvc.Name] = pilotSvc

	pilotAutoscaler := &autoscalev2beta1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-pilot-autoscaler",
			Namespace: istio.Namespace,
		},
		Spec: autoscalev2beta1.HorizontalPodAutoscalerSpec{
			MaxReplicas: 5,
			MinReplicas: util.IntPointer(1),
			ScaleTargetRef: autoscalev2beta1.CrossVersionObjectReference{
				Name:       "istio-pilot-deployment",
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
			Metrics: targetAvgCpuUtil80(),
		},
	}
	controllerutil.SetControllerReference(istio, pilotAutoscaler, r.scheme)
	pilotResources[pilotAutoscaler.Name] = pilotAutoscaler

	// wait until galley is up and running, otherwise admission controller will fail

	defaultGw := &networkingv1alpha3.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-autogenerated-k8s-ingress",
			Namespace: istio.Namespace,
		},
		Spec: v1alpha3.Gateway{
			Servers: []*v1alpha3.Server{
				{
					Port: &v1alpha3.Port{
						Name:     "http",
						Protocol: "HTTP2",
						Number:   80,
					},
					Hosts: []string{"*"},
				},
			},
			Selector: map[string]string{
				"istio": "ingress",
			},
		},
	}
	controllerutil.SetControllerReference(istio, defaultGw, r.scheme)
	pilotResources[defaultGw.Name] = defaultGw

	for name, res := range pilotResources {
		err := k8sutil.ReconcileResource(log, r.client, istio.Namespace, name, res)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", res.GetObjectKind().GroupVersionKind().Kind, "name", name)
		}
	}

	return nil
}

func meshConfig(ns string) (string, error) {
	meshConfig := map[string]interface{}{
		"disablePolicyChecks": false,
		"enableTracing":       true,
		"accessLogFile":       "/dev/stdout",
		"mixerCheckServer":    fmt.Sprintf("istio-policy.%s.svc.cluster.local:9091", ns),
		"mixerReportServer":   fmt.Sprintf("istio-telemetry.%s.svc.cluster.local:9091", ns),
		"policyCheckFailOpen": false,
		"sdsUdsPath":          "",
		"sdsRefreshDelay":     "15s",
		"defaultConfig": map[string]interface{}{
			"connectTimeout":         "10s",
			"configPath":             "/etc/istio/proxy",
			"binaryPath":             "/usr/local/bin/envoy",
			"serviceCluster":         "istio-proxy",
			"drainDuration":          "45s",
			"parentShutdownDuration": "1m0s",
			"proxyAdminPort":         15000,
			"concurrency":            0,
			"zipkinAddress":          fmt.Sprintf("zipkin.%s:9411", ns),
			"controlPlaneAuthPolicy": "NONE",
			"discoveryAddress":       fmt.Sprintf("istio-pilot.%s:15007", ns),
		},
	}
	marshaledConfig, err := yaml.Marshal(meshConfig)
	if err != nil {
		return "", emperror.Wrap(err, "failed to marshal istio config")
	}
	return string(marshaledConfig), nil
}
