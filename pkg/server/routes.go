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

package server

import (
	"os"
	"net/http"

	"github.com/gobuffalo/packr/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// GetAssetsFS returns http.FyleSystem for assets
func (s *Server) GetAssetsFS() http.FileSystem {
	assetsDir, err := os.Stat(s.Config.Web.AssetsDir)
	if (err != nil || !assetsDir.IsDir()) {
		log.Debug("assets directory not found. using embedded assets")
		return packr.New("assets", "../../web/assets/")
	} else {
		log.Debug("assets directory found. using assets from directory")
		return http.Dir(s.Config.Web.AssetsDir)
	}
}

// Routes setup the server router
func (s *Server) Routes() {
	s.router.GET("/", s.HandleLogin)
	s.router.GET("/callback", s.HandleGetCallback)
	s.router.GET("/healthz", s.HandleGetHealthz)
	s.router.ServeFiles("/assets/*filepath", s.GetAssetsFS())
	log.Debug("routes loaded")
}

// PrometheusRoutes setup the prometheus router
func (s *Server) PrometheusRoutes() {
	s.promrouter.Handler("GET", "/metrics", promhttp.Handler())
	log.Debug("prometheus routes loaded")
}
