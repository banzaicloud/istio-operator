global:

  # ImagePullSecrets for control plane ServiceAccount, list of secrets in the same namespace
  # to use for pulling any images in pods that reference this ServiceAccount.
  # Must be set for any cluster configured with private docker registry.
  imagePullSecrets: []

  # Used to locate istiod.
  istioNamespace: {{ .Namespace }}

  istiod:
    enableAnalysis: {{ or .GetSpec.GetIstiod.GetEnableAnalysis false }}

  configValidation: true
  externalIstiod: {{ or .GetSpec.GetIstiod.GetExternalIstiod.GetEnabled false }}

  revision: "{{ .Name }}"

base:
  # Used for helm2 to add the CRDs to templates.
  enableCRDTemplates: false

  # For istioctl usage to disable istio config crds in base
  enableIstioConfigCRDs: true
