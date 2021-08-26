# whereabouts-stopwatch
Templating engine + script to see how Whereabouts reacts to massive pod creation

## Dependencies
- make
- a working kubernetes cluster, with multus and whereabouts deployed in it
  - the macvlan cni plugin must be available on the cluster nodes
- properly set up kubeconfig, pointing at your cluster

## How to use
First build the project:
```bash
make build
```

Then template the network attachment definition / replica-set:
```bash
$ build/bin/templating-device --ipam-range 10.10.0.0/16 \
    --number-of-replicas 500 \
    --input-dir templates \
    --output-dir deployment \
    --retry-period 15
```

Finally, execute the script:
```bash
./whereabouts-stopwatch.sh whereabouts-scale-test \
    deployment/net-attach-def.yaml \
    deployment/replica-set.yaml
Provisioning NAD: deployment/net-attach-def.yaml
networkattachmentdefinition.k8s.cni.cncf.io/network1 created
Provisioning Replica Set: deployment/replica-set.yaml
Start: 11:17:48.269618929
replicaset.apps/whereabouts-scale-test created
End: 11:27:47.597009203
```

**Note:** the above figures come from a locally deployed Kubernetes, with 3 nodes (just VMs on a laptop).

## Templating tool examples
```bash
$ build/bin/templating-device --help
Usage of build/bin/templating-device:
      --app-name string          The name of the scale checking application (default "whereabouts-scale-test")
      --dump-stdout              Also print the templates to stdout
      --help                     Print help and quit (default false)
      --input-dir string         Directory with the templates (default "")
      --ipam-range string        The CIDR range to assign addresses from (default "")
      --lease-duration int       How long is an active lease maintained (default 1500)
      --lower-device string      The name of the lower device on which the macvlan interfaces will be created (default "eth0")
      --namespace string         The namespace to use (default "default")
      --network-name string      The name of the network for which whereabouts will provide IPAM (default "network1")
      --number-of-replicas int   How many replicas in the replica-set (default 100)
      --output-dir string        Output file dir (default "")
      --renew-deadline int       Time after which the lease is forcefully re-acquired (default 1000)
      --retry-period int         Period upon which the acquiring the lease is retried (default 500)

$ build/bin/templating-device --ipam-range 10.10.0.0/16 \
    --number-of-replicas 500 \
    --input-dir templates \
    --output-dir deployment

$ build/bin/templating-device --ipam-range 10.10.0.0/16 \
    --number-of-replicas 500 \
    --input-dir templates \
    --output-dir deployment \
    --retry-period 15
```
## Running the tests
Tests can be ran by:
```bash
$ make test
```

