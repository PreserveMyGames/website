# AGENTS.md

Guidance for AI coding agents working on the PreserveMyGames website repository.

## Project summary

Single-binary Go web application for [PreserveMyGames.org](https://preservemygames.org). Server-rendered HTML with optional progressive enhancement (HTMX, Fuse.js). No database, no third-party CDNs, no tracking.

Supported locales: `en`, `de`, `ru` (discovered from `internal/i18n/locales/*.json`).

## Repository layout

| Path | Purpose |
|------|---------|
| `cmd/web/main.go` | Application entrypoint |
| `internal/web/` | HTTP server, routes, middleware, caches |
| `internal/render/templates/` | HTML templates (`{{t "key"}}` for i18n) |
| `internal/i18n/locales/` | UI translation JSON (keep keys in sync across locales) |
| `internal/blog/content/blog/` | Markdown blog posts per locale |
| `internal/staticfiles/static/` | Embedded CSS, JS, vendored assets |
| `docker/` | Dockerfile and compose files |
| `scripts/vendor.sh` | Download vendored frontend assets |

## Commands

```bash
make ci           # format check, go fix, vet, gosec, test, build
make test         # go test ./...
make build        # bin/preservemygames-web
make docker-test  # compose build, health wait, smoke curls
```

Always run `make ci` before finishing a change.

## Conventions

- **Scope**: Minimal diffs. Match existing style and patterns.
- **i18n**: Add every new UI string to all locale files (`en.json`, `de.json`, `ru.json`). Use `{{t "message.id"}}` in templates. `TestLocaleKeyParity` enforces key parity.
- **No JS required**: Core pages must work without JavaScript. Use native HTML (`details`/`summary` for dropdowns) where possible.
- **Security**: No unsafe HTML in markdown. Validate slugs and static paths. No tracking or external CDNs.
- **Comments**: Only for non-obvious logic. No TODOs. No emojis in code or docs.
- **Commits**: Do not commit unless the user asks.

## Templates and UI

- Layout: `internal/render/templates/layout.html`
- Header nav: Home, About, Blog
- Header tools: locale dropdown (globe button + JS panel, `<noscript>` fallback links), theme toggle (icon only)
- Footer: tagline, links (Home, About, Blog, RSS, Privacy, Contact), copyright
- Search exists at `/{{lang}}/search` but is not in the main navbar
- Error pages: localized HTML for 404, 405, and 500. JSON/XML paths keep plain errors.
- Locale home routes use `GET /{lang}/{$}` (exact) so unknown `/en/...` paths return 404.
- Panic recovery middleware and graceful shutdown on SIGINT/SIGTERM.

## Adding a locale

1. Copy `internal/i18n/locales/en.json` to `{code}.json` and translate.
2. Add `lang.{code}` to every locale file.
3. Add blog posts under `internal/blog/content/blog/{code}/` if needed.
4. Run `go test ./internal/i18n/...` and `make ci`.

## Adding a blog post

Create `internal/blog/content/blog/{locale}/slug.md` with YAML front matter (`title`, `description`, `date`, `draft`, `tags`). Drafts are hidden when `APP_ENV=production`.

## Docker

- Local: `docker/docker-compose.yml`
- Production (Coolify): `docker/docker-compose.coolify.yml`
- Container runs `/app/web` as UID 65532 on port 8080

## Environment variables

| Variable | Default | Notes |
|----------|---------|-------|
| `PORT` | `8080` | Listen port |
| `APP_ENV` | `development` | `production` hides draft posts, enables HSTS |
| `SITE_URL` | `http://localhost:8080` | Canonical URL for SEO and feeds |
| `CONTACT_EMAIL` | `contact@preservemygames.org` | Contact page |
| `ACCESS_LOG` | `false` | Request logging |

## What not to do

- Do not add databases or external runtime dependencies.
- Do not load assets from CDNs in production templates.
- Do not register locales manually in Go when JSON files suffice.
- Do not edit generated plan files outside this repo unless asked.
