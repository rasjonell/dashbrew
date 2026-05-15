#!/usr/bin/env bash
# Error-rate histogram input: counts of errors bucketed by HTTP status code.
# Numbers wobble each refresh so the gradient bars repaint visibly.

echo "200_ok: $((9500 + RANDOM % 400))"
echo "301_redir: $((100 + RANDOM % 60))"
echo "400_bad: $((10 + RANDOM % 30))"
echo "401_auth: $((5 + RANDOM % 15))"
echo "404_miss: $((20 + RANDOM % 40))"
echo "429_rate: $((2 + RANDOM % 12))"
echo "500_err: $((1 + RANDOM % 8))"
echo "502_gw: $((0 + RANDOM % 5))"
