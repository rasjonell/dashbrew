#!/usr/bin/env bash
# p99 latency samples per call, centered around a moving baseline. Default
# emits 6 samples for the live append-mode chart; pass a larger count for the
# "snapshot" binding that swaps the chart to a denser one-shot view.

count="${1:-6}"
baseline=$((35 + RANDOM % 25))

for i in $(seq 1 "$count"); do
  jitter=$((RANDOM % 18))
  # Occasional latency spike for visual interest.
  spike=0
  if [ $((RANDOM % 10)) -eq 0 ]; then
    spike=$((40 + RANDOM % 80))
  fi
  echo $((baseline + jitter + spike))
done
