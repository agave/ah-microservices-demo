package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/agave/ah-microservices/services/users/db"
	"github.com/agave/ah-microservices/services/users/util"
	config "github.com/gypsydiver/go-config"
)

func init() {
	configureLogger()
	configInit()
	dbClientInit()
}

func configureLogger() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func configInit() {
	if err := config.GetConfig(&util.Config, "config.yml"); err != nil {
		log.WithField("error", err).Fatalln("Error loading configuration")
	}
}

func dbClientInit() {
	url := util.FormatDBURL("postgres", &util.Config.DB)
	err := db.InitDB("postgres", url)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"url":   url,
		}).Fatalln("Error starting database")
	}
}

func main() {
	log.Info("Awesome sauce")
	// events.LaunchConsumer()
	// events.LaunchProducer()
	// server.StartServer()
}