package app

import (
	"context"
	"cur/internal/config"
	"cur/internal/infrastructure/dbConnection"
	"cur/internal/service/okx"
	"cur/internal/store"
	"os"
	"os/signal"
	"syscall"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

type App struct {
	config      *config.Config
	log         *log.Logger
	store       *store.Store
	cron        *cron.Cron
	okxService  *okx.OkxService
	cancelStack []context.CancelFunc
}

func (app *App) fetchTrades() {
	// Create context for background task
	ctx, cancel := context.WithCancel(context.Background())
	app.cancelStack = append(app.cancelStack, cancel)

	// Run FetchTrades in background as a goroutine
	go func() {
		app.okxService.FetchTrades(ctx)
	}()
}

func (app *App) initLogger() {
	app.log = log.New()
	app.log.SetFormatter(&log.JSONFormatter{})
	app.log.SetOutput(os.Stdout)
}

func StartApplication() {
	app := newApp()
	err := app.initConfig()
	app.initLogger()
	err = app.initStore()
	app.initOkxService()
	// Handle Graceful Shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	app.fetchTrades()
	app.fetchHistoricalCandlesData()
	app.initScheduledTasks()

	if err != nil {
		app.log.Error(err)
		os.Exit(1)
	}

	<-sigs

	for _, f := range app.cancelStack {
		f()
	}

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
	db, err := dbConnection.GetDbConnection(app.Config().DbConfig())
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

func (app *App) initScheduledTasks() {

	app.cron = cron.New()

	// Running at 8:00 и 21:00 every day
	_, err := app.cron.AddFunc("0 8,21 * * *", func() {
		app.log.Info("process update currencies started")
		if err := app.okxService.UpdateCurrencies(); err != nil {
			app.log.Error(err)
		}
		app.log.Info("process update currencies finished")
	})
	if err != nil {
		app.log.Error(err)
		return
	}

	// Running every hour
	_, err = app.cron.AddFunc("0 * * * *", func() {
		app.log.Info("process update candles started")
		app.okxService.UpdateCandles()
		app.log.Info("process update candles finished")
	})
	if err != nil {
		app.log.Error(err)
		return
	}

	// running
	app.cron.Start()
}

func (app *App) initOkxService() {
	app.okxService = okx.NewOkxService(
		app.store.Currency(),
		app.store.Candle(),
		app.config.OkxApiConfig(),
		app.config.KafkaConfig(),
		app.log,
	)
}

// fetchHistoricalCandlesData fetch candles historical data
func (app *App) fetchHistoricalCandlesData() {
	app.okxService.UpdateHistoricalCandles()
}
