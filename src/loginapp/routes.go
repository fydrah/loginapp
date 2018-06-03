package main

func (s *Server) Routes() {
	s.router.GET("/", s.HandleGetIndex)
	s.router.GET("/callback/cli", s.HandleGetCallbackCLI)
	s.router.GET("/callback/dashboard", s.HandleGetCallbackDashboard)
	s.router.POST("/login", s.HandlePostLogin)
	s.router.POST("/cli", s.HandlePostCLI)
	s.router.POST("/dashboard", s.HandlePostDashboard)
}
