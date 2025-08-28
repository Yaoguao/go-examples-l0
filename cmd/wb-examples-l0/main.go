package main

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
	"wb-examples-l0/internal/models"
	"wb-examples-l0/internal/storage/cache"
	"wb-examples-l0/internal/storage/postgres"
)

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

	// middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(log2.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

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

	//	TEST -------------------------------
	//order := createTestOrder()
	//
	//time.Sleep(1 * time.Second)
	//
	//v := validator.New()
	//models.ValidateOrder(v, &order)
	//
	//if !v.Valid() {
	//	log.Error("error", v.Errors)
	//	return
	//}
	//
	////err = storage.SaveOrder(&order)
	////if err != nil {
	////	return
	////}
	//
	////cache.Put(order.OrderUID, &order)
	//valdb, err := storage.GetOrderByUID(order.OrderUID)
	//if err != nil {
	//	return
	//}
	//log.Debug("order from db", valdb)
	//
	//valcache, _ := cache.Get(order.OrderUID)
	//log.Debug("order from cache", valcache)
	//
	//log.Info("SERVER VSEE")

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

func createTestOrder() models.Order {
	return models.Order{
		OrderUID:    "b563feb7b2b84b6test",
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Delivery: models.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: models.Payment{
			Transaction:  "b563feb7b2b84b6test",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []models.Item{
			{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				Rid:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       time.Date(2021, 11, 26, 6, 22, 19, 0, time.UTC),
		OofShard:          "1",
	}
}
