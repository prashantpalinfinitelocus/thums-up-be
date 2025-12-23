package main

import (
	"github.com/Infinite-Locus-Product/thums_up_backend/cmd"
	_ "github.com/Infinite-Locus-Product/thums_up_backend/docs"
)

// @title Thums Up Backend API
// @version 1.0
// @description API documentation for Thums Up Backend Service including Contest Week Management

// @contact.name API Support
// @contact.email support@thumsup.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token

// @securityDefinitions.apikey APIKey
// @in header
// @name X-API-Key
// @description Admin API key for protected operations

func main() {
	cmd.Execute()
}
