/*
Copyright 2018 fydrah

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
	"net/http"
)

func (s *Server) Routes() {
	if s.config.WebOutput.SkipMainPage {
		s.router.GET("/", s.HandleLogin)
		logger.Debug("routes loaded, skipping main page")
	} else {
		s.router.GET("/", s.HandleGetIndex)
		s.router.POST("/login", s.HandleLogin)
		logger.Debug("routes loaded, using main page")
	}
	s.router.GET("/callback", s.HandleGetCallback)
	s.router.GET("/healthz", s.HandleGetHealthz)
	s.router.ServeFiles("/assets/*filepath", http.Dir(s.config.WebOutput.AssetsDir))
}
