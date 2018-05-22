.PHONY: check
check:
	golint -set_exit_status pkg/...
	golint -set_exit_status main.go
	( find pkg/ -name '*.go' ; echo *.go ) | xargs go tool vet

.PHONY: test
test:
	cd pkg; go test -v -race ./...

.PHONY: e2e
e2e: vsphere-affinity-scheduling-plugin
	cd test/e2e; TARGET=../../vsphere-affinity-scheduling-plugin go test

vsphere-affinity-scheduling-plugin:
	go build
