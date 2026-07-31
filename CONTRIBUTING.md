# Contributing

Thank you for helping with Preserve My Games.

## Principles

- No tracking, ads, or third-party CDNs
- Pages must work without JavaScript
- Keep dependencies minimal
- Run `make ci` before opening a pull request

## Development setup

```bash
git clone git@github.com:PreserveMyGames/website.git
cd website
make vendor
make run
```

## Deployment

The release artifact is a single static binary (`bin/preservemygames-web`). Set `SITE_URL` and `APP_ENV=production`, then run it directly or use the Docker image. See [README.md](README.md) for configuration and compose files.

## Blog posts

Add Markdown files under `internal/blog/content/blog/{locale}/` with front matter:

```yaml
---
title: Post title
description: Short summary
date: 2026-07-01
updated: 2026-07-15
draft: false
author: Ivan
author_url: https://github.com/Sudo-Ivan
tags:
  - preservation
---
```

Set `draft: true` while work is in progress. Drafts are hidden when `APP_ENV=production`.

## Translations

UI strings live in `internal/i18n/locales/*.json`. Locales are loaded automatically from that directory.

To add a language:

1. Copy `en.json` to `{code}.json` (ISO 639-1, e.g. `fr.json`).
2. Translate all `translation` values. Do not rename `id` fields.
3. Add `lang.{code}` to every locale file for the header language switcher.
4. Add optional blog posts under `internal/blog/content/blog/{code}/`.
5. Run `go test ./internal/i18n/...` and `make ci`.

Templates use `{{t "message.id"}}`. See [README.md](README.md) for the full i18n guide.

## Pull requests

1. Fork the repository
2. Create a branch from `main`
3. Make focused changes
4. Run `make ci`
5. Open a pull request with a clear summary and test plan

`make ci` runs formatting checks, `go fix`, `go vet`, `gosec`, tests, and a release build.

AI agents should read [AGENTS.md](AGENTS.md) for repository conventions.

## Reporting issues

Use GitHub issues for bugs, content requests, and feature ideas. Include steps to reproduce for bugs.

## Security

Do not open public issues for sensitive security reports. Email contact@preservemygames.org instead.
