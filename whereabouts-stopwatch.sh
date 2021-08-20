#!/bin/bash -e

function print_help {
    echo "Usage:"
    echo "$0 <appName> <path to NAD spec> <path to replicaset spec>"
}

APP_NAME="$1"
NAD_PATH="$2"
REPLICASET_PATH="$3"

if [ $# != 3 ]; then
    print_help
    exit 1
fi

echo "Provisioning NAD: $NAD_PATH"
kubectl apply -f "$NAD_PATH"

echo "Provisioning Replica Set: $REPLICASET_PATH"
echo "Start: $(date -u +"%T.%N")" && \
    kubectl apply -f "$REPLICASET_PATH" && \
    while [[ $(kubectl get pods -l tier="$APP_NAME" \
        -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}' | \
        tr ' ' '\n' | sort | uniq) != "True" ]]; do sleep 0.1; done && \
    echo "End: $(date -u +"%T.%N")"
