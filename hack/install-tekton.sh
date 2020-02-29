#!/bin/bash
#
# Intall Tekton in the "kind" cluster
#

set -eu

TEKTON_VERSION="${TEKTON_VERSION:-v0.10.1}"

TEKTON_HOST="storage.googleapis.com"
TEKTON_HOST_PATH="tekton-releases/pipeline/previous"

echo "# Deploying Tekton Pipelines Operator '${TEKTON_VERSION}"

kubectl apply \
    --filename="https://${TEKTON_HOST}/${TEKTON_HOST_PATH}/${TEKTON_VERSION}/release.yaml" \
    --output="yaml"
