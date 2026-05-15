#!/usr/bin/env bash
# Fake structured log lines for the NOC dashboard.

routes=("/v1/orders" "/v1/users/me" "/v1/checkout" "/health" "/v1/search" "/v1/cart" "/v1/orders/refund" "/static/app.js")
methods=(GET POST GET GET POST PUT DELETE GET)
levels=(info info info info info info warn warn error)
codes=(200 200 200 200 200 201 204 301 304 400 404 500 502)
peers=("10.0.0.4" "10.0.0.7" "10.0.0.12" "10.0.0.18" "10.0.0.23")

for i in $(seq 1 14); do
  ts=$(date "+%H:%M:%S")
  lvl=${levels[$((RANDOM % ${#levels[@]}))]}
  method=${methods[$((RANDOM % ${#methods[@]}))]}
  route=${routes[$((RANDOM % ${#routes[@]}))]}
  code=${codes[$((RANDOM % ${#codes[@]}))]}
  peer=${peers[$((RANDOM % ${#peers[@]}))]}
  lat=$((1 + RANDOM % 380))
  rid=$(printf "%08x" "$RANDOM$RANDOM" 2>/dev/null || printf "%08x" "$RANDOM")
  printf "%s  %-5s %-6s %-21s %s  %3dms  peer=%s  rid=%s\n" \
    "$ts" "$lvl" "$method" "$route" "$code" "$lat" "$peer" "$rid"
done
