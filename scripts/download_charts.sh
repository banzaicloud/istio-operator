#!/bin/sh

set -e

DIR=$(pwd)/tmp/_output/helm

ISTIO_VERSION=1.1.0
#ISTIO_BRANCH=release-1.1

RELEASES_DIR=${DIR}/istio-releases

PLATFORM=linux
if [ -n "${ISTIO_VERSION}" ] ; then
  ISTIO_NAME=istio-${ISTIO_VERSION}
  ISTIO_FILE="${ISTIO_NAME}-${PLATFORM}.tar.gz"
  ISTIO_URL="https://github.com/istio/istio/releases/download/${ISTIO_VERSION}/${ISTIO_FILE}"
  EXTRACT_CMD="tar --strip-components=4 -C ./${ISTIO_NAME} -xvzf ${ISTIO_FILE} ${ISTIO_NAME}/install/kubernetes/helm"
  RELEASE_DIR="${RELEASES_DIR}/${ISTIO_NAME}"
else
  ISTIO_NAME=istio-${ISTIO_BRANCH}
  ISTIO_FILE="${ISTIO_BRANCH}.zip"
  ISTIO_URL="https://github.com/istio/istio/archive/${ISTIO_FILE}"
  EXTRACT_CMD="unzip ${ISTIO_FILE} ${ISTIO_NAME}/install/kubernetes/helm"
  RELEASE_DIR="${RELEASES_DIR}/${ISTIO_NAME}"
fi

ISTIO_NAME=${ISTIO_NAME//./-}

HELM_DIR=${RELEASE_DIR}

function retrieveIstioRelease() {
  if [ -d "${RELEASE_DIR}" ] ; then
    rm -rf "${RELEASE_DIR}"
  fi
  mkdir -p "${RELEASE_DIR}"

  if [ ! -f "${RELEASES_DIR}/${ISTIO_FILE}" ] ; then
    (
      echo "downloading Istio Release: ${ISTIO_URL}"
      cd "${RELEASES_DIR}"
      curl -LO "${ISTIO_URL}"
    )
  fi

  (
      echo "extracting Istio Helm charts to ${RELEASES_DIR}"
      cd "${RELEASES_DIR}"
      ${EXTRACT_CMD}
  )
}

# The following modifications are made to the generated helm template for the Istio yaml files
# - remove the create customer resources job, we handle this in the installer to deal with potential races
# - remove the cleanup secrets job, we handle this in the installer
function patchTemplates() {
  echo "patching Helm charts"
  # - remove the create customer resources job, we handle this in the installer to deal with potential races
  rm ${HELM_DIR}/istio/charts/grafana/templates/create-custom-resources-job.yaml

  # - remove the cleanup secrets job, we handle this in the installer
  rm ${HELM_DIR}/istio/charts/security/templates/cleanup-secrets.yaml

  # - we create custom resources in the normal way
  rm ${HELM_DIR}/istio/charts/security/templates/create-custom-resources-job.yaml
  rm ${HELM_DIR}/istio/charts/security/templates/configmap.yaml

  # now make sure they're available
  sed -i -e 's/define "security-default\.yaml\.tpl"/if and .Values.createMeshPolicy .Values.global.mtls.enabled/' ${HELM_DIR}/istio/charts/security/templates/enable-mesh-mtls.yaml
  sed -i -e 's/define "security-permissive\.yaml\.tpl"/if and .Values.createMeshPolicy (not .Values.global.mtls.enabled)/' ${HELM_DIR}/istio/charts/security/templates/enable-mesh-permissive.yaml

  # remove namespace from the mutatingwebhook configuration because it's cluster scoped and setting the namespace provides reconciliation errors
  sed -i -e '0,/namespace/{//d}' ${HELM_DIR}/istio/charts/sidecarInjectorWebhook/templates/mutatingwebhook.yaml
}

retrieveIstioRelease
patchTemplates
