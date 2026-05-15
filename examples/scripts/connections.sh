#!/usr/bin/env bash
# Generates fake active-session lines. Each line is a hex session id, an
# ip:port pair, a country, and a status flag.

countries=(US DE BR JP SE TR IN NG CN FR)
flags=(EST EST EST EST EST FIN_W TIME_W LAST_ACK SYN_RCVD)

for i in $(seq 1 14); do
  hash=$(head -c 6 /dev/urandom | xxd -p 2>/dev/null || printf "%08x" "$RANDOM")
  ip=$(printf "%d.%d.%d.%d" $((RANDOM%255+1)) $((RANDOM%255)) $((RANDOM%255)) $((RANDOM%255)))
  port=$((1024 + RANDOM % 64000))
  c=${countries[$((RANDOM % ${#countries[@]}))]}
  f=${flags[$((RANDOM % ${#flags[@]}))]}
  printf "%-12s  %-21s  %s   %s\n" "$hash" "$ip:$port" "$c" "$f"
done
