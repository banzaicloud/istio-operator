#!/usr/bin/env bash

######################################################################
##
## The script is intended to be used via `make bump-versions`
## command. Usage:
##
##   1. Update the istio version in the Makefile
##   2. Set versions of components you'd like to update in this file
##   3. Run `make bump-versions` command
##   4. Synchronize manifests with the upstream istio
##
######################################################################

set -euo pipefail

ABSOLUTE_PATH="$(greadlink -e $(dirname "$0"))/.."

BUMP_GOLANG=${BUMP_GOLANG:-true}
BUMP_ISTIO=${BUMP_ISTIO:-true}
UPDATE_ENVOY_FILTERS=${UPDATE_ENVOY_FILTERS:-true}
BUMP_DEMOAPP=${BUMP_DEMOAPP:-true}
BUMP_K8S=${BUMP_K8S:-true}
BUMP_OPERATOR=${BUMP_OPERATOR:-true}
BUMP_KUBE_RBAC_PROXY=${BUMP_KUBE_RBAC_PROXY:-true}

ISTIO_VERSION=${ISTIO_VERSION:-1.17.1}
OPERATOR_VERSION=${OPERATOR_VERSION:-2.17.0}
CHART_VERSION=${CHART_VERSION:-2.1.2}
GOLANG_VERSION=${GOLANG_VERSION:-1.18}
DEMOAPP_VERSION=${DEMOAPP_VERSION:-1.17.0}
K8S_MIN_VERSION=${K8S_MIN_VERSION:-1.23}
K8S_MAX_VERSION=${K8S_MAX_VERSION:-1.26}
KUBE_RBAC_PROXY_VERSION=${KUBE_RBAC_PROXY_VERSION:-0.8.0}

# Major istio versions (f.e., 1.16, 1.17)
ISTIO_MAJOR_VERSION=$(echo $ISTIO_VERSION | cut -d. -f1,2)
ISTIO_PREV_MAJOR_VERSION=$(echo $ISTIO_MAJOR_VERSION | awk -F'.' '{ v = $2 - 1; print $1"."v }')
ISTIO_TWO_VERSIONS_BEFORE=$(echo $ISTIO_MAJOR_VERSION | awk -F'.' '{ v = $2 - 2; print $1"."v }')

# Major istio versions without dots (f.e., 116, 117)
ISTIO_MAJOR_VERSION_WITHOUT_DOTS=$(echo $ISTIO_MAJOR_VERSION | tr -d '.')

# Major istio version regexes (f.e., 1\.16, 1\.17)
ISTIO_MAJOR_VERSION_REGEX=$(echo $ISTIO_VERSION | awk -F'.' '{ print $1"\\\\."$2 }')
ISTIO_PREV_MAJOR_VERSION_REGEX=$(echo $ISTIO_MAJOR_VERSION | awk -F'.' '{ v = $2 - 1; print $1"\\\\."v }')
ISTIO_TWO_VERSIONS_BEFORE_REGEX=$(echo $ISTIO_MAJOR_VERSION | awk -F'.' '{ v = $2 - 2; print $1"\\\\."v }')

bumpGoVersion() {
    echo "Updating Go version if needed..."

    gsed -e "s/\(^go\s\)[0-9]*\.[0-9]*$/\1$GOLANG_VERSION/" -i "$ABSOLUTE_PATH/go.mod"
    gsed -e "s/\(^go\s\)[0-9]*\.[0-9]*$/\1$GOLANG_VERSION/" -i "$ABSOLUTE_PATH/api/go.mod"
    gsed -e "s/\(^go\s\)[0-9]*\.[0-9]*$/\1$GOLANG_VERSION/" -i "$ABSOLUTE_PATH/deploy/charts/go.mod"
    gsed -e "s/\(FROM golang\:\)[0-9]*\.[0-9]*\(.*\)/\1$GOLANG_VERSION\2/" -i "$ABSOLUTE_PATH/Dockerfile"
    gsed -e "s/\(^\s*GO_VERSION:\s\)[0-9]*\.[0-9]*$/\1$GOLANG_VERSION/" -i "$ABSOLUTE_PATH/.github/workflows/ci.yaml"
}

bumpIstioVersion() {
    echo "Updating Istio version if needed..."

    # Update go.mod, scripts and *.md files
    gsed -e "s/\(^\s*istio\.io\/client-go\sv\)[0-9]*\.[0-9]*\.[0-9]*$/\1$ISTIO_VERSION/" -i "$ABSOLUTE_PATH/go.mod"
    gsed -e "s/\(^ISTIO_VERSION=\${1:-\"\)[0-9]*\.[0-9]*\.[0-9]*/\1$ISTIO_VERSION/" -i "$ABSOLUTE_PATH/scripts/label-crds.sh"
    gsed -e "s@\(https\://github\.com/banzaicloud/istio-operator/tree/release-\)[0-9]*\.[0-9]*\()\)@\1$ISTIO_MAJOR_VERSION\2@" -i "$ABSOLUTE_PATH/deploy/charts/istio-operator/README.md"
    gsed -e "s@\(icp-v\)[0-9]*\(x-sample\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/README.md"
    gsed -e "s@\(istio-operator-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/docs/multi-cluster-mesh/active-active/multi-cluster-active-active.md"
    gsed -e "s@\(istio\.io/rev=icp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/docs/multi-cluster-mesh/active-active/multi-cluster-active-active.md"
    gsed -e "s@\(istio-operator-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/docs/multi-cluster-mesh/active-passive/multi-cluster-active-passive.md"
    gsed -e "s@\(istio\.io/rev=icp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/docs/multi-cluster-mesh/active-passive/multi-cluster-active-passive.md"

    # Update *.yaml files
    gsed -e "s@\(icp-v\)[0-9]*\(x-sample\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/config/samples/servicemesh_v1alpha1_istiocontrolplane.yaml"
    gsed -z -e "s@\(spec:\n\s*version:\s\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@" -i "$ABSOLUTE_PATH/config/samples/servicemesh_v1alpha1_istiocontrolplane.yaml"
    gsed -e "s@\(gcr.io/istio-release/\w*:\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@" -i "$ABSOLUTE_PATH/config/samples/servicemesh_v1alpha1_istiocontrolplane.yaml"
    gsed -e "s@\(icp-v\)[0-9]*\(x-sample\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/config/samples/servicemesh_v1alpha1_istiomeshgateway.yaml"
    gsed -z -e "s@\(spec:\n\s*version:\s\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@" -i "$ABSOLUTE_PATH/docs/multi-cluster-mesh/active-active/active-icp-1.yaml"
    gsed -e "s@\(icp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/docs/multi-cluster-mesh/active-active/active-icp-1.yaml"
    gsed -z -e "s@\(spec:\n\s*version:\s\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@" -i "$ABSOLUTE_PATH/docs/multi-cluster-mesh/active-active/active-icp-2.yaml"
    gsed -e "s@\(icp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/docs/multi-cluster-mesh/active-active/active-icp-2.yaml"
    gsed -e "s@\(icp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/docs/multi-cluster-mesh/active-active/demoapp-1.yaml"
    gsed -z -e "s@\(spec:\n\s*version:\s\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@" -i "$ABSOLUTE_PATH/docs/multi-cluster-mesh/active-passive/active-icp.yaml"
    gsed -e "s@\(icp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/docs/multi-cluster-mesh/active-passive/active-icp.yaml"
    gsed -z -e "s@\(spec:\n\s*version:\s\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@" -i "$ABSOLUTE_PATH/docs/multi-cluster-mesh/active-passive/passive-icp.yaml"
    gsed -e "s@\(icp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/docs/multi-cluster-mesh/active-passive/passive-icp.yaml"
    gsed -e "s@\(icp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/docs/multi-cluster-mesh/active-passive/demoapp-1.yaml"
    gsed -e "/global:/{:a; \$!N; /\nbase:/!ba; s@\(\s*tag:\s\"\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@}" -i "$ABSOLUTE_PATH/internal/assets/manifests/istio-discovery/values.yaml"
    gsed -e "s@\(\s\{2\}version:\s\"\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@" -i "$ABSOLUTE_PATH/internal/components/base/testdata/icp-test-cr.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/base/testdata/icp-test-cr.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/base/testdata/icp-expected-values.yaml"
    gsed -e "s@\(\s\{2\}version:\s\"\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@" -i "$ABSOLUTE_PATH/internal/components/cni/testdata/icp-test-cr.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/cni/testdata/icp-test-cr.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/cni/testdata/icp-expected-resource-dump.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/cni/testdata/icp-expected-values.yaml"
    gsed -e "s@\(\s\{2\}version:\s\"\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@" -i "$ABSOLUTE_PATH/internal/components/discovery/testdata/icp-passive-expected-resource-dump.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/discovery/testdata/icp-passive-expected-resource-dump.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/discovery/testdata/icp-passive-expected-values.yaml"
    gsed -e "s@\(\s\{2\}version:\s\"\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@" -i "$ABSOLUTE_PATH/internal/components/discovery/testdata/icp-passive-test-cr.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/discovery/testdata/icp-passive-test-cr.yaml"
    gsed -e "s@\(\s\{2\}version:\s\"\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@" -i "$ABSOLUTE_PATH/internal/components/discovery/testdata/icp-test-cr.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/discovery/testdata/icp-test-cr.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/discovery/testdata/icp-expected-resource-dump.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/discovery/testdata/icp-expected-values.yaml"
    gsed -e "s@\(\s\{2\}version:\s\"\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@" -i "$ABSOLUTE_PATH/internal/components/istiomeshgateway/testdata/icp-test-cr.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/istiomeshgateway/testdata/icp-test-cr.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/istiomeshgateway/testdata/imgw-expected-resource-dump.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/istiomeshgateway/testdata/imgw-test-cr.yaml"
    gsed -e "s@\(\s\{2\}version:\s\"\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@" -i "$ABSOLUTE_PATH/internal/components/meshexpansion/testdata/icp-test-cr.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/meshexpansion/testdata/icp-test-cr.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/meshexpansion/testdata/mex-expected-resource-dump.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/meshexpansion/testdata/mex-expected-values.yaml"
    gsed -e "s@\(\s\{2\}version:\s\"\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@" -i "$ABSOLUTE_PATH/internal/components/resourcesyncrule/testdata/icp-active-test-cr.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/resourcesyncrule/testdata/icp-active-test-cr.yaml"
    gsed -e "s@\(\s\{2\}version:\s\"\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@" -i "$ABSOLUTE_PATH/internal/components/resourcesyncrule/testdata/icp-passive-test-cr.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/resourcesyncrule/testdata/icp-passive-test-cr.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/resourcesyncrule/testdata/rsr-expected-active-resource-dump.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/resourcesyncrule/testdata/rsr-expected-active-values.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/resourcesyncrule/testdata/rsr-expected-passive-resource-dump.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/resourcesyncrule/testdata/rsr-expected-passive-values.yaml"
    gsed -e "s@\(\s\{2\}version:\s\"\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@" -i "$ABSOLUTE_PATH/internal/components/sidecarinjector/testdata/icp-test-cr.yaml"
    gsed -e "s@\(banzaicloud/istio-sidecar-injector:v\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@" -i "$ABSOLUTE_PATH/internal/components/sidecarinjector/testdata/icp-test-cr.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/sidecarinjector/testdata/icp-test-cr.yaml"
    gsed -e "s@\(banzaicloud/istio-sidecar-injector:v\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@" -i "$ABSOLUTE_PATH/internal/components/sidecarinjector/testdata/icp-expected-resource-dump.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/sidecarinjector/testdata/icp-expected-resource-dump.yaml"
    gsed -e "s@\(banzaicloud/istio-sidecar-injector:v\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@" -i "$ABSOLUTE_PATH/internal/components/sidecarinjector/testdata/icp-expected-values.yaml"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/sidecarinjector/testdata/icp-expected-values.yaml"
    gsed -e "/\s\{2\}values:\s|-/{:a; \$!N; /\n\s\{6\}},\$/!ba; s@\(\s\{8\}\"tag\": \"\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@}" -i "$ABSOLUTE_PATH/internal/components/discovery/testdata/icp-passive-expected-resource-dump.yaml"

    # Update *.go files
    gsed -e "s@\(should\saccept\sall\s\)[0-9]*\.[0-9]*\(\sversions\)@\1$ISTIO_MAJOR_VERSION\2@" -i "$ABSOLUTE_PATH/controllers/version_test.go"
    gsed -e "s@\(controllers\.IsIstioVersionSupported(\"\)[0-9]*\.[0-9]*\([-\"]\)@\1$ISTIO_MAJOR_VERSION\2@" -i "$ABSOLUTE_PATH/controllers/version_test.go"
    gsed -e "s@\(controllers\.IsIstioVersionSupported(\"\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@" -i "$ABSOLUTE_PATH/controllers/version_test.go"
    ISTIO_REGEX=$(echo $ISTIO_MAJOR_VERSION | awk -F'.' '{ print $1"\\\\\\\\\."$2 }') && gsed -e "s@\(supportedIstioMinorVersionRegex\s=\s\"\^\)[0-9]*\\\\\\\.[0-9]*\((\)@\1$ISTIO_REGEX\2@" -i "$ABSOLUTE_PATH/controllers/version.go"
    gsed -e "/func TestICPBaseResourceDump.*{/{:a; \$!N; /\n}\$/!ba; s@\(logger.NewWithLogrLogger[^\n]*\n\t*\"\)[0-9]*\.[0-9]*\.[0-9]*@\1$ISTIO_VERSION@}" -i "$ABSOLUTE_PATH/internal/components/base/base_test.go"
    gsed -e "s@\(cp-v\)[0-9]*\(x\)@\1$ISTIO_MAJOR_VERSION_WITHOUT_DOTS\2@g" -i "$ABSOLUTE_PATH/internal/components/istiomeshgateway/istiomeshgateway_test.go"
}

updateEnvoyFilters() {
    echo "Updating envoy filters and creating istiod telemetry file if needed..."
    TELEMETRY_OLD_FILE_PATH="$ABSOLUTE_PATH/internal/assets/manifests/istio-discovery/templates/telemetryv2_$ISTIO_TWO_VERSIONS_BEFORE.yaml"
    TELEMETRY_NEW_FILE_PATH="$ABSOLUTE_PATH/internal/assets/manifests/istio-discovery/templates/telemetryv2_$ISTIO_MAJOR_VERSION.yaml"

    if [ -f "$TELEMETRY_OLD_FILE_PATH" ]; then
        # Update telemetryv2_*.yaml files
        gsed -e "s@\(stats-filter-\)[0-9]*\.[0-9]*@\1$ISTIO_MAJOR_VERSION@" -i "$TELEMETRY_OLD_FILE_PATH"
        gsed -e "s@\(stackdriver-filter-\)[0-9]*\.[0-9]*@\1$ISTIO_MAJOR_VERSION@" -i "$TELEMETRY_OLD_FILE_PATH"
        gsed -e "s@\(stackdriver-sampling-accesslog-filter-\)[0-9]*\.[0-9]*@\1$ISTIO_MAJOR_VERSION@" -i "$TELEMETRY_OLD_FILE_PATH"
        gsed -E -e "s@(proxyVersion:\s'\^?)[0-9]+\\\.[0-9]+(\.\*')@\1$ISTIO_MAJOR_VERSION_REGEX\2@" -i "$TELEMETRY_OLD_FILE_PATH"
        mv "$TELEMETRY_OLD_FILE_PATH" "$TELEMETRY_NEW_FILE_PATH"

        # Update gen-istio.yaml
        gsed -e "s@\(templates/telemetryv2_\)$ISTIO_PREV_MAJOR_VERSION@\1$ISTIO_MAJOR_VERSION@" -i "$ABSOLUTE_PATH/internal/assets/manifests/istio-discovery/resources/gen-istio.yaml"
        gsed -e "s@\(stats-filter-\)$ISTIO_PREV_MAJOR_VERSION@\1$ISTIO_MAJOR_VERSION@" -i "$ABSOLUTE_PATH/internal/assets/manifests/istio-discovery/resources/gen-istio.yaml"
        gsed -E -e "s@(proxyVersion:\s'\^?)$ISTIO_PREV_MAJOR_VERSION_REGEX(\.\*')@\1$ISTIO_MAJOR_VERSION_REGEX\2@" -i "$ABSOLUTE_PATH/internal/assets/manifests/istio-discovery/resources/gen-istio.yaml"
        gsed -e "s@\(templates/telemetryv2_\)$ISTIO_TWO_VERSIONS_BEFORE@\1$ISTIO_PREV_MAJOR_VERSION@" -i "$ABSOLUTE_PATH/internal/assets/manifests/istio-discovery/resources/gen-istio.yaml"
        gsed -e "s@\(stats-filter-\)$ISTIO_TWO_VERSIONS_BEFORE@\1$ISTIO_PREV_MAJOR_VERSION@" -i "$ABSOLUTE_PATH/internal/assets/manifests/istio-discovery/resources/gen-istio.yaml"
        gsed -E -e "s@(proxyVersion:\s'\^?)$ISTIO_TWO_VERSIONS_BEFORE_REGEX(\.\*')@\1$ISTIO_PREV_MAJOR_VERSION_REGEX\2@" -i "$ABSOLUTE_PATH/internal/assets/manifests/istio-discovery/resources/gen-istio.yaml"

        # Update discovery/testdata/icp-expected-resource-dump.yaml
        gsed -e "s@\(stats-filter-\)$ISTIO_PREV_MAJOR_VERSION@\1$ISTIO_MAJOR_VERSION@" -i "$ABSOLUTE_PATH/internal/components/discovery/testdata/icp-expected-resource-dump.yaml"
        gsed -E -e "s@(proxyVersion:\s\^?)$ISTIO_PREV_MAJOR_VERSION_REGEX(\.\*)@\1$ISTIO_MAJOR_VERSION_REGEX\2@" -i "$ABSOLUTE_PATH/internal/components/discovery/testdata/icp-expected-resource-dump.yaml"
        gsed -e "s@\(stats-filter-\)$ISTIO_TWO_VERSIONS_BEFORE@\1$ISTIO_PREV_MAJOR_VERSION@" -i "$ABSOLUTE_PATH/internal/components/discovery/testdata/icp-expected-resource-dump.yaml"
        gsed -E -e "s@(proxyVersion:\s\^?)$ISTIO_TWO_VERSIONS_BEFORE_REGEX(\.\*)@\1$ISTIO_PREV_MAJOR_VERSION_REGEX\2@" -i "$ABSOLUTE_PATH/internal/components/discovery/testdata/icp-expected-resource-dump.yaml"
    fi
}

bumpDemoappVersion() {
    echo "Updating demoapp version if needed..."

    gsed -e "s@\(docker\.io/istio/.*:\)[0-9]*\.[0-9]*\.[0-9]*@\1$DEMOAPP_VERSION@" -i "$ABSOLUTE_PATH/docs/multi-cluster-mesh/active-active/demoapp-1.yaml"
    gsed -e "s@\(docker\.io/istio/.*:\)[0-9]*\.[0-9]*\.[0-9]*@\1$DEMOAPP_VERSION@" -i "$ABSOLUTE_PATH/docs/multi-cluster-mesh/active-active/demoapp-2.yaml"
    gsed -e "s@\(docker\.io/istio/.*:\)[0-9]*\.[0-9]*\.[0-9]*@\1$DEMOAPP_VERSION@" -i "$ABSOLUTE_PATH/docs/multi-cluster-mesh/active-passive/demoapp-1.yaml"
    gsed -e "s@\(docker\.io/istio/.*:\)[0-9]*\.[0-9]*\.[0-9]*@\1$DEMOAPP_VERSION@" -i "$ABSOLUTE_PATH/docs/multi-cluster-mesh/active-passive/demoapp-2.yaml"
}

bumpK8sVersion() {
    echo "Updating k8s versions if needed..."

    gsed -e "s@\(kubernetes\scluster\s(version\s\)[0-9]*\.[0-9]*\(+)\)@\1$K8S_MIN_VERSION\2@" -i "$ABSOLUTE_PATH/README.md"
    gsed -e "s@\(-\sKubernetes\s\)[0-9]*\.[0-9]*\(\.0\s-\s\)[0-9]*\.[0-9]*\(\.x\)@\1$K8S_MIN_VERSION\2$K8S_MAX_VERSION\3@" -i "$ABSOLUTE_PATH/deploy/charts/istio-operator/README.md"
    K8S_NEXT=$(echo $K8S_MAX_VERSION | awk -F'.' '{ v = $2 + 1; print $1"."v }') && gsed -e "s@\(kubeVersion:\s\">=\s\)[0-9]*\.[0-9]*\(\.0-0\s<\s\)[0-9]*\.[0-9]*\(\.0-0\"\)@\1$K8S_MIN_VERSION\2$K8S_NEXT\3@" -i "$ABSOLUTE_PATH/deploy/charts/istio-operator/Chart.yaml"
}

bumpOperatorVersion() {
    echo "Updating the operator and chart versions if needed..."

    gsed -e "s@\(Operator\scontainer\simage\stag\s*|\s\`v\)[0-9]*\.[0-9]*\.[0-9]*\(\`\)@\1$OPERATOR_VERSION\2@" -i "$ABSOLUTE_PATH/deploy/charts/istio-operator/README.md"
    gsed -e "s@\(appVersion:\s\"v\)[0-9]*\.[0-9]*\.[0-9]*\(\"\)@\1$OPERATOR_VERSION\2@" -i "$ABSOLUTE_PATH/deploy/charts/istio-operator/Chart.yaml"
    gsed -e "s@\(^\s\{2\}tag:\s\"v\)[0-9]*\.[0-9]*\.[0-9]*\(\"\)@\1$OPERATOR_VERSION\2@" -i "$ABSOLUTE_PATH/deploy/charts/istio-operator/values.yaml"
    gsed -e "s@\(^version:\s\)[0-9]*\.[0-9]*\.[0-9]*@\1$CHART_VERSION@" -i "$ABSOLUTE_PATH/deploy/charts/istio-operator/Chart.yaml"
}

bumpKubeRbacProxyVersion() {
    echo "Updating kube-rbac-proxy version if needed..."

    gsed -e "s@\(gcr\.io/kubebuilder/kube-rbac-proxy:v\)[0-9]*\.[0-9]*\.[0-9]*@\1$KUBE_RBAC_PROXY_VERSION@" -i "$ABSOLUTE_PATH/config/default/manager_auth_proxy_patch.yaml"
    gsed -e "s@\(^\s\{6\}tag:\s\"v\)[0-9]*\.[0-9]*\.[0-9]*\(\"\)@\1$KUBE_RBAC_PROXY_VERSION\2@" -i "$ABSOLUTE_PATH/deploy/charts/istio-operator/values.yaml"
    gsed -e "s@\(Auth\sproxy\scontainer\simage\stag\s*|\s\`v\)[0-9]*\.[0-9]*\.[0-9]*\(\`\)@\1$KUBE_RBAC_PROXY_VERSION\2@" -i "$ABSOLUTE_PATH/deploy/charts/istio-operator/README.md"
}

main() {
    if [ $BUMP_GOLANG = "true" ]; then
        bumpGoVersion
    fi

    if [ $BUMP_ISTIO = "true" ]; then
        bumpIstioVersion
    fi

    if [ $UPDATE_ENVOY_FILTERS = "true" ]; then
        updateEnvoyFilters
    fi

    if [ $BUMP_DEMOAPP = "true" ]; then
        bumpDemoappVersion
    fi

    if [ $BUMP_K8S = "true" ]; then
        bumpK8sVersion
    fi

    if [ $BUMP_OPERATOR = "true" ]; then
        bumpOperatorVersion
    fi

    if [ $BUMP_KUBE_RBAC_PROXY = "true" ]; then
        bumpKubeRbacProxyVersion
    fi
}

main
