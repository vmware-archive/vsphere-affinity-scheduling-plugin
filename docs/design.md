## Design

What we are trying to do is to build a bridge between Kubernetes scheduler from
the virtual world and the vSphere scheduler from the physical world, so both are
correctly configured with the information from the other side. The plugin has 2
components: bridger and kube-scheduler extender.

### Bridger

The bridger builds a bridge between Kubernetes scheduler and vSphere scheduler.
The information flow is bi-directional, between vSphere and Kubernetes. The
information is used in different situations, which will be described in detail
later in this section.

#### vSphere => Kubernetes
The plugin use a topology key `alpha.cna.vmware.com/host` to label the physical
host on all the nodes in a Kubernetes cluster. The plugin uses IP address to
identify node, so vmware tool is required to be installed on all Kubernetes
nodes. The bridger will automatically detect a virtual machine belongs to a
given Kubernetes cluster, and it retrieves the physical host name and labels it
as `alpha.cna.vmware.com/host=<physical_host>`. When vMotion happens, bridger
also detects the event, and modify the label accordingly. For example, a cluster
might look like this when plugin is enabled:

```console
$ kubectl get nodes -L alpha.cna.vmware.com/host
NAME                                        STATUS    AGE       VERSION    HOST
ip-10-0-14-19.us-west-1.compute.internal    Ready     49d       v1.8.9     host1
ip-10-0-28-65.us-west-1.compute.internal    Ready     49d       v1.8.9     host2
ip-10-0-40-131.us-west-1.compute.internal   Ready     49d       v1.8.9     host1
ip-10-0-51-255.us-west-1.compute.internal   Ready     49d       v1.8.9     host2

```

This means the 1st and 3rd node is running on host1 and 2nd and 4th node is
running on host2. With this information, Kube-scheduler is able to make the
correct placement, as long as the special topology key is given in the spec. You
can also use the plugin extender optionally to achieve similar goal.

#### Kubernetes => vSphere (not finished yet)
Once all the pods are placed in the right Kubernetes nodes, with all the pod
affinity and anti-affinity policy correctly satisfied, it's still not done yet.
Why? The reason is vSphere's feature called [DRS](https://www.vmware.com/products/vsphere/drs-dpm.html).
One of the things that DRS does, is *Automated Load Balancing*. Basically, when
DRS detects the virtual machine workloads across vSphere physical hosts are not
balanced, based on your DRS configuration, it will migrate (vMotion) some
virtual machines to other hosts within the same cluster to maximize overall
performance. It's a fancy feature, but when DRS doesn't know anything about the
Kubernetes world, it might work again each other, which breaks all the original
placement decisions made by Kubernetes, and make them not useful.

What bridger does is whenever a pod with affinity policy placed on a node, label
the virtual machine and physical host on vSphere with similar policy, so that
when DRS happens, it knows which physical hosts this virtual machine can/cannot
be migrated to.

### Kube-scheduler extender (optional)

Kubernetes support extending scheduler by an [http extender](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/scheduling/scheduler_extender.md).
This plugin also implements an http extender to replace the pod-affinity policy
algorithm in　kubernetes default scheduler. It's optional. You can achieve the
benefit by merely using　bridger, without the complex of configuring default
scheduler to work with this extender.　But when you choose to use it, it gives
you:

- No need to replace `kubernetes.io/hostname` to `alpha.cna.vmware.com/host` in
  pod spec.
- Better caching and faster scheduling (TODO: numbers)
