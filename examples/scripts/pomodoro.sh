#!/usr/bin/env bash
# Synthetic focus-level samples for the pomodoro chart. Pattern: 25 minute
# work phase + 5 minute break, repeating. Returns one number per call (with
# jitter) so the chart in append mode draws a sawtooth-ish trace of focus.

minute=$(date +%M)
cycle_pos=$((10#$minute % 30))

if [ "$cycle_pos" -lt 25 ]; then
  base=82  # work phase
else
  base=18  # break phase
fi

jitter=$((RANDOM % 14))
echo $((base + jitter))
