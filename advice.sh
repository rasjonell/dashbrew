#!/bin/env bash
curl -s https://api.adviceslip.com/advice | jq .slip.advice
