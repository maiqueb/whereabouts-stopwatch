#!/bin/bash
MULTUS_REPO="${MULTUS_REPO:-$HOME/kubernetes/multus-cni}"
OVN_K_REPO="${OVN_K_REPO:-$HOME/kubernetes/ovn-kubernetes}"
WHEREABOUTS_REPO="${WHEREABOUTS_REPO:-$HOME/github/whereabouts}"
NAD_PATH="${NAD_PATH:-$HOME/whereabouts-scale/net-attach-def.yaml}"
REPLICASET_PATH="${REPLICASET_PATH:-$HOME/whereabouts-scale/replica-set.yaml}"

pushd $OVN_K_REPO/contrib
./kind.sh
popd

cat "$MULTUS_REPO"/images/multus-daemonset.yml | kubectl apply -f -

pushd $WHEREABOUTS_REPO
kubectl apply \
    -f ./doc/daemonset-install.yaml \
    -f ./doc/whereabouts.cni.cncf.io_ippools.yaml \
    -f ./doc/whereabouts.cni.cncf.io_overlappingrangeipreservations.yaml
popd

echo "Provisioning NAD: $NAD_PATH"
kubectl apply -f $NAD_PATH

echo "Provisioning Replica Set: $REPLICASET_PATH"
date -u +"%T.%N" && \
    kubectl apply -f ~/whereabouts-scale/replica-set.yaml && \
    while [[ $(kubectl get pods -l tier=whereabouts-test \
        -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}' | \
        tr ' ' '\n' | sort | uniq) != "True" ]]; do sleep 1; done && echo "" && date -u +"%T.%N"
