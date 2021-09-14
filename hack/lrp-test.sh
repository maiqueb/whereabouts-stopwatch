#!/bin/bash

# Leader retry period (LRP) test
# For each leader retry period value create a number of pods based on the number of nodes (N) and a capacity limit per node (C).
# Required input args:
#   -c - pod capacity per compute node
#   -n - number of compute nodes
#   -p - leader retry period in miliseconds - included default value of 500 to give baseline
#   -d - bridge device name

set -o errexit
set -o pipefail
set -o xtrace

while true; do
  case "$1" in
    -c|--capacity)
      capacity="$2"
      shift
      shift
      ;;
    -n|--node-count)
      node_count="$2"
      shift
      shift
      ;;
    -p|--leader-retry-period)
      leader_retry_period="$2"
      shift
      shift
      ;;
    -d|--device)
      host_device_name_for_bridge="$2"
      shift
      shift
      ;;
    *)
    break
    ;;
  esac
done

[ -z "$capacity" ] || [ -z "$node_count" ] || [ -z "$leader_retry_period" ] || [ -z "$host_device_name_for_bridge" ] && \
  echo "all flags not set. Set node capacity -c and total number of compute node -n, leader retry period (list) -p and interface device name -d" && \
  exit 1

FILE_PATH=$(dirname "$(readlink --canonicalize "${BASH_SOURCE[0]}")")
ROOT=$(readlink --canonicalize "$FILE_PATH/..")
TEMPLATING_BIN_PATH="$ROOT/build/bin/templating-device"
TEMPLATES_INPUT_DIR_PATH="$ROOT/templates"
TEMPLATES_OUTPUT_DIR_PATH="$ROOT/deployment"

# initial condition checks
[[ ! -f "$TEMPLATING_BIN_PATH" ]] && echo "have you built the project? Templating binary not found" && exit 2
[[ ! $(type -P kubectl) ]] && echo "you need kubectl in PATH" && exit 3

leader_retry_period=($(echo "$leader_retry_period" | tr ',' '\n'))
# total number of pods to be created in a single replica set
pods_creation_total=($(echo 0.2*$capacity*$node_count/1 | bc) $(echo 0.4*$capacity*$node_count/1 | bc) $(echo 0.8*$capacity*$node_count/1 | bc) $(echo 1*$capacity*$node_count/1 | bc))
rm -rf --preserve-root "$TEMPLATES_OUTPUT_DIR_PATH"
mkdir -p "$TEMPLATES_OUTPUT_DIR_PATH"

for lrp in ${leader_retry_period[@]}; do
  for number_of_pods in ${pods_creation_total[@]}; do
    echo -e "test time $(date):\nLeader retry period: $lrp\nNumber of pods in replica: $number_of_pods"

    # Generate test artifacts
    "$ROOT"/build/bin/templating-device \
      --ipam-range 10.10.0.0/16 \
      --number-of-replicas "$number_of_pods" \
      --input-dir "$TEMPLATES_INPUT_DIR_PATH" \
      --output-dir "$TEMPLATES_OUTPUT_DIR_PATH" \
      --retry-period "$lrp" \
      --lower-device "$host_device_name_for_bridge"

    # Start test
    time "$ROOT"/whereabouts-stopwatch.sh whereabouts-scale-test \
      "$TEMPLATES_OUTPUT_DIR_PATH"/net-attach-def.yaml \
      "$TEMPLATES_OUTPUT_DIR_PATH"/replica-set.yaml

    number_of_uniq_ip="$(kubectl get pods --output=jsonpath='{range .items[*].status}{.podIP}{"\n"}{end}' | uniq | wc -l)"
    if [[ $number_of_pods != "$number_of_uniq_ip" ]]; then
      echo "number of pods '$number_of_pods' did not match number of unique IP(s) '$number_of_uniq_ip'"
      echo -e "IP(s) seen:\n$(kubectl get pods --output=jsonpath='{range .items[*].status}{.podIP}{"\n"}{end}')"
    fi

    # Delete replicate set
    kubectl delete -f "$TEMPLATES_OUTPUT_DIR_PATH/net-attach-def.yaml" -f "$TEMPLATES_OUTPUT_DIR_PATH/replica-set.yaml"

    echo "timing for how long pods to delete"
    time kubectl wait --for=delete pod --selector tier=whereabouts-scale-test --timeout=120m

    echo -e "done\n\n"
    sleep 120
  done
done
