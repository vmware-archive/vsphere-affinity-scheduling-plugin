

# vsphere-affinity-scheduling-plugin

## Overview

### Kubernetes

Kubernetes is gaining popularity in container management, automating the deployment,
upgrade and scaling of containerized applications. Kube-scheduler is the component
to that decides the placement of pods onto a pool of worker machines called nodes.
Kube-scheduler has a rich set of features, which allows users to specify resource
constraint to pods as well as other policies. One of the most popular policies is
pod affinity/anti-affinity, which allows users to specify pods that reside together
or separately.

### Pod affinity/anti-affinity on vSphere

Pod affinity/anti-affinity policies are very useful when users want either pod
adjacency for performance, or separation for redundancy. For example, users want
2 pods on the same host, so the communication between those pods are not traveling
externally to network devices. A more critical use case is users want a 3-node
etcd cluster to be deployed on 3 different nodes, so losing any one of them does
not cause the cluster lose quorum. Both use cases require pod affinity and pod
anti-affinity policy support from Kube-scheduler.

Everything should work perfectly if the Kubernetes cluster runs on physical nodes.
However, when the cluster node is running on virtual machines (most likely it
does), like on vSphere, things become more complicated. Take anti-affinity as
an example, if all the etcd pods running on different virtual machines, but all
virtual machines are running on the same physical host, does the policy give you
any guarantee that you assume you would have? The answer is clearly no. This
plugin is to extend Kubernetes scheduler so that it has both additional information
from vSphere to help it making better affinity decisions.

Refer to this [intro](./docs/intro.md) for more details.

## Try it out

### Compatibility

This plugin is compatible with

- Kubernetes 1.8+
- vSphere 6.5+

Any Kubernetes distribution running on vSphere should be compatible with this
plugin, no matter it's managed by [PKS](https://pivotal.io/platform/pivotal-container-service),
[tectonic](https://coreos.com/tectonic/), or any other tools. However, if you
do find issues with compatibility, please let us know by submitting issues
[here](https://github.com/vmware/vsphere-affinity-scheduling-plugin/issues).

### Prerequisites

* [go](https://golang.org)
* [docker](https://docker.io)

### Build

Just running `go build` on the root folder of the project will compile into
`vsphere-affinity-scheduling-plugin` in the same folder.

Running `go install` will build and install it into your `GOPATH`.

### Test

`make check` will check the style by govet and golint.

`make test` will run unit tests.

`make e2e` will run e2e tests.

### Run

Check the CLI helper with `./vsphere-affinity-scheduling-plugin -h`

## Contributing

The vsphere-affinity-scheduling-plugin project team welcomes contributions from the community. If you wish to contribute code and you have not
signed our contributor license agreement (CLA), our bot will update the issue when you open a Pull Request. For any
questions about the CLA process, please refer to our [FAQ](https://cla.vmware.com/faq). For more detailed information,
refer to [CONTRIBUTING.md](CONTRIBUTING.md).

## License

This software is available under Apache 2 license.
