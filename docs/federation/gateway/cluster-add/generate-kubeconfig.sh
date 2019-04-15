#!/bin/bash
if [ -z "$1" ]; then
    echo "Usage: $0 <k8s context name>"
    exit
fi

if ! kubectl config get-contexts -o name $1 >/dev/null; then
    exit
fi

CONTEXT=$1

export WORK_DIR=$(pwd)
CLUSTER_NAME=$(kubectl --context ${CONTEXT} config view --minify=true -o "jsonpath={.clusters[].name}")
export KUBECFG_FILE=${WORK_DIR}/${CLUSTER_NAME}
SERVER=$(kubectl --context ${CONTEXT} config view --minify=true -o "jsonpath={.clusters[].cluster.server}")
NAMESPACE=istio-system
SERVICE_ACCOUNT=istio-operator
SECRET_NAME=$(kubectl --context ${CONTEXT} get sa ${SERVICE_ACCOUNT} -n ${NAMESPACE} -o jsonpath='{.secrets[].name}')
CA_DATA=$(kubectl --context ${CONTEXT} get secret ${SECRET_NAME} -n ${NAMESPACE} -o "jsonpath={.data['ca\.crt']}")
TOKEN=$(kubectl --context ${CONTEXT} get secret ${SECRET_NAME} -n ${NAMESPACE} -o "jsonpath={.data['token']}" | base64 --decode)

cat <<EOF > ${KUBECFG_FILE}
apiVersion: v1
clusters:
   - cluster:
       certificate-authority-data: ${CA_DATA}
       server: ${SERVER}
     name: ${CLUSTER_NAME}
contexts:
   - context:
       cluster: ${CLUSTER_NAME}
       user: ${CLUSTER_NAME}
     name: ${CLUSTER_NAME}
current-context: ${CLUSTER_NAME}
kind: Config
preferences: {}
users:
   - name: ${CLUSTER_NAME}
     user:
       token: ${TOKEN}
EOF

echo -n "${CLUSTER_NAME}"
