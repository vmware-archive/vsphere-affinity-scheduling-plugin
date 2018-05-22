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

package server

import (
	"io"
	"log"
	"net/http"
)

// LoggingDecorator decorates a http handler to include logging capability
func LoggingDecorator(handler http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		r.Body = &logWrapper{r.Body}
		handler.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}

type logWrapper struct {
	r io.ReadCloser
}

func (l *logWrapper) Read(data []byte) (int, error) {
	n, err := l.r.Read(data)
	log.Print("> " + string(data[:n]))
	return n, err
}

func (l *logWrapper) Close() error {
	return l.r.Close()
}
