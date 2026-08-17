# Food Delivery System

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-008ECF?style=for-the-badge&logo=gin&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white)
![RabbitMQ](https://img.shields.io/badge/RabbitMQ-FF6600?style=for-the-badge&logo=rabbitmq&logoColor=white)

An event-driven food delivery platform built as a set of independent Go microservices, fronted by a single API gateway. Customers order from restaurants, restaurant owners manage menus and fulfill orders, and couriers pick up and deliver them.

## Architecture

```mermaid
flowchart LR
    Client --> Gateway

    Gateway --> Auth
    Gateway --> Restaurant
    Gateway --> Order
    Gateway --> Delivery

    Order --> OrderDB[(Order DB)]
    Restaurant --> RestaurantDB[(Restaurant DB)]
    Delivery --> DeliveryDB[(Delivery DB)]
    Auth --> AuthDB[(Auth DB)]

    Order --> RabbitMQ[(RabbitMQ)]
    RabbitMQ --> Delivery
```

## Services

### Auth Service

Responsible for:

* User registration
* User login
* Password hashing with bcrypt
* JWT generation
* Role management

Supported roles:

* customer
* restaurant
* courier

### Restaurant Service

Responsible for:

* Restaurant creation
* Restaurant ownership validation
* Menu management
* Public restaurant browsing

### Order Service

Responsible for:

* Customer orders
* Order lifecycle management
* Order ownership validation
* Restaurant order management
* Publishing order events to RabbitMQ

### Delivery Service

Responsible for:

* Courier management
* Delivery creation
* Event-driven courier assignment
* Delivery lifecycle tracking
* Consuming RabbitMQ events

### API Gateway

Responsible for:

* Single entry point for clients
* JWT validation
* User identity propagation
* Request routing

---

## Tech Stack

### Backend

* Go
* Gin

### Database

* PostgreSQL

### Messaging

* RabbitMQ

### Authentication

* JWT
* bcrypt

### Architecture

* Microservices
* Event-Driven Architecture
* REST APIs

---

## Features

* JWT Authentication
* Role-based authorization
* Independent microservices
* Event-driven delivery assignment with RabbitMQ
* Automatic courier assignment with retry and requeue support
* Restaurant menu management
* Order lifecycle tracking
* API Gateway routing
* Database-per-service architecture

---

## Order Lifecycle

### Restaurant Flow

```mermaid
stateDiagram-v2
    [*] --> pending

    pending --> accepted
    pending --> rejected

    accepted --> preparing
    preparing --> ready_for_pickup

    ready_for_pickup --> delivery_created
    delivery_created --> picked_by_courier
    picked_by_courier --> delivered


    rejected --> [*]
    delivered --> [*]
```

---

## Event Flow

```mermaid
sequenceDiagram
    participant Restaurant
    participant Order
    participant RabbitMQ
    participant Delivery

    Restaurant->>Order: Update status to ready_for_pickup
    Order->>RabbitMQ: Publish OrderReadyForPickup
    RabbitMQ->>Delivery: Consume event
    Delivery->>Delivery: Create delivery
    Delivery->>Delivery: Assign courier
    Delivery->>Order: Update status to delivery_created
```

---

## Courier Assignment Retry Flow

```mermaid
flowchart LR
    A[Restaurant sets ready_for_pickup]
    B[Order Service publishes event]
    C[RabbitMQ]
    D[Delivery Service consumes event]
    E[Create delivery]
    F{Courier available?}
    G[Assign courier]
    H[Mark delivery_created]
    I[Requeue event]

    A --> B
    B --> C
    C --> D
    D --> E
    E --> F

    F -->|Yes| G
    G --> H

    F -->|No| I
    I --> C

```
The Delivery Service:


1. Creates a delivery record
2. Attempts to find an available courier
3. If a courier exists, assigns the courier, marks the courier unavailable, and updates the order status to `delivery_created`
4. If no courier is available, the RabbitMQ message is requeued and retried later

---

## Authentication

All protected endpoints require:

```http
Authorization: Bearer <jwt_token>
```

The API Gateway validates the token and forwards:

```http
X-User-ID
X-User-Role
```

to downstream services.

---

## Getting started

### 1. Prerequisites

- Go 1.26+
- PostgreSQL (a separate database per service)
- RabbitMQ

### 2. Create the databases and run migrations

Each service ships its own migration file — apply each to its own database:

```
auth-service/migrations/001_init.sql
restaurant-service/migrations/001_init.sql
order-service/migrations/001_init.sql
delivery-service/migrations/001_init.sql
```

### 3. Configure environment variables

Each service loads its own `.env` file at startup and exits if a required variable is missing.

**api-gateway/.env**
```
JWT_SECRET=<same value as auth-service>
```

**auth-service/.env**
```
DATABASE_URL=postgres://user:password@localhost:5432/food_delivery_auth
JWT_SECRET=<same value as api-gateway>
```

**restaurant-service/.env**
```
DATABASE_URL=postgres://user:password@localhost:5432/food_delivery_restaurant
```

**order-service/.env**
```
DATABASE_URL=postgres://user:password@localhost:5432/food_delivery_order
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

**delivery-service/.env**
```
DATABASE_URL=postgres://user:password@localhost:5432/food_delivery_delivery
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

> `JWT_SECRET` must match between `api-gateway` and `auth-service`. `RABBITMQ_URL` must point at the same broker for `order-service` and `delivery-service`.

### 4. Run everything

Start RabbitMQ and Postgres, then in separate terminals:

```bash
cd auth-service           && go run ./cmd/server   # :8084
cd restaurant-service     && go run ./cmd/server   # :8081
cd order-service          && go run ./cmd/server   # :8082
cd delivery-service       && go run ./cmd/server   # :8083
cd api-gateway            && go run ./cmd/server   # :8080
```

All client traffic should go through the gateway at `http://localhost:8080`.

---

## Public Endpoints

### Authentication

```http
POST /auth/register
POST /auth/login
```

### Restaurants

```http
GET /restaurants
GET /restaurants/:id
GET /restaurants/:id/menu
GET /menu-items/:id
```

---

## Protected Endpoints

### Customer

```http
POST /orders/customer
GET  /orders/customer
GET  /orders/customer/:id
```

### Restaurant

```http
POST   /restaurants
GET    /restaurants/me

POST   /menu-items
PUT    /menu-items/:id
DELETE /menu-items/:id

GET    /orders/restaurant
GET    /orders/restaurant/:id
PATCH  /orders/restaurant/:id
```

### Courier

```http
GET   /couriers/me

PATCH /couriers/me/available
PATCH /couriers/me/unavailable

GET   /deliveries/me

PATCH /deliveries/me/pickup
PATCH /deliveries/me/deliver
PATCH /deliveries/me/reject

GET   /orders/courier/:id
```

---

## RabbitMQ Events

### Published

#### Order Ready For Pickup

```json
{
  "order_id": "uuid"
}
```

Queue:

```text
order.ready_for_pickup
```

### Consumed

Delivery Service consumes:

```text
order.ready_for_pickup
```

and automatically creates deliveries and assigns couriers. If no courier is available, the message is requeued and retried later.

---

## Security Features

* bcrypt password hashing
* JWT authentication
* Ownership validation
* Role-based authorization
* Input validation
* UUID validation

**Note: Only `api-gateway` should be reachable outside your private network.** Internal services authenticate requests purely by trusting the `X-User-ID` / `X-User-Role` headers the gateway sets — there is no shared secret, mTLS, or independent verification between services. If an internal service is reachable directly, those headers can be forged to impersonate any user or role.

---

## Future Improvements

* Docker support
* Dead-letter queues (DLQ)
* Delayed retries / backoff queues
* Kafka support
* Redis caching
* Distributed tracing
* Service discovery
* Rate limiting
* Circuit breakers
* Kubernetes deployment
* Monitoring & observability