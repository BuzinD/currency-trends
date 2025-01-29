package app

import (
	"cur/internal/config"
	"os"

	log "github.com/sirupsen/logrus"
)

type App struct {
	config *config.Config
	log    *log.Logger
}

func (app *App) initLogger() {
	app.log.SetFormatter(&log.JSONFormatter{})
	app.log.SetOutput(os.Stdout)
}

var app *App

func StartApplication() error {
	app = newApp()
	app.initConfig()
	app.initLogger()

	return nil
}

func newApp() *App {
	return &App{}
}

func (app *App) initConfig() error {
	config.LoadEnvs()               // Load environment variables
	app.config = config.NewConfig() // Initialize configuration
	return nil                      //todo
}

func (app *App) Log() *log.Logger {
	return app.log
}
