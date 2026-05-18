#!/usr/bin/env bash
# Dense filtered log output for the replace_pane bindings on devops-noc.json.
# Usage: log-filtered.sh LEVEL  (error | warn).

level="${1:-error}"

case "$level" in
  error)
    codes=(500 500 502 503 504)
    msgs=(
      "upstream connection refused"
      "context deadline exceeded"
      "internal server error"
      "bad gateway from worker pool"
      "panic recovered: nil pointer deref"
      "redis timeout after 2s"
      "circuit breaker open: payments"
      "503 from downstream: shipping-svc"
    )
    label="error"
    ;;
  warn)
    codes=(400 401 403 429 422)
    msgs=(
      "rate limit applied · 500/min"
      "auth token expiring in 60s"
      "permission check failed"
      "deprecation: v1 endpoint sunset 2026-06-01"
      "validation: shipping_zip missing"
      "retry attempt 2/3"
      "slow query (>1s) on orders table"
      "idempotency key reused"
    )
    label="warn "
    ;;
  *)
    echo "unknown level: $level"
    exit 1
    ;;
esac

routes=("/v1/orders" "/v1/checkout" "/v1/auth/refresh" "/v1/users/me" "/v1/cart" "/v1/orders/refund")
methods=("POST" "GET" "POST" "GET" "PUT" "DELETE")
peers=("10.0.0.4" "10.0.0.7" "10.0.0.12" "10.0.0.18" "10.0.0.23")

for i in $(seq 1 30); do
  ts=$(date "+%H:%M:%S")
  route=${routes[$((RANDOM % ${#routes[@]}))]}
  method=${methods[$((RANDOM % ${#methods[@]}))]}
  code=${codes[$((RANDOM % ${#codes[@]}))]}
  msg=${msgs[$((RANDOM % ${#msgs[@]}))]}
  peer=${peers[$((RANDOM % ${#peers[@]}))]}
  rid=$(printf "%08x" "$RANDOM$RANDOM" 2>/dev/null || printf "%08x" "$RANDOM")
  printf "%s  %s  %-6s %-22s  %s  peer=%s  rid=%s · %s\n" \
    "$ts" "$label" "$method" "$route" "$code" "$peer" "$rid" "$msg"
done
