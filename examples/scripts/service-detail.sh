#!/usr/bin/env bash
# Renders a deeper "drill down" view for one service. Used by the "d"
# binding on each service tile in examples/devops-noc.json.

name="${1:-api}"

case "$name" in
  api)
    cat <<EOF

  service: api-gateway
  cluster: prod-us-east-1
  region:  us-east-1 · multi-az
  replicas: 6/6 healthy

  routes (last 60s):
    POST /v1/orders        184 rps   p99  87ms
    GET  /v1/users/me       92 rps   p99  41ms
    POST /v1/checkout       47 rps   p99 132ms
    GET  /health          1240 rps   p99   3ms

  recent deploys:
    v2.4.1  47m ago   author=alice    rolled out 100%
    v2.4.0  3h ago    author=bob      rolled back
    v2.3.9  yesterday author=carol    superseded

  on-call:  payments-team
  runbook:  https://wiki.internal/runbook/api
EOF
    ;;
  db)
    cat <<EOF

  service: postgres-primary
  cluster: prod-us-east-1
  region:  us-east-1a
  replicas: 1 primary · 3 replicas

  connections:    187/500
  cache hit rate: 97.3%
  replication lag: 12 ms

  slow queries (last 60s): 2
    SELECT ... FROM orders WHERE user_id = $1   1.8s
    SELECT ... FROM events WHERE ts > $1        2.4s

  last failover: 14 days ago (planned)
  next maintenance window: Sun 02:00 UTC
EOF
    ;;
  cache)
    cat <<EOF

  service: redis-cache
  cluster: prod-us-east-1
  region:  us-east-1 · 3 shards

  keys:         4.2M
  memory used:  3.1 GB / 8 GB
  hit rate:     91.4%
  evictions/s:  42

  hottest keys (last 60s):
    session:*       1.8M ops
    user:*:profile  640k ops
    cart:*           212k ops

  on-call:  platform-team
EOF
    ;;
  queue)
    cat <<EOF

  service: rabbitmq
  cluster: prod-us-east-1
  region:  us-east-1 · 3 nodes

  queue              depth   rate     consumers
  ─────────────────────────────────────────────
  orders.created      241    180/s        12
  orders.refund        18      4/s         3
  emails.outbox        37    140/s         6
  webhooks.retry      102      9/s         2

  dead-letter queue: 4 messages (last 1h)
  oldest unacked: 12s
EOF
    ;;
  *)
    echo "no detail page for service '$name'"
    ;;
esac
