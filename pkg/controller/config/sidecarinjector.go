package config

import (
	"fmt"
	"strings"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/util"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileConfig) ReconcileSidecarInjector(log logr.Logger, istio *istiov1beta1.Config) error {

	siResources := make(map[string]runtime.Object)

	siSa := &apiv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-sidecar-injector-service-account",
			Namespace: istio.Namespace,
			Labels: map[string]string{
				"app": "istio-sidecar-injector",
			},
		},
	}
	controllerutil.SetControllerReference(istio, siSa, r.scheme)
	siResources["serviceaccount."+siSa.Name] = siSa

	siCr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "istio-sidecar-injector-cluster-role",
			Labels: map[string]string{
				"app": "istio-sidecar-injector",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"configmaps"},
				Resources: []string{"*"},
				Verbs:     []string{"get", "watch", "list"},
			},
			{
				APIGroups: []string{"admissionregistration.k8s.io"},
				Resources: []string{"mutatingwebhookconfigurations"},
				Verbs:     []string{"get", "list", "watch", "patch"},
			},
		},
	}
	controllerutil.SetControllerReference(istio, siCr, r.scheme)
	siResources["clusterrole."+siCr.Name] = siCr

	siCrb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "istio-sidecar-injector-cluster-role-binding",
			Labels: map[string]string{
				"app": "istio-sidecar-injector",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     siCr.Name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      siSa.Name,
				Namespace: istio.Namespace,
			},
		},
	}
	controllerutil.SetControllerReference(istio, siCrb, r.scheme)
	siResources["clusterrolebinding."+siCrb.Name] = siCrb

	siConfig, err := siConfig()
	if err != nil {
		return emperror.With(err)
	}
	siCm := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-sidecar-injector-config",
			Namespace: istio.Namespace,
			Labels: map[string]string{
				"app":   "istio",
				"istio": "sidecar-injector",
			},
		},
		Data: map[string]string{
			"config": siConfig,
		},
	}
	controllerutil.SetControllerReference(istio, siCm, r.scheme)
	siResources["configmap."+siCm.Name] = siCm

	siDeploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-sidecar-injector-deployment",
			Namespace: istio.Namespace,
			Labels: map[string]string{
				"app":   "istio-sidecar-injector",
				"istio": "sidecar-injector",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: util.IntPointer(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"istio": "sidecar-injector",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"istio": "sidecar-injector",
					},
					Annotations: defaultDeployAnnotations(),
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: siSa.Name,
					Containers: []apiv1.Container{
						{
							Name:            "sidecar-injector-webhook",
							Image:           "docker.io/istio/sidecar_injector:1.0.5",
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Args: []string{
								"--caCertFile=/etc/istio/certs/root-cert.pem",
								"--tlsCertFile=/etc/istio/certs/cert-chain.pem",
								"--tlsKeyFile=/etc/istio/certs/key.pem",
								"--injectConfig=/etc/istio/inject/config",
								"--meshConfig=/etc/istio/config/mesh",
								"--healthCheckInterval=2s",
								"--healthCheckFile=/health",
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "config-volume",
									MountPath: "/etc/istio/config",
									ReadOnly:  true,
								},
								{
									Name:      "certs",
									MountPath: "/etc/istio/certs",
									ReadOnly:  true,
								},
								{
									Name:      "inject-config",
									MountPath: "/etc/istio/inject",
									ReadOnly:  true,
								},
							},
							ReadinessProbe: siProbe(),
							LivenessProbe:  siProbe(),
							Resources:      defaultResources(),
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "certs",
							VolumeSource: apiv1.VolumeSource{
								Secret: &apiv1.SecretVolumeSource{
									SecretName: fmt.Sprintf("istio.%s", siSa.Name),
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
						{
							Name: "inject-config",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: siCm.Name,
									},
									Items: []apiv1.KeyToPath{
										{
											Key:  "config",
											Path: "config",
										},
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
	controllerutil.SetControllerReference(istio, siDeploy, r.scheme)
	siResources["deployment."+siDeploy.Name] = siDeploy

	siSvc := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-sidecar-injector",
			Namespace: istio.Namespace,
			Labels: map[string]string{
				"istio": "sidecar-injector",
			},
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{Port: 443},
			},
			Selector: map[string]string{
				"istio": "sidecar-injector",
			},
		},
	}
	controllerutil.SetControllerReference(istio, siSvc, r.scheme)
	siResources["service."+siSvc.Name] = siSvc

	fail := admissionv1beta1.Fail
	siWebhook := &admissionv1beta1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-sidecar-injector",
			Namespace: istio.Namespace,
			Labels: map[string]string{
				"app": "istio-sidecar-injector",
			},
		},
		Webhooks: []admissionv1beta1.Webhook{
			{
				Name: "sidecar-injector.istio.io",
				ClientConfig: admissionv1beta1.WebhookClientConfig{
					Service: &admissionv1beta1.ServiceReference{
						Name:      "istio-sidecar-injector",
						Namespace: istio.Namespace,
						Path:      util.StrPointer("/inject"),
					},
					CABundle: []byte{},
				},
				Rules: []admissionv1beta1.RuleWithOperations{
					{
						Operations: []admissionv1beta1.OperationType{
							admissionv1beta1.Create,
						},
						Rule: admissionv1beta1.Rule{
							Resources:   []string{"pods"},
							APIGroups:   []string{""},
							APIVersions: []string{"*"},
						},
					},
				},
				FailurePolicy: &fail,
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"istio-injection": "enabled",
					},
				},
			},
		},
	}
	controllerutil.SetControllerReference(istio, siWebhook, r.scheme)
	siResources["mutatingwebhook."+siWebhook.Name] = siWebhook

	for name, res := range siResources {
		err := k8sutil.ReconcileResource(log, r.Client, istio.Namespace, strings.Split(name, ".")[1], res)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", res.GetObjectKind().GroupVersionKind().Kind, "name", name)
		}
	}

	return nil
}

func siProbe() *apiv1.Probe {
	return &apiv1.Probe{
		Handler: apiv1.Handler{
			Exec: &apiv1.ExecAction{
				Command: []string{
					"/usr/local/bin/sidecar-injector",
					"probe",
					"--probe-path=/health",
					"--interval=4s",
				},
			},
		},
		InitialDelaySeconds: 4,
		PeriodSeconds:       4,
	}
}

func siConfig() (string, error) {
	siConfig := map[string]string{
		"policy":   "enabled",
		"template": templateConfig(),
	}
	marshaledConfig, err := yaml.Marshal(siConfig)
	if err != nil {
		return "", emperror.Wrap(err, "failed to marshal sidecar injector config")
	}
	return string(marshaledConfig), nil

}

func templateConfig() string {
	return `initContainers:
- name: istio-init
  image: docker.io/istio/proxy_init:1.0.5
  args:
  - "-p"
  - [[ .MeshConfig.ProxyListenPort ]]
  - "-u"
  - 1337
  - "-m"
  - [[ annotation .ObjectMeta ` + "`" + `sidecar.istio.io/interceptionMode` + "`" + ` .ProxyConfig.InterceptionMode ]]
  - "-i"
  - "[[ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/includeOutboundIPRanges` + "`" + ` "*" ]]"
  - "-x"
  - "[[ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/excludeOutboundIPRanges` + "`" + ` "" ]]"
  - "-b"
  - "[[ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/includeInboundPorts` + "`" + ` (includeInboundPorts .Spec.Containers) ]]"
  - "-d"
  - "[[ excludeInboundPort (annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` "0" ) (annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/excludeInboundPorts` + "`" + ` "" ) ]]"
  imagePullPolicy: IfNotPresent
  securityContext:
    capabilities:
      add:
      - NET_ADMIN
    privileged: true
  restartPolicy: Always
containers:
- name: istio-proxy
  image: "[[ annotation .ObjectMeta ` + "`" + `sidecar.istio.io/proxyImage` + "`" + ` "docker.io/istio/proxyv2:1.0.5" ]]"
  ports:
  - containerPort: 15090
    protocol: TCP
    name: http-envoy-prom
  args:
  - proxy
  - sidecar
  - --configPath
  - [[ .ProxyConfig.ConfigPath ]]
  - --binaryPath
  - [[ .ProxyConfig.BinaryPath ]]
  - --serviceCluster
  [[ if ne "" (index .ObjectMeta.Labels "app") -]]
  - [[ index .ObjectMeta.Labels "app" ]]
  [[ else -]]
  - "istio-proxy"
  [[ end -]]
  - --drainDuration
  - [[ formatDuration .ProxyConfig.DrainDuration ]]
  - --parentShutdownDuration
  - [[ formatDuration .ProxyConfig.ParentShutdownDuration ]]
  - --discoveryAddress
  - [[ annotation .ObjectMeta ` + "`" + `sidecar.istio.io/discoveryAddress` + "`" + ` .ProxyConfig.DiscoveryAddress ]]
  - --discoveryRefreshDelay
  - [[ formatDuration .ProxyConfig.DiscoveryRefreshDelay ]]
  - --zipkinAddress
  - [[ .ProxyConfig.ZipkinAddress ]]
  - --connectTimeout
  - [[ formatDuration .ProxyConfig.ConnectTimeout ]]
  - --proxyAdminPort
  - [[ .ProxyConfig.ProxyAdminPort ]]
  [[ if gt .ProxyConfig.Concurrency 0 -]]
  - --concurrency
  - [[ .ProxyConfig.Concurrency ]]
  [[ end -]]
  - --controlPlaneAuthPolicy
  - [[ annotation .ObjectMeta ` + "`" + `sidecar.istio.io/controlPlaneAuthPolicy` + "`" + ` .ProxyConfig.ControlPlaneAuthPolicy ]]
[[- if (ne (annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` "0" ) "0") ]]
  - --statusPort
  - [[ annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` "0" ]]
  - --applicationPorts
  - [[ annotation .ObjectMeta ` + "`" + `readiness.status.sidecar.istio.io/applicationPorts` + "`" + ` (applicationPorts .Spec.Containers) ]]
[[- end ]]
  env:
  - name: POD_NAME
    valueFrom:
      fieldRef:
        fieldPath: metadata.name
  - name: POD_NAMESPACE
    valueFrom:
      fieldRef:
        fieldPath: metadata.namespace
  - name: INSTANCE_IP
    valueFrom:
      fieldRef:
        fieldPath: status.podIP
  - name: ISTIO_META_POD_NAME
    valueFrom:
      fieldRef:
        fieldPath: metadata.name
  - name: ISTIO_META_INTERCEPTION_MODE
    value: [[ or (index .ObjectMeta.Annotations "sidecar.istio.io/interceptionMode") .ProxyConfig.InterceptionMode.String ]]
  [[ if .ObjectMeta.Annotations ]]
  - name: ISTIO_METAJSON_ANNOTATIONS
    value: |
           [[ toJson .ObjectMeta.Annotations ]]
  [[ end ]]
  [[ if .ObjectMeta.Labels ]]
  - name: ISTIO_METAJSON_LABELS
    value: |
           [[ toJson .ObjectMeta.Labels ]]
  [[ end ]]
  imagePullPolicy: IfNotPresent
  [[ if (ne (annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` "0" ) "0") ]]
  readinessProbe:
    httpGet:
      path: /healthz/ready
      port: [[ annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` "0" ]]
    initialDelaySeconds: [[ annotation .ObjectMeta ` + "`" + `readiness.status.sidecar.istio.io/initialDelaySeconds` + "`" + ` "1" ]]
    periodSeconds: [[ annotation .ObjectMeta ` + "`" + `readiness.status.sidecar.istio.io/periodSeconds` + "`" + ` "2" ]]
    failureThreshold: [[ annotation .ObjectMeta ` + "`" + `readiness.status.sidecar.istio.io/failureThreshold` + "`" + ` "30" ]]
  [[ end -]]
  securityContext:
    readOnlyRootFilesystem: true
    [[ if eq (annotation .ObjectMeta ` + "`" + `sidecar.istio.io/interceptionMode` + "`" + ` .ProxyConfig.InterceptionMode) "TPROXY" -]]
    capabilities:
      add:
      - NET_ADMIN
    runAsGroup: 1337
    [[ else -]]
    runAsUser: 1337
    [[ end -]]
  restartPolicy: Always
  resources:
    [[ if (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPU` + "`" + `) -]]
    requests:
      cpu: [[ index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPU` + "`" + ` ]]
      memory: [[ index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyMemory` + "`" + ` ]]
  [[ else -]]
    requests:
      cpu: 10m
  [[ end -]]
  volumeMounts:
  - mountPath: /etc/istio/proxy
    name: istio-envoy
  - mountPath: /etc/certs/
    name: istio-certs
    readOnly: true
volumes:
- emptyDir:
    medium: Memory
  name: istio-envoy
- name: istio-certs
  secret:
    optional: true
    [[ if eq .Spec.ServiceAccountName "" -]]
    secretName: istio.default
    [[ else -]]
    secretName: [[ printf "istio.%s" .Spec.ServiceAccountName ]]
    [[ end -]]
`
}
