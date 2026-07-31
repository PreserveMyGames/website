#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
STATIC_DIR="internal/staticfiles/static"

mkdir -p "${STATIC_DIR}/vendor"/{htmx,fuse,feather,fonts}

HTMX_VERSION="2.0.4"
FUSE_VERSION="7.0.0"
FEATHER_VERSION="v4.29.2"
FONT_VERSION="1.1.0"

curl -fsSL "https://unpkg.com/htmx.org@${HTMX_VERSION}/dist/htmx.min.js" \
  -o "${STATIC_DIR}/vendor/htmx/htmx.min.js"

curl -fsSL "https://unpkg.com/fuse.js@${FUSE_VERSION}/dist/fuse.min.js" \
  -o "${STATIC_DIR}/vendor/fuse/fuse.min.js"

for icon in sun moon search menu x globe chevron-down; do
  curl -fsSL "https://raw.githubusercontent.com/feathericons/feather/${FEATHER_VERSION}/icons/${icon}.svg" \
    -o "${STATIC_DIR}/vendor/feather/${icon}.svg"
done

curl -fsSL "https://github.com/IBM/plex/releases/download/v${FONT_VERSION}/OpenType.zip" \
  -o /tmp/plex-opentype.zip
unzip -qo /tmp/plex-opentype.zip -d /tmp/plex-opentype
install -D -m 0644 /tmp/plex-opentype/OpenType/IBM-Plex-Sans/IBMPlexSans-Regular.otf \
  "${STATIC_DIR}/vendor/fonts/IBMPlexSans-Regular.otf"
install -D -m 0644 /tmp/plex-opentype/OpenType/IBM-Plex-Sans/IBMPlexSans-SemiBold.otf \
  "${STATIC_DIR}/vendor/fonts/IBMPlexSans-SemiBold.otf"
install -D -m 0644 /tmp/plex-opentype/OpenType/IBM-Plex-Sans/IBMPlexSans-Bold.otf \
  "${STATIC_DIR}/vendor/fonts/IBMPlexSans-Bold.otf"
rm -rf /tmp/plex-opentype /tmp/plex-opentype.zip

mkdir -p "${STATIC_DIR}/css" "${STATIC_DIR}/js"

cat > "${STATIC_DIR}/vendor/htmx/LICENSE.txt" <<'EOF'
htmx is distributed under the BSD 2-Clause License.
https://github.com/bigskysoftware/htmx
EOF

cat > "${STATIC_DIR}/vendor/fuse/LICENSE.txt" <<'EOF'
Fuse.js is distributed under the Apache License 2.0.
https://github.com/krisk/Fuse
EOF

cat > "${STATIC_DIR}/vendor/feather/LICENSE.txt" <<'EOF'
Feather icons are distributed under the MIT License.
https://github.com/feathericons/feather
EOF

cat > "${STATIC_DIR}/vendor/fonts/LICENSE.txt" <<'EOF'
IBM Plex Sans is distributed under the SIL Open Font License 1.1.
https://github.com/IBM/plex
EOF

echo "Vendor assets updated."
