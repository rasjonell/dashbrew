#!/usr/bin/env bash
# Emits the current BTC USD price as a single float. Used with the chart
# component's refresh_mode: "append" so each tick appends one point.
set -e
curl -sf --max-time 8 https://api.coinpaprika.com/v1/tickers/btc-bitcoin \
  | jq -r '.quotes.USD.price'
