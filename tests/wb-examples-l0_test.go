package find

import (
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// Попытка в тесты :D
const baseURL = "http://localhost:8081"

func TestGetOrder_AgainstRunningServer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping functional test in short mode")
	}

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  baseURL,
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewCompactPrinter(t),
		},
	})

	t.Run("get existing order", func(t *testing.T) {
		existingOrderUID := "b563feb7b2b84b6052695"

		e.GET("/order/{order_uid}", existingOrderUID).
			Expect().
			Status(200).
			JSON().
			Object().
			ContainsKey("order")
	})

	t.Run("order not found", func(t *testing.T) {
		nonExistentUID := "nonExitUID_84b6296329"

		e.GET("/order/{order_uid}", nonExistentUID).
			Expect().
			Status(404).
			JSON().
			Object().
			HasValue("error", "Order not found")
	})

	t.Run("caching behavior", func(t *testing.T) {
		orderUID := "b563feb7b2b84b6900839"

		start := time.Now()
		e.GET("/order/{order_uid}", orderUID).
			Expect().
			Status(200)
		timeDb := time.Since(start)

		time.Sleep(300 * time.Millisecond)

		start = time.Now()
		e.GET("/order/{order_uid}", orderUID).
			Expect().
			Status(200)
		timeCache := time.Since(start)

		assert.True(t, timeCache < timeDb,
			"Cache should be faster. DB: %v, Cache: %v", timeDb, timeCache)
	})
}

func TestGetOrder_ResponseSchema(t *testing.T) {
	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  baseURL,
		Reporter: httpexpect.NewAssertReporter(t),
	})

	e.GET("/order/{order_uid}", "b563feb7b2b84b6778664").
		Expect().
		Status(200).
		JSON().
		Object().
		ContainsKey("order").
		Value("order").Object().
		ContainsKey("order_uid").
		ContainsKey("track_number").
		ContainsKey("entry").
		ContainsKey("delivery").
		ContainsKey("payment").
		Value("order_uid").String()
}
