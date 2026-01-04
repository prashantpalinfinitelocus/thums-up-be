package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/Infinite-Locus-Product/thums_up_backend/handlers"
)

func SetupStateRoutes(api *gin.RouterGroup, stateHandler *handlers.StateHandler) {
	api.GET("/states", stateHandler.GetAllStates)
}
