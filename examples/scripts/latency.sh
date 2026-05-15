#!/usr/bin/env bash
# Six p99 latency samples per call, centered around a moving baseline so the
# chart has texture. Append mode in dashbrew turns each call into a trend.

baseline=$((35 + RANDOM % 25))
for i in $(seq 1 6); do
  jitter=$((RANDOM % 18))
  # Occasional latency spike for visual interest.
  spike=0
  if [ $((RANDOM % 10)) -eq 0 ]; then
    spike=$((40 + RANDOM % 80))
  fi
  echo $((baseline + jitter + spike))
done
