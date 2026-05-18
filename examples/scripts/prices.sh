#!/usr/bin/env bash
# Simulates a crypto spot price for one symbol using a random walk with mild
# mean reversion. State is persisted to /tmp/dashbrew-sim/ so consecutive
# calls produce smooth-ish movement (good for chart append-mode visuals).
#
# Usage: prices.sh SYMBOL
# Output: a single float on stdout, the new current price.

set -e

sym="${1:-BTC}"
dir="${DASHBREW_SIM_DIR:-/tmp/dashbrew-sim}"
mkdir -p "$dir"
file="$dir/$sym"

# Per-symbol baseline and absolute step size (volatility). Stablecoins use
# a much tighter step and get clamped near 1.0 below.
case "$sym" in
  BTC)  init=80000   ; step=120     ;;
  ETH)  init=3500    ; step=8       ;;
  SOL)  init=180     ; step=0.7     ;;
  BNB)  init=650     ; step=1.4     ;;
  XRP)  init=2.5     ; step=0.015   ;;
  ADA)  init=1.1     ; step=0.008   ;;
  TRX)  init=0.27    ; step=0.002   ;;
  DOGE) init=0.4     ; step=0.003   ;;
  USDT) init=1.0     ; step=0.0006  ;;
  USDC) init=1.0     ; step=0.0006  ;;
  *)    init=100     ; step=0.5     ;;
esac

if [ ! -f "$file" ]; then
  printf '%s\n' "$init" > "$file"
fi
current=$(cat "$file")

r1=$RANDOM
r2=$RANDOM

new=$(awk -v c="$current" -v s="$step" -v i="$init" -v r1="$r1" -v r2="$r2" '
  BEGIN {
    direction = (r1 % 2 == 0) ? 1 : -1
    magnitude = (r2 % 1000) / 1000.0
    # Soft mean reversion: a tiny pull back toward the baseline so prices
    # do not drift off to infinity over a long session.
    drift = (i - c) * 0.001
    new = c + direction * s * magnitude + drift
    if (new < 0) new = i
    # Stablecoins clamp tightly to baseline.
    if (i == 1.0 || i < 0.01) {
      max_dev = i * 0.005
      if (new > i + max_dev) new = i + max_dev
      if (new < i - max_dev) new = i - max_dev
    }
    printf "%.6f", new
  }
')

printf '%s\n' "$new" > "$file"
printf '%s\n' "$new"
