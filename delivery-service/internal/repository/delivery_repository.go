package repository

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/Food-Delivery-System/delivery-service/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDeliveryNotFound = errors.New("delivery not found")
var ErrCourierNotFound = errors.New("courier not found")

const getCurrentDeliveryByCourierIDQuery = `
	SELECT
		id,
		order_id,
		courier_id,
		status,
		created_at,
		updated_at
	FROM deliveries
	WHERE courier_id=$1
	AND status IN (
		'assigned',
		'picked_by_courier'
	)
	ORDER BY created_at DESC
	LIMIT 1
`

const updateDeliveryStatusQuery = `
	UPDATE deliveries
	SET status=$1
	WHERE id=$2
	RETURNING
		id,
		order_id,
		courier_id,
		status,
		created_at,
		updated_at
`

const rejectDeliveryQuery = `
	UPDATE deliveries
	SET courier_id=$1
	WHERE id=$2
`

const createRejectionQuery = `
	INSERT INTO delivery_rejections (
		id,
		delivery_id,
		courier_id
	)
	VALUES ($1,$2,$3)
`

const unavailableQuery = `
	UPDATE couriers
	SET is_available=FALSE
	WHERE id=$1
`

const availableQuery = `
	UPDATE couriers
	SET is_available=TRUE
	WHERE id=$1
`

func GetCurrentDeliveryByCourierID(ctx context.Context, pool *pgxpool.Pool, courierID uuid.UUID) (model.Delivery, error) {
	var delivery model.Delivery
	err := pool.QueryRow(
		ctx,
		getCurrentDeliveryByCourierIDQuery,
		courierID,
	).Scan(
		&delivery.ID,
		&delivery.OrderID,
		&delivery.CourierID,
		&delivery.Status,
		&delivery.CreatedAt,
		&delivery.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Delivery{}, ErrDeliveryNotFound
		}
		return model.Delivery{}, errors.New("internal server error")
	}
	return delivery, nil
}

func UpdateDeliveryStatus(ctx context.Context, pool *pgxpool.Pool, deliveryID uuid.UUID, status string) (model.Delivery, error) {
	var updatedDelivery model.Delivery
	err := pool.QueryRow(
		ctx,
		updateDeliveryStatusQuery,
		status,
		deliveryID,
	).Scan(
		&updatedDelivery.ID,
		&updatedDelivery.OrderID,
		&updatedDelivery.CourierID,
		&updatedDelivery.Status,
		&updatedDelivery.CreatedAt,
		&updatedDelivery.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Delivery{}, ErrDeliveryNotFound
		}
		return model.Delivery{}, errors.New("internal server error")
	}
	return updatedDelivery, nil
}

func RejectDelivery(ctx context.Context, pool *pgxpool.Pool, deliveryID uuid.UUID, courierID uuid.UUID) error {
	tag, err := pool.Exec(
		ctx,
		rejectDeliveryQuery,
		uuid.Nil,
		deliveryID,
	)
	if err != nil {
		return errors.New("internal server error")
	}
	if tag.RowsAffected() == 0 {
		return ErrDeliveryNotFound
	}
	return nil
}

func CreateDeliveryRejection(ctx context.Context, pool *pgxpool.Pool, deliveryRejection model.DeliveryRejecetion) error {
	_, err := pool.Exec(
		ctx,
		createRejectionQuery,
		deliveryRejection.ID,
		deliveryRejection.DeliveryID,
		deliveryRejection.CourierID,
	)
	if err != nil {
		return errors.New("internal server error")
	}
	return nil
}

func SetUnavailable(ctx context.Context, pool *pgxpool.Pool, courierID uuid.UUID) error {
	tag, err := pool.Exec(
		ctx,
		unavailableQuery,
		courierID,
	)
	if err != nil {
		return errors.New("internal server error")
	}
	if tag.RowsAffected() == 0 {
		return ErrCourierNotFound
	}
	return nil
}

func SetAvailable(ctx context.Context, pool *pgxpool.Pool, courierID uuid.UUID) error {
	tag, err := pool.Exec(
		ctx,
		availableQuery,
		courierID,
	)
	if err != nil {
		return errors.New("internal server error")
	}
	if tag.RowsAffected() == 0 {
		return ErrCourierNotFound
	}
	return nil
}
