# 🛒 Mini-Commerce App

An event-driven mini-commerce backend built with **Go**, **Kafka**, **PostgreSQL**, **Docker**, and later orchestrated with **Kubernetes** (with **Kyverno** for policies).  
This project is a learning journey in microservices, event-driven architecture, and container orchestration.

---

## 📌 Current Status
- ✅ PostgreSQL running in Docker via `docker-compose.yml`
- ✅ Basic database schema (users, products, carts, orders)
- 🚧 Next: Go services (Cart, Order), Kafka event backbone, Kubernetes deployment

---

## ⚙️ Tech Stack
- **Go** — microservices (Cart, Order, etc.)
- **Kafka** — event backbone for async communication
- **PostgreSQL** — relational database for persistence
- **Docker** — containerized dev environment
- **Kubernetes** — orchestration (later step)
- **Kyverno** — cluster policy enforcement (later step)

---

## 🚀 Getting Started

### Prerequisites
- [Docker Desktop](https://www.docker.com/products/docker-desktop) (with WSL2/Linux containers enabled)
- [psql](https://www.postgresql.org/download/) (Postgres CLI client)
- [Go](https://go.dev/dl/) (for service development — coming later)

### 1. Start PostgreSQL with Docker Compose
```bash
docker compose up -d
```

This starts a Postgres container on port **5432** with a database called `simple_ecommerce`.

### 2. Connect to the Database
```bash
psql -h localhost -p 5432 -U postgres -d simple_ecommerce
# Password: postgres  (default from docker-compose.yml)
```

### 3. Enable UUID Support
```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;
```

### 4. Create an App User (optional, recommended)
```sql
CREATE ROLE app_user WITH LOGIN PASSWORD 'changeme_app';
GRANT CONNECT ON DATABASE simple_ecommerce TO app_user;
GRANT USAGE ON SCHEMA public TO app_user;
```

### 5. Verify Tables
After running schema migrations:
```sql
\dt       -- list tables
\d users  -- describe users table
```

---

## 📂 Database Schema (MVP)

- **users** → customer accounts  
- **products** → store catalog  
- **carts** → active shopping sessions  
- **cart_items** → items in each cart  
- **orders** → completed purchases  
- **order_items** → products inside orders  

---

## 🔮 Roadmap
- [ ] Add Cart service (Go)  
- [ ] Add Order service (Go)  
- [ ] Add Kafka (event backbone)  
- [ ] Add Kubernetes manifests  
- [ ] Add Kyverno policies  

---

## 📝 Notes
- Default Postgres credentials are set in `docker-compose.yml`.  
- For production, **use environment variables** (never commit real passwords).  
- Data persists in the `pgdata` volume created by Docker.
