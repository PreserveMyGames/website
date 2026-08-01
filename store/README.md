# Preserve My Games Store

Independent store application for [store.preservemygames.org](https://store.preservemygames.org).

This directory is a **standalone Go module**. You can copy `./store` out of the parent website repo and run it on its own. It does not import anything from the parent website module.

## Features

- SQLite-backed catalog, stock reservations, and orders
- Mixed digital and physical products (YAML catalog files)
- Stripe Checkout (card payments)
- Self-hosted Monero payments via `monero-wallet-rpc`
- Display-only FX conversion (Frankfurter rates)
- Basic-auth admin UI for orders and stock
- Empty catalog state when no products exist yet

## Quick start

```bash
cd store
go mod tidy
make run
```

Open http://localhost:8080/en/

## Product catalog

Add YAML files under `internal/catalog/content/`:

```yaml
slug: example-sticker
type: physical
title: Example sticker
description: A sample physical item.
price_cents: 500
currency: USD
stock: 25
shipping_class: small
active: true
```

Digital example:

```yaml
slug: example-guide
type: digital
title: Example guide
description: A downloadable guide.
price_cents: 900
currency: USD
stock: null
digital_delivery_kind: url
digital_payload: https://example.com/download/guide.pdf
active: true
```

`stock: null` (or omitted) means unlimited. Live stock is preserved across catalog syncs once set in the database.

## Environment

| Variable | Default | Notes |
|----------|---------|-------|
| `PORT` | `8080` | Listen port |
| `APP_ENV` | `development` | `production` enables HSTS |
| `SITE_URL` | `http://localhost:8080` | Public store URL |
| `DATA_DIR` | `./data` | SQLite directory |
| `BASE_CURRENCY` | `USD` | Charge currency |
| `MAIN_SITE_URL` | `https://preservemygames.org` | Nav link back to the main site |
| `STRIPE_SECRET_KEY` | | Enables Stripe checkout |
| `STRIPE_WEBHOOK_SECRET` | | Stripe webhook signing secret |
| `MONERO_RPC_URL` | | Enables Monero checkout |
| `MONERO_RPC_USER` | | Optional RPC basic auth |
| `MONERO_RPC_PASS` | | Optional RPC basic auth |
| `MONERO_MIN_CONFIRMATIONS` | `10` | Required confirmations |
| `ADMIN_USER` | | Enables `/internal/orders` |
| `ADMIN_PASSWORD` | | Admin basic auth password |

## Docker

```bash
make docker-build
docker compose -f docker/docker-compose.yml up
```

Compose includes the store app, a pruned `monerod`, and `monero-wallet-rpc`. Expect significant disk and sync time for the Monero node.

## Commands

```bash
make ci      # fmt, vet, test, build
make test
make build
make run
```
