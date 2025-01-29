package app

import (
	"cur/internal/config"
	"cur/internal/infrastructure/dbConnection"
	"cur/internal/service/okx"
	"cur/internal/store"
	"os"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

type App struct {
	config *config.Config
	log    *log.Logger
	store  *store.Store
}

func (app *App) initLogger() {
	app.log = log.New()
	app.log.SetFormatter(&log.JSONFormatter{})
	app.log.SetOutput(os.Stdout)
}

func StartApplication() error {
	app := newApp()
	app.initConfig()
	app.initLogger()
	app.initStore()

	okxService := okx.NewOkxService(app.store.Currency()) //TODO set config during creation

	c := cron.New()

	// Running at 9:00 и 21:00 every day
	c.AddFunc("0 8,21 * * *", func() {
		if err := okxService.UpdateCurrencies(app.Config().OkxApiConfig()); err != nil {
			app.log.Error(err)
		}
	})

	// Running at 8:00 every day
	c.AddFunc("0 9 * * *", func() {
		if err := okx.UpdateTickers(app.Config()); err != nil {
			app.log.Error(err)
		}
	})

	// running
	c.Start()

	// Waiting
	select {}
}

func newApp() *App {
	return &App{}
}

func (app *App) initConfig() error {
	config.LoadEnvs()               // Load environment variables
	app.config = config.NewConfig() // Initialize configuration
	return nil                      //todo
}

func (app *App) initStore() error {
	db, err := dbConnection.GetDbConnection()
	if err != nil {
		return err
	}

	app.store = store.NewStore(db) // Initialize configuration
	return nil
}

func (app *App) Log() *log.Logger {
	return app.log
}

func (app *App) Config() *config.Config {
	return app.config
}
