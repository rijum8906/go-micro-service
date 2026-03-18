#!/bin/sh

cat > /app/dist/config.js << EOF
window.__CONFIG__ = {
  GRAPHQL_URL: "${GRAPHQL_URL:-http://localhost:8080/query}"
}
EOF

serve -s dist -l 3000