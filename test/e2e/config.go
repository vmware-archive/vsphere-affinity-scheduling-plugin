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
	"log"
	"os"
	"strconv"
)

// Config is the test configuration
type Config struct {
	pluginPath    string
	pluginLogFile string
	govmomiURL    string
	Port          int
}

// Parse parses the configuration
func (c *Config) Parse() {
	var err error

	c.pluginPath = getenv(
		"TARGET",
		"",
		"Target plugin command line path")
	c.pluginLogFile = getenv(
		"PLUGIN_LOG_FILE",
		"/tmp/vsphere-sched-plugin.log",
		"Log file path")
	c.govmomiURL = getenv(
		"GOVMOMI_URL",
		"https://user:pass@127.0.0.1:8989/sdk",
		"ESX or vCenter URL")

	port := getenv(
		"PORT",
		"12346",
		"http port to connect to for http extender")
	c.Port, err = strconv.Atoi(port)
	if err != nil {
		log.Fatalf("failed to parse int for port %s: %s", port, err)
	}
}

func getenv(env, dft, doc string) string {
	value := os.Getenv(env)

	if value != "" {
		return value
	}

	if dft != "" {
		return dft
	}

	log.Fatalf("missing env %s: %s", env, doc)
	return ""
}
