#!/usr/bin/env bash
# Fake hourly productivity histogram for today's working hours. Bars wobble
# slightly each refresh so the gradient repaints visibly.

hours=(09 10 11 12 13 14 15 16 17)
levels=(72 88 92 38 76 84 68 54 60)

for i in $(seq 0 $((${#hours[@]} - 1))); do
  base=${levels[$i]}
  jitter=$((RANDOM % 6))
  echo "${hours[$i]}h: $((base + jitter))"
done
