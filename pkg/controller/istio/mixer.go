package istio

import (
	istiov1alpha1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1alpha1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/util"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"k8s.io/apimachinery/pkg/util/intstr"
	autoscalev2beta1 "k8s.io/api/autoscaling/v2beta1"
	"fmt"
)

func (r *ReconcileIstio) ReconcileMixer(log logr.Logger, istio *istiov1alpha1.Istio) error {

	mixerResources := make(map[string]runtime.Object)

	mixerSa := &apiv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-mixer-service-account",
			Namespace: istio.Namespace,
			Labels: map[string]string{
				"app": "mixer",
			},
		},
	}
	controllerutil.SetControllerReference(istio, mixerSa, r.scheme)
	mixerResources[mixerSa.Name] = mixerSa

	mixerCr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "istio-mixer-cluster-role",
			Labels: map[string]string{
				"app": "mixer",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"config.istio.io"},
				Resources: []string{"*"},
				Verbs:     []string{"create", "get", "list", "watch", "patch"},
			},
			{
				APIGroups: []string{"rbac.istio.io"},
				Resources: []string{"*"},
				Verbs:     []string{"get", "watch", "list"},
			},
			{
				APIGroups: []string{"apiextensions.k8s.io"},
				Resources: []string{"customresourcedefinitions"},
				Verbs:     []string{"get", "watch", "list"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps", "endpoints", "pods", "services", "namespaces", "secrets", "replicationcontrollers"},
				Verbs:     []string{"get", "watch", "list"},
			},
			{
				APIGroups: []string{"extensions"},
				Resources: []string{"replicasets"},
				Verbs:     []string{"get", "watch", "list"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"replicasets"},
				Verbs:     []string{"get", "watch", "list"},
			},
		},
	}
	controllerutil.SetControllerReference(istio, mixerCr, r.scheme)
	mixerResources[mixerCr.Name] = mixerCr

	mixerCrb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "istio-mixer-cluster-role-binding",
			Labels: map[string]string{
				"app": "mixer",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     mixerCr.Name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      mixerSa.Name,
				Namespace: istio.Namespace,
			},
		},
	}
	controllerutil.SetControllerReference(istio, mixerCrb, r.scheme)
	mixerResources[mixerCrb.Name] = mixerCrb

	policyDeploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-policy-deployment",
			Namespace: istio.Namespace,
			Labels: map[string]string{
				"istio": "mixer",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: util.IntPointer(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"istio":            "mixer",
					"app":              "policy",
					"istio-mixer-type": "policy",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"istio":            "mixer",
						"app":              "policy",
						"istio-mixer-type": "policy",
					},
					Annotations: defaultDeployAnnotations(),
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: mixerSa.Name,
					Volumes:            mixerVolumes(mixerSa.Name),
					Affinity:           &apiv1.Affinity{},
					Containers: []apiv1.Container{
						mixerContainer(true, istio.Namespace),
						istioProxyContainer("policy"),
					},
				},
			},
		},
	}
	controllerutil.SetControllerReference(istio, policyDeploy, r.scheme)
	mixerResources[policyDeploy.Name] = policyDeploy

	telemetryDeploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-telemetry-deployment",
			Namespace: istio.Namespace,
			Labels: map[string]string{
				"istio": "mixer",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: util.IntPointer(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"istio":            "mixer",
					"app":              "telemetry",
					"istio-mixer-type": "telemetry",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"istio":            "mixer",
						"app":              "telemetry",
						"istio-mixer-type": "telemetry",
					},
					Annotations: defaultDeployAnnotations(),
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: mixerSa.Name,
					Volumes:            mixerVolumes(mixerSa.Name),
					Affinity:           &apiv1.Affinity{},
					Containers: []apiv1.Container{
						mixerContainer(false, istio.Namespace),
						istioProxyContainer("telemetry"),
					},
				},
			},
		},
	}
	controllerutil.SetControllerReference(istio, telemetryDeploy, r.scheme)
	mixerResources[telemetryDeploy.Name] = telemetryDeploy

	for _, mixer := range []string{"policy", "telemetry"} {
		mixerSvc := mixerService(mixer, istio.Namespace)
		controllerutil.SetControllerReference(istio, mixerSvc, r.scheme)
		mixerResources[mixerSvc.Name] = mixerSvc

		mixerAs := mixerAutoscaler(mixer, istio.Namespace)
		controllerutil.SetControllerReference(istio, mixerAs, r.scheme)
		mixerResources[mixerAs.Name] = mixerAs
	}

	// not sure if it's needed or not
	statsdPromBridgeCm := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-statsd-prom-bridge",
			Namespace: istio.Namespace,
			Labels: map[string]string{
				"app":   "istio-statsd-prom-bridge",
				"istio": "mixer",
			},
		},
		Data: map[string]string{
			"mapping.conf": "",
		},
	}
	controllerutil.SetControllerReference(istio, statsdPromBridgeCm, r.scheme)
	mixerResources[statsdPromBridgeCm.Name] = statsdPromBridgeCm

	for name, res := range mixerResources {
		err := k8sutil.ReconcileResource(log, r.client, istio.Namespace, name, res)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", res.GetObjectKind().GroupVersionKind().Kind, "name", name)
		}
	}
	return nil
}

func mixerVolumes(serviceAccount string) []apiv1.Volume {
	return []apiv1.Volume{
		{
			Name: "istio-certs",
			VolumeSource: apiv1.VolumeSource{
				Secret: &apiv1.SecretVolumeSource{
					SecretName: fmt.Sprintf("istio.%s", serviceAccount),
					Optional:   util.BoolPointer(true),
				},
			},
		},
		{
			Name: "uds-socket",
			VolumeSource: apiv1.VolumeSource{
				EmptyDir: &apiv1.EmptyDirVolumeSource{},
			},
		},
	}
}

func mixerContainer(policy bool, ns string) apiv1.Container {
	c := apiv1.Container{
		Name:            "mixer",
		Image:           "docker.io/istio/mixer:1.0.5",
		ImagePullPolicy: apiv1.PullIfNotPresent,
		Ports: []apiv1.ContainerPort{
			{ContainerPort: 9093},
			{ContainerPort: 42422},
		},
		Args: []string{
			"--address",
			"unix:///sock/mixer.socket",
			"--configStoreURL=k8s://",
			fmt.Sprintf("--configDefaultNamespace=%s", ns),
			"--trace_zipkin_url=http://zipkin:9411/api/v1/spans",
		},
		Env: []apiv1.EnvVar{
			{Name: "GODEBUG", Value: "gctrace=2"},
		},
		Resources: defaultResources(),
		VolumeMounts: []apiv1.VolumeMount{
			{Name: "uds-socket", MountPath: "/sock",},
		},
		LivenessProbe: &apiv1.Probe{
			Handler: apiv1.Handler{
				HTTPGet: &apiv1.HTTPGetAction{
					Path: "/version",
					Port: intstr.FromInt(9093),
				},
			},
			InitialDelaySeconds: 5,
			PeriodSeconds:       5,
		},
	}
	if policy {
		c.Args = append(c.Args, "--numCheckCacheEntries=0")
	}
	return c
}

func istioProxyContainer(mixer string) apiv1.Container {
	return apiv1.Container{
		Name:            "istio-proxy",
		Image:           "docker.io/istio/proxyv2:1.0.5",
		ImagePullPolicy: apiv1.PullIfNotPresent,
		Ports: []apiv1.ContainerPort{
			{ContainerPort: 9091},
			{ContainerPort: 15004},
			{ContainerPort: 15090, Protocol: apiv1.ProtocolTCP, Name: "http-envoy-prom"},
		},
		Args: []string{
			"proxy",
			"--serviceCluster",
			fmt.Sprintf("istio-%s", mixer),
			"--templateFile",
			fmt.Sprintf("/etc/istio/proxy/envoy_%s.yaml.tmpl", mixer),
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
			{
				Name:      "uds-socket",
				MountPath: "/sock",
			},
		},
	}
}

func mixerService(mixer, ns string) *apiv1.Service {
	svc := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("istio-%s", mixer),
			Namespace: ns,
			Labels: map[string]string{
				"istio": "mixer",
			},
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{Name: "grpc-mixer", Port: 9091},
				{Name: "grpc-mixer-mtls", Port: 15004},
				{Name: "http-monitoring", Port: 9093},
			},
			Selector: map[string]string{
				"istio":            "mixer",
				"istio-mixer-type": mixer,
			},
		},
	}
	if mixer == "telemetry" {
		svc.Spec.Ports = append(svc.Spec.Ports, apiv1.ServicePort{Name: "prometheus", Port: 42422})
	}
	return svc
}

func mixerAutoscaler(mixer, ns string) *autoscalev2beta1.HorizontalPodAutoscaler {
	return &autoscalev2beta1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("istio-%s-autoscaler", mixer),
			Namespace: ns,
		},
		Spec: autoscalev2beta1.HorizontalPodAutoscalerSpec{
			MaxReplicas: 5,
			MinReplicas: util.IntPointer(1),
			ScaleTargetRef: autoscalev2beta1.CrossVersionObjectReference{
				Name:       fmt.Sprintf("istio-%s-deployment", mixer),
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
			Metrics: targetAvgCpuUtil80(),
		},
	}
}
