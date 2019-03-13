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

package main

import (
	"fmt"
	"net/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/julienschmidt/httprouter"
	"time"
)

const (
	MetricsPrefix = "loginapp_"
)

var (
	TotalRequestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: MetricsPrefix+"request_total",
		Help: "The total number of http request",
	}, []string{"code", "method"})

	RequestDuration = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: MetricsPrefix+"request_duration",
		Help: "Duration of http request in seconds",
	}, []string{"code", "method"})
)

func prometheusRouter() *httprouter.Router {
	pr := httprouter.New()
	pr.Handler("GET", "/metrics", promhttp.Handler())
	return pr
}

// PrometheusMetrics setup prometheus metrics exporter
func PrometheusMetrics(port int) error {
	if err := fmt.Errorf("%v", http.ListenAndServe(fmt.Sprintf(":%v", port), LoggingHandler(prometheusRouter()))); err != nil {
		return err
	}
	return nil
}

func PromIncRequest(sc int, m string) {
	TotalRequestCounter.With(prometheus.Labels{"code": fmt.Sprintf("%v",sc), "method": m}).Inc()
}

func PromAddRequestDuration(sc int, m string, d time.Duration) {
	RequestDuration.With(prometheus.Labels{"code": fmt.Sprintf("%v",sc), "method": m}).Add(d.Seconds())
}
