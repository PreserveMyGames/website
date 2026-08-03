#!/bin/sh
set -e

git_rev="$(git rev-parse --short=12 HEAD 2>/dev/null || echo dev)"
content_hash="$(find internal/i18n internal/render/templates internal/staticfiles/static -type f ! -path '*/.*' 2>/dev/null | LC_ALL=C sort | xargs sha256sum 2>/dev/null | sha256sum | awk '{print substr($1, 1, 12)}')"
if [ -z "$content_hash" ]; then
	content_hash="nocontent"
fi
printf '%s-%s' "$git_rev" "$content_hash"
