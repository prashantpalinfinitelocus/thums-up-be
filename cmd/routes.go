package cmd

import (
	"github.com/gin-gonic/gin"

	"github.com/Infinite-Locus-Product/thums_up_backend/middlewares"
	"github.com/Infinite-Locus-Product/thums_up_backend/routes"
)

func (s *Server) setupRouter() *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middlewares.CORSMiddleware())
	router.Use(middlewares.ErrorHandler())

	routes.SetupHealthAndDocs(router)
	s.setupAPIRoutes(router)

	return router
}

func (s *Server) setupAPIRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")

	routes.SetupAuthRoutes(api, s.handlers.auth)

	routes.SetupProfileRoutes(
		api,
		s.db,
		s.repositories.user,
		s.handlers.profile,
		s.handlers.address,
	)

	routes.SetupQuestionRoutes(
		api,
		s.db,
		s.repositories.user,
		s.handlers.question,
	)

	routes.SetupThunderSeatRoutes(
		api,
		s.db,
		s.repositories.user,
		s.handlers.thunderSeat,
	)

	routes.SetupWinnerRoutes(api, s.handlers.winner)

	routes.SetupAvatarRoutes(
		api,
		s.db,
		s.repositories.user,
		s.handlers.avatar,
	)

	routes.SetupAdminRoutes(api, s.handlers.winner)
}
