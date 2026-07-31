# PreserveMyGames Website

Website for [PreserveMyGames.org](https://preservemygames.org). Go server-rendered HTML with optional HTMX and Fuse.js. No tracking, ads, or telemetry.

Supported languages: English (`en`), German (`de`), Russian (`ru`).

## Single binary deployment

This project ships as **one static binary**. Templates, translations, blog posts, and static assets are embedded at build time. There is no database, no Node runtime, and no external services required at run time.

```bash
make build
export SITE_URL=https://your-domain.example
export APP_ENV=production
./bin/preservemygames-web
```

Copy `bin/preservemygames-web` to any Linux server (or build for your OS with `GOOS`/`GOARCH`), set environment variables, and run it behind a reverse proxy if you like. Idle memory is typically around 16 MB.

Docker is optional and runs the same binary in a minimal Alpine image.

## Requirements

- Go 1.26+ (only for building from source)
- curl and unzip for `make vendor` (refreshing frontend assets)
- Docker (optional, for container deployment)

## Build

```bash
make vendor   # only when updating vendored JS, fonts, or icons
make build
```

Binary output: `bin/preservemygames-web`

Cross-compile example:

```bash
GOOS=linux GOARCH=arm64 make build
```

## Run locally

```bash
export SITE_URL=http://localhost:8080
make run
```

Open `http://localhost:8080`. The root path redirects using `Accept-Language`. Locale URLs use a prefix:

- `http://localhost:8080/en/`
- `http://localhost:8080/de/`
- `http://localhost:8080/ru/`

## Test

```bash
make ci
make test-race
make test-fuzz
```

`make ci` runs formatting checks, `go fix`, `go vet`, `gosec`, tests, and a release build.

## Security

The server applies Linux Landlock sandboxing when supported (`internal/sandbox`). Docker deployments use a read-only root filesystem, dropped capabilities, and resource limits in compose files.

## Docker

Local compose:

```bash
make docker-test    # build, wait for healthy, curl smoke checks
make docker-down
```

For a long-running local container without smoke checks:

```bash
make docker-up
```

Coolify uses `docker/docker-compose.coolify.yml`. Set that path in Coolify and assign a domain with container port `8080`.

## Configuration

Copy `.env.example` to `.env` for local reference. The binary reads environment variables directly.

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Listen port |
| `APP_ENV` | `development` | Set `production` for production mode |
| `SITE_URL` | `http://localhost:8080` | Canonical site URL (used in RSS, sitemap, SEO) |
| `CONTACT_EMAIL` | `contact@preservemygames.org` | Contact address |
| `ACCESS_LOG` | `false` | Enable request logging |
| `SITE_NOTICE` | _(empty)_ | Site banner: `construction`, `maintenance`, `info`, or `warning` |
| `SITE_NOTICE_MESSAGE` | _(empty)_ | Optional custom banner text (overrides default i18n message) |

Wiki and Forums links in the navbar are derived from `SITE_URL` (`wiki.{domain}` and `forums.{domain}`).

## Content

- Blog posts: `internal/blog/content/blog/{locale}/`
- UI translations: `internal/i18n/locales/`
- HTML templates: `internal/render/templates/` (use `{{t "message.id"}}` for translated strings)
- Static assets: `internal/staticfiles/static/`

## Internationalization

### How UI translation works

1. Message strings live in JSON files under `internal/i18n/locales/` (one file per language, e.g. `en.json`, `de.json`).
2. Templates call the `t` function: `{{t "nav.home"}}`.
3. Each locale gets its own render engine at startup. The active language comes from the URL prefix (`/de/about`).
4. The header includes a language switcher built from the same message keys (`lang.en`, `lang.de`, `lang.ru`).
5. `hreflang` alternates and the sitemap include every supported locale automatically.

Locales are discovered from `locales/*.json` at startup. You do not register languages in Go code.

### Add a new language

Example for French (`fr`):

1. Copy `internal/i18n/locales/en.json` to `internal/i18n/locales/fr.json`.
2. Translate every `translation` value. Keep each `id` unchanged.
3. Add labels for the switcher:
   - In every locale file, add or update `lang.fr` with the display name (e.g. `Français` in `fr.json`, `French` in `en.json`).
4. Add blog content under `internal/blog/content/blog/fr/` (optional but recommended for a complete site).
5. Run tests:

```bash
go test ./internal/i18n/...
make ci
```

`TestLocaleKeyParity` fails if any locale file is missing keys compared to the others.

### Template checklist for new UI text

1. Add the message `id` to **all** locale JSON files.
2. Use `{{t "your.message.id"}}` in the template.
3. For dynamic values, use go-i18n template data in the handler (`TWith`) and `{{.Field}}` in the JSON string, as with `contact.body`.

### Blog posts per language

Posts are not auto-translated. Each locale has its own Markdown files:

```
internal/blog/content/blog/en/my-post.md
internal/blog/content/blog/de/my-post.md
internal/blog/content/blog/ru/my-post.md
```

Slugs can match across languages so the language switcher can link to the same post path where translations exist.

## License

MIT. See [LICENSE](LICENSE).
