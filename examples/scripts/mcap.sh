#!/usr/bin/env bash
# Top crypto market caps formatted as histogram input: "SYMBOL: <billions>".
set -e
curl -sf --max-time 8 'https://api.coinpaprika.com/v1/tickers?limit=8' \
  | jq -r '.[] | "\(.symbol): \(.quotes.USD.market_cap / 1e9 | floor)"'
