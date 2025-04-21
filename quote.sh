#!/bin/env bash
curl -s https://qapi.vercel.app/api/random | jq .quote
