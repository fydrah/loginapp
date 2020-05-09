// Copyright 2018 fydrah
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Some code comes from @ericchiang (Dex - CoreOS)

package server

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}

// LoggingHandler catch requests,
// add metadata and log user requests
func LoggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw := newLoggingResponseWriter(w)
		t1 := time.Now()
		next.ServeHTTP(lw, r)
		t2 := time.Now()
		log.WithFields(log.Fields{
			"method":           r.Method,
			"path":             r.URL.String(),
			"request_duration": t2.Sub(t1).String(),
			"protocol":         r.Proto,
			"remote_address":   r.RemoteAddr,
			"code":             lw.statusCode,
		}).Info()
		PromIncRequest(lw.statusCode, r.Method)
		PromAddRequestDuration(lw.statusCode, r.Method, t2.Sub(t1))
	})
}
