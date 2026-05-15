#!/usr/bin/env bash
# Attack-vector histogram input. Counts drift each refresh so the gradient
# bars repaint visibly without losing the relative ordering.

base_xss=42;    j_xss=$((RANDOM % 8))
base_sqli=28;   j_sqli=$((RANDOM % 6))
base_brute=21;  j_brute=$((RANDOM % 5))
base_phish=18;  j_phish=$((RANDOM % 4))
base_ddos=12;   j_ddos=$((RANDOM % 4))
base_mitm=7;    j_mitm=$((RANDOM % 3))
base_zero=3;    j_zero=$((RANDOM % 2))

echo "XSS:   $((base_xss + j_xss))"
echo "SQLi:  $((base_sqli + j_sqli))"
echo "Brute: $((base_brute + j_brute))"
echo "Phish: $((base_phish + j_phish))"
echo "DDoS:  $((base_ddos + j_ddos))"
echo "MITM:  $((base_mitm + j_mitm))"
echo "0day:  $((base_zero + j_zero))"
