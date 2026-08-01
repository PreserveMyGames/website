# AGENTS.md

Guidance for AI coding agents working on the Preserve My Games Store.

## Project summary

Standalone Go web store for `store.preservemygames.org`. Server-rendered HTML, SQLite persistence, Stripe and self-hosted Monero payments. This module does not depend on the parent website package.

Supported locales: `en`, `de`, `ru` from `internal/i18n/locales/*.json`.

## Layout

| Path | Purpose |
|------|---------|
| `cmd/store/main.go` | Entrypoint |
| `internal/web/` | HTTP server and routes |
| `internal/catalog/` | Products, stock, orders |
| `internal/db/` | SQLite open and migrations |
| `internal/fx/` | Display-only FX rates |
| `internal/payments/stripe/` | Stripe Checkout + webhooks |
| `internal/payments/monero/` | Wallet RPC, quotes, confirmation poller |
| `internal/render/templates/` | HTML templates |
| `internal/i18n/locales/` | UI translations |
| `docker/` | Dockerfile and compose |

## Conventions

- Keep this module independent. Do not import `github.com/PreserveMyGames/website/...`.
- Add every new UI string to all locale files.
- Core pages must work without JavaScript.
- No tracking or external CDNs.
- Secrets only from environment variables.
- Comments only for non-obvious logic. No TODOs. No emojis.

## Payments and inventory

- Creating a checkout reserves stock until `expires_at` (30 minutes).
- Stock decrements only when an order is marked paid.
- Stripe: Checkout Session + signed webhook.
- Monero: per-order subaddress, XMR quote, background confirmation poller.

## Commands

```bash
make ci
make test
make build
```
