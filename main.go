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

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/algorithm"
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/algorithm/filters"
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/bridgecache"
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/k8s/cache"
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/k8s/client"
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/server"
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/services"
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/vsphere"
	"k8s.io/apimachinery/pkg/util/wait"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

// Config is the configuration options for the scheduler extender
type Config struct {
	// Debug mode
	Debug bool

	// Port is the port to listen
	Port int

	// ClusterName is the name of the cluster where all the affinity rules are set
	ClusterName string
}

var config Config

func init() {
	flag.IntVar(&config.Port, "port", 12346, "the port the extender listens on")
	flag.BoolVar(&config.Debug, "debug", false, "debug mode")
	flag.StringVar(&config.ClusterName, "cluster", "cluster1",
		"vSphere cluster name to setup affinity/anti-affinity rules")

	flag.Parse()

	log.Printf("config: %+v", config)
}

func main() {
	// Init client
	k8sClient, err := client.New()
	if err != nil {
		panic(err)
	}

	// Init k8scache
	cache := k8scache.New(k8sClient)

	// Init vsphere Client
	vsclient := vsphere.NewCachedClient(config.ClusterName)
	defer vsclient.Logout()

	// Init bcache
	bcache := bridgecache.NewCache(cache.NodeInformer(), vsclient)

	// Setup Filters
	var filter algorithm.Filters
	filter = append(filter, filters.NewPodAffinity(cache, cache))
	filter = append(filter, filters.NewPodAntiAffinity(cache, cache))

	// Setup handler
	var handler http.Handler = &server.SchedExtenderHandler{
		Filter: filter,
	}

	// Add logging for debug mode
	if config.Debug {
		handler = server.LoggingDecorator(handler)
	}

	// start node labeller
	// nodeLabeller := services.NewNodeLabeller(cache, nodeupdater.New(k8sClient),
	// 	vsclient, bcache)
	// go nodeLabeller.Run(wait.NeverStop)

	// Start DRSRuler
	ruler := services.NewDRSRuler(cache.PodInformer(), bcache, cache, vsclient)
	go ruler.Run(wait.NeverStop)

	go cache.Run(wait.NeverStop)

	// Start scheduler extender
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: handler,
	}

	log.Printf("start kubernetes scheduler extender on :%d", config.Port)
	s.ListenAndServe()
}
