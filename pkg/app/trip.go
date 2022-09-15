package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/haandol/hexagonal/pkg/adapter/primary/consumer"
	"github.com/haandol/hexagonal/pkg/adapter/primary/router"
	"github.com/haandol/hexagonal/pkg/adapter/secondary/producer"
	"github.com/haandol/hexagonal/pkg/config"
	"github.com/haandol/hexagonal/pkg/connector/database"
	"github.com/haandol/hexagonal/pkg/port/primaryport/consumerport"
	"github.com/haandol/hexagonal/pkg/port/primaryport/routerport"
	"github.com/haandol/hexagonal/pkg/port/secondaryport/producerport"
	"github.com/haandol/hexagonal/pkg/util"
)

type TripApp struct {
	server      *http.Server
	routerGroup routerport.RouterGroup
	routers     []routerport.Router
	consumers   []consumerport.Consumer
	producers   []producerport.Producer
}

func NewServer(cfg config.Config, h http.Handler) *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.Port),
		Handler: h,
	}
}

func NewTripApp(
	server *http.Server,
	ginRouter *router.GinRouter,
	tripRouter *router.TripRouter,
	tripConsumer *consumer.TripConsumer,
	tripProducer *producer.TripProducer,
) *TripApp {
	routers := []routerport.Router{
		tripRouter,
	}
	consumers := []consumerport.Consumer{
		tripConsumer,
	}
	producers := []producerport.Producer{
		tripProducer,
	}

	return &TripApp{
		server:      server,
		routerGroup: ginRouter,
		routers:     routers,
		consumers:   consumers,
		producers:   producers,
	}
}

func (app *TripApp) Init() {
	logger := util.GetLogger().With(
		"module", "TripApp",
		"func", "Init",
	)
	logger.Info("Initializing...")

	v1 := app.routerGroup.Group("v1")
	for _, router := range app.routers {
		router.Route(v1)
	}
	logger.Info("routers are initialized.")

	for _, c := range app.consumers {
		c.Init()
	}
	logger.Info("consumers are initialized.")

	util.InitXray()
}

func (app *TripApp) Start() {
	logger := util.GetLogger().With(
		"module", "TripApp",
		"func", "Start",
	)
	logger.Info("Starting...")

	go func() {
		logger.Infow("Started and serving HTTP", "addr", app.server.Addr, "pid", os.Getpid())
		if err := app.server.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				logger.Info("server closed.")
			} else {
				logger.Panicw("ListenAndServe fail", "error", err)
			}
		}
	}()

	for _, c := range app.consumers {
		go c.Consume()
	}
}

func (app *TripApp) Cleanup(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	logger := util.GetLogger().With(
		"module", "TripApp",
		"func", "Cleanup",
	)
	logger.Info("Cleaning up...")

	logger.Info("Shutting down server...")
	if err := app.server.Shutdown(ctx); err != nil {
		logger.Error("Error on server shutdown:", err)
	}
	logger.Info("Server shutdown.")

	logger.Info("Closing database connection...")
	if err := database.Close(ctx); err != nil {
		logger.Error("Error on database close:", err)
	}
	logger.Info("Database connection closed.")

	logger.Info("Closing producers...")
	for _, producer := range app.producers {
		if err := producer.Close(ctx); err != nil {
			logger.Error("Error on producer close:", err)
		}
	}
	logger.Info("Producer connection closed.")

	logger.Info("Closing consumers...")
	for _, c := range app.consumers {
		c.Close(ctx)
	}
	logger.Info("Consumer connection closed.")

	logger.Info("Cleanup done.")
}
