package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/Infinite-Locus-Product/thums_up_backend/handlers"
)

// SetupStateRoutes sets up state-related routes
//
//	@Summary		Setup state routes
//	@Description	Configures all state-related API endpoints
//	@Tags			Location
func SetupStateRoutes(api *gin.RouterGroup, stateHandler *handlers.StateHandler) {
	// GET /api/v1/states - Get all active states
	// See handlers/state_handler.go for swagger documentation
	api.GET("/states", stateHandler.GetAllStates)
}
