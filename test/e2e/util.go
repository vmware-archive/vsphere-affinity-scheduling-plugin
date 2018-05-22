package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	schedv1 "k8s.io/kubernetes/pkg/scheduler/api/v1"
)

func postFilterRequest(url string, data interface{}) (*schedv1.ExtenderFilterResult, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http: bad status %s", resp.Status)
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &schedv1.ExtenderFilterResult{}
	return result, json.Unmarshal(d, result)
}
