.PHONY: image
image: vsphere-affinity-scheduling-plugin
	docker build -t anfernee/vsphere-sched-plugin .

vsphere-affinity-scheduling-plugin:
	GOOS=linux go build ..
