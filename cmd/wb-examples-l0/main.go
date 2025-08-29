package main

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wb-examples-l0/internal/config"
	"wb-examples-l0/internal/http-server/handlers/order/find"
	log2 "wb-examples-l0/internal/http-server/middleware/logger"
	"wb-examples-l0/internal/kafka"
	"wb-examples-l0/internal/lib/logger/sl"
	"wb-examples-l0/internal/storage/cache"
	"wb-examples-l0/internal/storage/postgres"

	_ "wb-examples-l0/docs" // импортируем сгенерированные docs
)

// @title WB L0 Orders API
// @version 1.0
// @description API для работы с заказами WB L0
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
// @BasePath /
// @schemes http
func main() {
	cfg := config.MustLoad()

	log := sl.InitLogger(cfg.Env, os.Stdout)

	log.Debug("config", cfg)

	storage, err := postgres.New(cfg.Storage.PostgresDB_DSN)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	/*cache := */
	cache := cache.NewLRUCache(cfg.LruCache.Capacity, storage, log)

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"null", "file://", "http://localhost:3000", "http://127.0.0.1:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(log2.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	router.Get("/order/{order_uid}", find.New(log, storage, cache))

	orderConsumer, err := kafka.NewConsumer(
		cfg.Kafka.Addresses,
		cfg.Kafka.Consumer.OrderTopic,
		cfg.Kafka.Consumer.OrderGroup,
		kafka.NewOrderHandler(log, storage),
	)
	if err != nil {
		log.Info(err.Error(), nil)
		os.Exit(1)
	}

	go orderConsumer.Start()

	err = serve(log, cfg, router)
	log.Info("stop", err)
	if err := orderConsumer.Stop(); err != nil {
		log.Info(err.Error(), nil)
	}

}

func serve(log *slog.Logger, cfg *config.Config, h http.Handler) error {
	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      h,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit

		log.Info("caught signal", map[string]string{
			"signal": s.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	log.Info("starting server wb-examples-l0",
		slog.String("env", cfg.Env),
		slog.String("port", cfg.HTTPServer.Address),
	)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	log.Info("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
