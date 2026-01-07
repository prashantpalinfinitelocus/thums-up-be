package cmd

import (
	"github.com/gin-gonic/gin"

	"github.com/Infinite-Locus-Product/thums_up_backend/middlewares"
	"github.com/Infinite-Locus-Product/thums_up_backend/routes"
)

func (s *Server) setupRouter() *gin.Engine {
	router := gin.New()

	router.RedirectTrailingSlash = false
	router.MaxMultipartMemory = 200 << 20

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middlewares.CORSMiddleware())
	router.Use(middlewares.ErrorHandler())

	routes.SetupHealthAndDocs(router)
	s.setupAPIRoutes(router)

	return router
}

func (s *Server) setupAPIRoutes(router *gin.Engine) {
	api := router.Group("/backend/api/v1")

	routes.SetupAuthRoutes(api, s.handlers.auth, s.db, s.repositories.user)

	routes.SetupProfileRoutes(
		api,
		s.db,
		s.repositories.user,
		s.handlers.profile,
		s.handlers.address,
		s.handlers.question,
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

	routes.SetupWinnerRoutes(
		api,
		s.db,
		s.repositories.user,
		s.handlers.winner,
	)

	routes.SetupContestWeekRoutes(
		api,
		s.db,
		s.repositories.user,
		s.handlers.contestWeek,
	)

	routes.SetupAvatarRoutes(
		api,
		s.db,
		s.repositories.user,
		s.handlers.avatar,
	)

	routes.SetupWebsiteStatusRoutes(api, s.handlers.websiteStatus)

	routes.SetupStateRoutes(api, s.handlers.state)

	routes.SetupAdminRoutes(api, s.handlers.winner)
}
