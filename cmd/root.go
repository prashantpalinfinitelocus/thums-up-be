package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "thums-up-be",
	Short: "Thums Up Backend Service",
	Long:  "Backend service for Thums Up application",
}

func init() {
	initLogging()
}

func initLogging() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	env := os.Getenv("APP_ENV")
	if env == "development" {
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})
	}
}

func Execute() error {
	return rootCmd.Execute()
}

