package main

func ExampleTemplator() {
	templator, err := newTemplatingDevice(
		"../templates",
		"myapp",
		"network1",
		"10.10.0.0/16",
		1500,
		"eth0",
		"default",
		1000,
		1200,
		15,
		true)
	if err != nil {
		panic("boom")
	}

	if err := templator.spit("../deployment"); err != nil {
		panic("boom")
	}
	// Output: apiVersion: "k8s.cni.cncf.io/v1"
	// kind: NetworkAttachmentDefinition
	// metadata:
	//   name: network1
	//   namespace: default
	// spec:
	//   config: '{
	//       "cniVersion": "0.3.0",
	//       "type": "macvlan",
	//       "master": "eth0",
	//       "mode": "bridge",
	//       "ipam": {
	//         "type": "whereabouts",
	//         "leader_lease_duration": 1500,
	//         "leader_renew_deadline": 1000,
	//         "leader_retry_period": 15,
	//         "range": "10.10.0.0/16"
	//       }
	//     }'
	// apiVersion: apps/v1
	// kind: ReplicaSet
	// metadata:
	//   name: myapp
	//   namespace: default
	//   labels:
	//     tier: myapp
	// spec:
	//   replicas: 1200
	//   selector:
	//     matchLabels:
	//       tier: myapp
	//   template:
	//     metadata:
	//       labels:
	//         tier: myapp
	//       annotations:
	//         k8s.v1.cni.cncf.io/networks: network1
	//       namespace: default
	//     spec:
	//       containers:
	//       - name: samplepod
	//         command: ["/bin/ash", "-c", "trap : TERM INT; sleep infinity & wait"]
	//         image: quay.io/dougbtv/alpine:latest
}
