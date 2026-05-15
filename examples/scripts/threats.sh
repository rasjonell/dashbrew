#!/usr/bin/env bash
# Rotating fake threat feed lines.

threats=(
  "BRUTE_FORCE  ssh-bastion-3       72 attempts from 185.220.x.x"
  "DDoS         edge-lb-eu          43k pps · mitigation engaged"
  "XSS          api.payments        sanitiser caught, dropped"
  "SQLi         legacy-admin        WAF rule W-117 fired"
  "MITM         vpn-gw-2            cert pin mismatch"
  "EXFIL        worker-pool-4       1.2 GB outbound, quarantined"
  "SCAN         all-ports           nmap signature 0x9c2a"
  "LOCKOUT      account svc-bot     3rd failed mfa, paged on-call"
  "REPLAY       /v1/orders          jti reused, request blocked"
  "RATE_LIMIT   /search             ip 178.62.x.x · 500/min cap"
)

# Print 8 random lines, oldest at the bottom.
for i in $(seq 1 8); do
  idx=$((RANDOM % ${#threats[@]}))
  ts=$(date "+%H:%M:%S")
  printf "[%s] %s\n" "$ts" "${threats[$idx]}"
done
