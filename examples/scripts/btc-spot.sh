#!/usr/bin/env bash
# Emits the current simulated BTC USD price as a single float. Used by the
# trading-desk chart in refresh_mode "append" so each tick appends one point.
set -e
here="$(cd "$(dirname "$0")" && pwd)"
exec "$here/prices.sh" BTC
