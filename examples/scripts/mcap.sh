#!/usr/bin/env bash
# Simulated market caps (USD billions) for the trading-desk histogram. Each
# bar jitters by a small percent every refresh so the gradient repaints
# visibly. Order roughly matches reality.

vary() {
  awk -v base="$1" -v r="$RANDOM" 'BEGIN {
    jitter = (r % 100) / 100.0 - 0.5
    n = base + jitter * base * 0.02
    if (n < 1) printf "%.1f", n; else printf "%.0f", n
  }'
}

echo "BTC:  $(vary 1610)"
echo "ETH:  $(vary 270)"
echo "USDT: $(vary 190)"
echo "BNB:  $(vary 95)"
echo "XRP:  $(vary 90)"
echo "USDC: $(vary 75)"
echo "SOL:  $(vary 50)"
echo "TRX:  $(vary 30)"
