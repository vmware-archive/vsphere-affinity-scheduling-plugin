/*
Copyright (c) 201８ VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/constants"
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/k8s/client"
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/k8s/nodeupdater"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	schedv1 "k8s.io/kubernetes/pkg/scheduler/api/v1"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var (
	// config is the test configuration
	config Config

	// ctx is the test context
	ctx context
)

var _ = BeforeSuite(func() {
	// Config parse
	config.Parse()
	log.Printf("Using config Config%+v", config)

	// Init ctx.cmd
	ctx.cmd = exec.Command(config.pluginPath, "-url", config.govmomiURL)

	file, err := os.OpenFile(config.pluginLogFile, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}

	ctx.cmd.Stdout = file
	ctx.cmd.Stderr = file

	// Init ctx.k8sClient
	ctx.k8sClient, err = client.New()
	if err != nil {
		log.Fatal(err)
	}

	// Start the plugin
	go func() {
		err := ctx.cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()
})

var _ = AfterSuite(func() {
	// Stop the plugin
	ctx.cmd.Process.Kill()
	ctx.cmd.Wait()
})

var _ = Describe("e2e test", func() {
	BeforeEach(func() {
		log.Println("before e2e testcase")
	})

	AfterEach(func() {
		log.Println("after e2e testcase")
	})

	Context("node labeller", func() {
		It("should label the node with host correctly", func() {
			// Skip("xxx")

			// clear the label for all nodes
			log.Println("clearing labels for all nodes")
			nodeList := ctx.GetNodes()
			nodeUpdater := nodeupdater.New(ctx.k8sClient)

			for _, node := range nodeList {
				log.Printf("clearing label for node %s", node.Name)
				err := nodeUpdater.DeleteLabel(node.Name)
				if err != nil {
					log.Fatalf("failed to delete label for %s: %s", node.Name, err)
				}
			}

			// TODO: >10 second interval; remove this.
			time.Sleep(15 * time.Second)

			for _, node := range nodeList {
				_, ok := node.GetLabels()[constants.HostLabel]
				if !ok {
					Fail("expect HostLabel to be set")
				}
			}
		})
	})

	Context("scheduler http extender", func() {
		It("should filter the nodes based on affinity rule", func() {
			// Get nodenames
			nodes := ctx.GetNodes()
			nodeNames := ctx.GetNodeNames()

			// Select a node that runs on the same physical host
			nodeToHost := make(map[string]string)
			hostToNode := make(map[string][]string)

			for _, node := range nodes {
				hostName, ok := node.GetLabels()[constants.HostLabel]
				if ok {
					log.Printf("%s<=>%s", node.Name, hostName)
					nodeToHost[node.Name] = hostName
					hostToNode[hostName] = append(hostToNode[hostName], node.Name)
				}
			}

			var theNodes []string
			for _, nodes := range hostToNode {
				if len(nodes) > 1 {
					theNodes = nodes
				}
			}

			if len(theNodes) == 0 {
				Skip("Skip because there are no 2 nodes on the same physical host")
			}

			// Deploy an anchor pod
			anchorName := fmt.Sprintf("e2e-anchor-pod-%d-%d", time.Now().Unix(), rand.Intn(1000))
			anchorPod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      anchorName,
					Namespace: "default",
					Labels: map[string]string{
						"vsphere-plugin-e2e": anchorName,
					},
				},
				Spec: v1.PodSpec{
					NodeName: theNodes[0],
					Containers: []v1.Container{
						v1.Container{
							Image: "nginx",
							Name:  "nginx",
						},
					},
				},
			}

			_ = ctx.k8sClient.CoreV1().Pods("default").Delete(anchorPod.Name, nil)
			_, err := ctx.k8sClient.CoreV1().Pods("default").Create(anchorPod)
			Expect(err).NotTo(HaveOccurred())

			defer ctx.k8sClient.CoreV1().Pods("default").Delete(anchorPod.Name, nil)

			// Check Affinity rules
			for i := 0; i < 10; i++ {
				args := &schedv1.ExtenderArgs{
					Pod: v1.Pod{
						Spec: v1.PodSpec{
							Affinity: &v1.Affinity{
								PodAffinity: &v1.PodAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
										{
											LabelSelector: &metav1.LabelSelector{
												MatchLabels: map[string]string{
													"vsphere-plugin-e2e": anchorName,
												},
											},
											Namespaces:  []string{},
											TopologyKey: "alpha.cna.vmware.com/host",
										},
									},
								},
							},
						},
					},
					NodeNames: &nodeNames,
				}

				resp, err := postFilterRequest("http://localhost:12346/scheduler/filter", args)
				Expect(err).NotTo(HaveOccurred())
				Expect(*resp.NodeNames).To(ConsistOf(theNodes))
			}

			// Check anti-affinity rules
			for i := 0; i < 10; i++ {
				args := &schedv1.ExtenderArgs{
					Pod: v1.Pod{
						Spec: v1.PodSpec{
							Affinity: &v1.Affinity{
								PodAntiAffinity: &v1.PodAntiAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
										{
											LabelSelector: &metav1.LabelSelector{
												MatchLabels: map[string]string{
													"vsphere-plugin-e2e": anchorName,
												},
											},
											Namespaces:  []string{},
											TopologyKey: "alpha.cna.vmware.com/host",
										},
									},
								},
							},
						},
					},
					NodeNames: &nodeNames,
				}

				resp, err := postFilterRequest("http://localhost:12346/scheduler/filter", args)
				Expect(err).NotTo(HaveOccurred())
				Expect(*resp.NodeNames).NotTo(ConsistOf(theNodes))
			}
		})

		It("should filter the nodes based on anti-affinity rule", func() {
		})
	})
})
