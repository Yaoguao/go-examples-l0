package find

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"wb-examples-l0/internal/models"
)

type response struct {
	Order *models.Order `json:"order"`
	Error string        `json:"error,omitempty"`
}

type OrderFinder interface {
	GetOrderByUID(orderUID string) (*models.Order, error)
}

func New(log *slog.Logger, orderFinder OrderFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.order.find.New"

		log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		uid := chi.URLParam(r, "order_uid")
		if uid == "" {
			log.Error("orderUID is required")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response{Error: "orderUID is required"})
			return
		}

		order, err := orderFinder.GetOrderByUID(uid)
		if err != nil {
			log.Error("failed to get order", "error", err, "order_uid", uid)
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, response{Error: "Order not found"})
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, response{Order: order})

		//log.Info("order found successfully", "order_uid", uid)
	}
}
