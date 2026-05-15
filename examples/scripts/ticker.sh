#!/usr/bin/env bash
# Prints a formatted crypto ticker block for use in a text pane:
#
#     $66,432
#     +1.23% 24h
#     vol 23.4b
#
# Requires curl and jq. Argument is the upper-case symbol (BTC ETH SOL BNB ADA XRP).

set -e
sym="${1:-BTC}"

case "$sym" in
  BTC) slug="btc-bitcoin" ;;
  ETH) slug="eth-ethereum" ;;
  SOL) slug="sol-solana" ;;
  BNB) slug="bnb-binance-coin" ;;
  ADA) slug="ada-cardano" ;;
  XRP) slug="xrp-xrp" ;;
  DOGE) slug="doge-dogecoin" ;;
  *)   slug="btc-bitcoin" ;;
esac

resp=$(curl -sf --max-time 8 "https://api.coinpaprika.com/v1/tickers/$slug" || echo '{}')

price=$(printf '%s' "$resp" | jq -r '.quotes.USD.price // empty')
change=$(printf '%s' "$resp" | jq -r '.quotes.USD.percent_change_24h // empty')
vol=$(printf '%s' "$resp" | jq -r '.quotes.USD.volume_24h // empty')

if [ -z "$price" ]; then
  echo
  echo "  (offline)"
  echo
  exit 0
fi

printf "\n  \$%s\n  %+.2f%% 24h\n  vol %s\n" \
  "$(printf '%s' "$price" | awk '{printf "%\047.2f", $1}')" \
  "$change" \
  "$(printf '%s' "$vol" | awk '{
    n=$1;
    if (n>=1e9) printf "%.1fb", n/1e9;
    else if (n>=1e6) printf "%.0fm", n/1e6;
    else printf "%.0fk", n/1e3;
  }')"
