#!/usr/bin/env bash
# Formatted crypto ticker block for a text pane:
#
#     $80,452
#     +1.23% 24h
#     vol 35.0b
#
# Backed by prices.sh (no network). The "24h" change is the drift since the
# session's first invocation, snapshot to /tmp/dashbrew-sim/<sym>.baseline.

set -e

sym="${1:-BTC}"
dir="${DASHBREW_SIM_DIR:-/tmp/dashbrew-sim}"
mkdir -p "$dir"
here="$(cd "$(dirname "$0")" && pwd)"

price=$("$here/prices.sh" "$sym")

baseline_file="$dir/${sym}.baseline"
if [ ! -f "$baseline_file" ]; then
  printf '%s\n' "$price" > "$baseline_file"
fi
baseline=$(cat "$baseline_file")

# Format price: thousands separator above 1000, two decimals above 1, four below.
formatted_price=$(awk -v p="$price" 'BEGIN {
  if (p >= 1000) printf "%\047.2f", p
  else if (p >= 1)  printf "%.2f", p
  else              printf "%.4f", p
}')

change=$(awk -v p="$price" -v b="$baseline" 'BEGIN {
  if (b == 0) printf "+0.00"
  else        printf "%+.2f", (p - b) / b * 100
}')

# Volume jitters around a per-symbol baseline. Pick the baseline in shell so
# the awk expression stays simple and portable across BSD / GNU awk.
case "$sym" in
  BTC)  vol_base=35  ;;
  ETH)  vol_base=18  ;;
  SOL)  vol_base=4   ;;
  BNB)  vol_base=2   ;;
  XRP)  vol_base=6   ;;
  USDT) vol_base=110 ;;
  USDC) vol_base=40  ;;
  TRX)  vol_base=1.5 ;;
  DOGE) vol_base=1.8 ;;
  ADA)  vol_base=1.0 ;;
  *)    vol_base=1   ;;
esac

vol=$(awk -v r="$RANDOM" -v base="$vol_base" 'BEGIN {
  jitter = ((r % 100) / 100.0 - 0.5) * base * 0.15
  printf "%.1fb", base + jitter
}')

printf "\n  \$%s\n  %s%% 24h\n  vol %s\n" "$formatted_price" "$change" "$vol"
