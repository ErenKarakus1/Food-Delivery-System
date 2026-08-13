CREATE TABLE IF NOT EXISTS deliveries (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL UNIQUE,
    courier_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'ready_for_pickup',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS couriers (
    id UUID PRIMARY KEY,
    is_available BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS delivery_rejections (
    id UUID PRIMARY KEY,
    delivery_id UUID NOT NULL,
    courier_id UUID NOT NULL REFERENCES couriers(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);