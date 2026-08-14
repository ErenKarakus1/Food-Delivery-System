package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ErenKarakus1/Food-Delivery-System/delivery-service/internal/model"
	"github.com/ErenKarakus1/Food-Delivery-System/delivery-service/internal/repository"
	"github.com/ErenKarakus1/Food-Delivery-System/delivery-service/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Worker struct {
	pool       *pgxpool.Pool
	httpClient *http.Client
}

func getOrdersReadyForPickup(client *http.Client) ([]model.Order, error) {
	url := "http://localhost:8082/orders/courier/ready-for-pickup"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return []model.Order{}, errors.New("internal server error")
	}
	req.Header.Set("X-User-Role", "courier")
	resp, err := client.Do(req)
	if err != nil {
		return []model.Order{}, errors.New("order service error")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []model.Order{}, errors.New("order service error")
	}
	var orders []model.Order
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		return []model.Order{}, errors.New("order service error")
	}
	if orders == nil {
		return []model.Order{}, nil
	}
	return orders, nil
}

func insertNewDeliveriesFromOrders(ctx context.Context, pool *pgxpool.Pool, client *http.Client, orders []model.Order) {
	for _, order := range orders {
		delivery := service.CreateDelivery(order.ID)
		if err := repository.CreateDelivery(ctx, pool, delivery); err != nil {
			continue
		}
		url := fmt.Sprintf("http://localhost:8082/orders/courier/%s/delivery_created", order.ID)
		req, err := http.NewRequest(http.MethodPatch, url, nil)
		if err != nil {
			continue
		}
		req.Header.Set("X-User-Role", "courier")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			continue
		}
	}
}

func assignCouriersToDeliveries(ctx context.Context, pool *pgxpool.Pool, deliveries []model.Delivery) {
	for _, delivery := range deliveries {
		courier, err := repository.GetAvailableCourierForDelivery(ctx, pool, delivery.ID)
		if err != nil {
			continue
		}
		if err := repository.AssignDelivery(ctx, pool, delivery.ID, courier.ID); err != nil {
			continue
		}
		if err := repository.SetUnavailable(ctx, pool, courier.ID); err != nil {
			continue
		}

	}
}

func (w *Worker) poll(ctx context.Context) {
	orders, err := getOrdersReadyForPickup(w.httpClient)
	if err != nil {
		return
	}
	insertNewDeliveriesFromOrders(ctx, w.pool, w.httpClient, orders)
	deliveries, err := repository.GetDeliveriesWaitingAssignment(ctx, w.pool)
	if err != nil {
		return
	}
	assignCouriersToDeliveries(ctx, w.pool, deliveries)
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	w.poll(ctx)
	for {
		select {
		case <-ticker.C:
			w.poll(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func NewWorker(pool *pgxpool.Pool) *Worker {
	return &Worker{
		pool:       pool,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}
