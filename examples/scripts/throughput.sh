#!/usr/bin/env bash
# Pulsing values for a throughput chart. Returns 6 floats per call. Centered
# around a slowly moving baseline so the chart looks like a real traffic
# signal rather than pure white noise.

baseline=$((40 + RANDOM % 30))
for i in $(seq 1 6); do
  jitter=$((RANDOM % 25))
  spike=0
  # Occasional spike to add visual interest.
  if [ $((RANDOM % 9)) -eq 0 ]; then
    spike=$((RANDOM % 40))
  fi
  echo $((baseline + jitter + spike))
done
