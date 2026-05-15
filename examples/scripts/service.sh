#!/usr/bin/env bash
# Mock service health card. Outputs status, p99 latency, success rate.
# Argument: api | db | cache | queue. Different services have different
# default latency ranges so the panel reads as plausible.

name="${1:-api}"

case "$name" in
  api)    lo=30;  hi=120 ;;
  db)     lo=5;   hi=40  ;;
  cache)  lo=1;   hi=8   ;;
  queue)  lo=2;   hi=20  ;;
  *)      lo=10;  hi=80  ;;
esac

range=$((hi - lo + 1))
lat=$((lo + RANDOM % range))
success=$((95 + RANDOM % 5))

# 1-in-9 chance of degraded status, 1-in-30 chance of down.
roll=$((RANDOM % 30))
if   [ "$roll" -eq 0 ];  then status="✗ DOWN";     success=$((50 + RANDOM % 30))
elif [ "$roll" -lt 4 ];  then status="⚠ DEGRADED"; success=$((80 + RANDOM % 15))
else                          status="✓ UP"
fi

echo
echo "  $status"
echo
printf "  p99    %3dms\n" "$lat"
printf "  succ   %d%%\n"  "$success"
