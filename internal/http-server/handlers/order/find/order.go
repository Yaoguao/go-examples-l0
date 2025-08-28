package find

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"wb-examples-l0/internal/models"
	"wb-examples-l0/internal/storage/cache"
)

type response struct {
	Order *models.Order `json:"order"`
	Error string        `json:"error,omitempty"`
}

type OrderFinder interface {
	GetOrderByUID(orderUID string) (*models.Order, error)
}

func New(log *slog.Logger, orderFinder OrderFinder, cache *cache.LRUCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.order.find.New"

		ctx := r.Context()
		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(ctx)),
		)

		uid := chi.URLParam(r, "order_uid")
		if uid == "" {
			log.Error("orderUID is required")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response{Error: "orderUID is required"})
			return
		}

		if cachedOrder, exists := cache.Get(uid); exists {
			log.Debug("order found in cache", "order_uid", uid)
			render.Status(r, http.StatusOK)
			render.JSON(w, r, response{Order: cachedOrder})
			return
		}

		log.Debug("order not found in cache, querying database", "order_uid", uid)

		order, err := orderFinder.GetOrderByUID(uid)
		if err != nil {
			log.Error("failed to get order from database", "error", err, "order_uid", uid)
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, response{Error: "Order not found"})
			return
		}

		cache.Put(uid, order)
		log.Info("order added to cache", "order_uid", uid)

		render.Status(r, http.StatusOK)
		render.JSON(w, r, response{Order: order})
		log.Info("order found successfully", "order_uid", uid)
	}
}
